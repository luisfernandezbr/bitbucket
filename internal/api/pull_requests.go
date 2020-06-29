package api

import (
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/pinpt/agent.next/sdk"
)

// FetchPullRequests gets team members
func (a *API) FetchPullRequests(reponame string, repoid string, updated time.Time,
	prchan chan<- *sdk.SourceCodePullRequest,
	prcommentchan chan<- *sdk.SourceCodePullRequestComment,
	prcommitchan chan<- *sdk.SourceCodePullRequestCommit,
	prreviewchan chan<- *sdk.SourceCodePullRequestReview,
) error {
	sdk.LogDebug(a.logger, "fetching pull requests", "repo", reponame)
	endpoint := sdk.JoinURL("repositories", reponame, "pullrequests")
	params := url.Values{}
	params.Add("state", "MERGED")
	params.Add("state", "SUPERSEDED")
	params.Add("state", "OPEN")
	if !updated.IsZero() {
		params.Set("q", `updated_on > `+updated.Format(updatedFormat))
	}
	params.Set("sort", "-updated_on")

	// Greater than 50 throws "Invalid pagelen"
	params.Set("pagelen", "50")

	out := make(chan objects)
	errchan := make(chan error)
	var count int
	go func() {
		for obj := range out {
			if len(obj) == 0 {
				continue
			}
			rawResponse := []prResponse{}
			if err := obj.Unmarshal(&rawResponse); err != nil {
				errchan <- err
				return
			}
			if err := a.processPullRequests(rawResponse, reponame, repoid, updated,
				prchan,
				prcommentchan,
				prcommitchan,
				prreviewchan,
			); err != nil {
				errchan <- err
				return
			}
			count += len(rawResponse)
		}
		errchan <- nil
	}()
	go func() {
		if err := a.paginate(endpoint, params, out); err != nil {
			errchan <- fmt.Errorf("error fetching prs. err %v", err)
		}
	}()
	if err := <-errchan; err != nil {
		return err
	}
	sdk.LogDebug(a.logger, "finished fetching pull requests", "repo", reponame, "count", count)
	return nil
}

func (a *API) processPullRequests(raw []prResponse, reponame string, repoid string, updated time.Time,
	prchan chan<- *sdk.SourceCodePullRequest,
	prcommentchan chan<- *sdk.SourceCodePullRequestComment,
	prcommitchan chan<- *sdk.SourceCodePullRequestCommit,
	prreviewchan chan<- *sdk.SourceCodePullRequestReview,
) error {
	async := sdk.NewAsync(10)
	for _, _pr := range raw {
		pr := _pr
		async.Do(func() error {
			return a.fetchPullRequestComments(pr, reponame, repoid, updated, prcommentchan)
		})
		async.Do(func() error {
			return a.fetchPullRequestCommits(pr, reponame, repoid, updated, prcommitchan)
		})
		a.sendPullRequestReview(pr, repoid, prreviewchan)
		a.sendPullRequest(pr, repoid, updated, prchan)
	}
	return async.Wait()
}

func (a *API) sendPullRequestReview(raw prResponse, repoid string, prreviewchan chan<- *sdk.SourceCodePullRequestReview) {
	for _, participant := range raw.Participants {
		if participant.Role == "REVIEWER" {
			review := &sdk.SourceCodePullRequestReview{
				CustomerID:    a.customerID,
				PullRequestID: strconv.FormatInt(raw.ID, 10),
				RefID:         sdk.Hash(raw.ID, participant.User.AccountID),
				RefType:       a.refType,
				RepoID:        sdk.NewSourceCodeRepoID(a.customerID, repoid, a.refType),
				UserRefID:     participant.User.UUID,
			}
			if participant.Approved {
				review.State = sdk.SourceCodePullRequestReviewStateApproved
			} else {
				review.State = sdk.SourceCodePullRequestReviewStatePending
			}
			prreviewchan <- review
		}
	}
}
func (a *API) sendPullRequest(raw prResponse, repoid string, updated time.Time, prchan chan<- *sdk.SourceCodePullRequest) {
	if raw.UpdatedOn.Before(updated) {
		return
	}
	pr := &sdk.SourceCodePullRequest{
		CustomerID:     a.customerID,
		RefType:        a.refType,
		RefID:          fmt.Sprint(raw.ID),
		RepoID:         sdk.NewSourceCodeRepoID(a.customerID, repoid, a.refType),
		BranchName:     raw.Source.Branch.Name,
		Title:          raw.Title,
		Description:    raw.Description,
		URL:            raw.Links.HTML.Href,
		Identifier:     fmt.Sprintf("#%d", raw.ID), // in bitbucket looks like #1 is the format for PR identifiers in their UI
		CreatedByRefID: raw.Author.UUID,
	}
	sdk.ConvertTimeToDateModel(raw.CreatedOn, &pr.CreatedDate)
	sdk.ConvertTimeToDateModel(raw.UpdatedOn, &pr.MergedDate)
	sdk.ConvertTimeToDateModel(raw.UpdatedOn, &pr.ClosedDate)
	sdk.ConvertTimeToDateModel(raw.UpdatedOn, &pr.UpdatedDate)
	switch raw.State {
	case "OPEN":
		pr.Status = sdk.SourceCodePullRequestStatusOpen
	case "DECLINED":
		pr.Status = sdk.SourceCodePullRequestStatusClosed
		pr.ClosedByRefID = raw.ClosedBy.AccountID
	case "MERGED":
		pr.MergeSha = raw.MergeCommit.Hash
		pr.MergeCommitID = sdk.NewSourceCodeCommitID(a.customerID, raw.MergeCommit.Hash, a.refType, pr.RepoID)
		pr.MergedByRefID = raw.ClosedBy.AccountID
		pr.Status = sdk.SourceCodePullRequestStatusMerged
	default:
		sdk.LogError(a.logger, "PR has an unknown state", "state", raw.State, "ref_id", pr.RefID)
	}
	prchan <- pr
}
