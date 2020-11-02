package api

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/pinpt/agent/v4/sdk"
)

// FetchPullRequests gets team members
func (a *API) FetchPullRequests(reponame string, repoid string, updated time.Time) error {
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

	out := make(chan json.RawMessage)
	errchan := make(chan error)
	var count int
	go func() {
		for obj := range out {
			if len(obj) == 0 {
				continue
			}
			rawResponse := []PullRequestResponse{}
			if err := json.Unmarshal(obj, &rawResponse); err != nil {
				errchan <- err
				return
			}
			if err := a.processPullRequests(rawResponse, reponame, repoid, updated); err != nil {
				errchan <- err
				return
			}
			count += len(rawResponse)
		}
		errchan <- nil
	}()
	if err := a.paginate(endpoint, params, out); err != nil {
		return fmt.Errorf("error fetching prs. err %v", err)
	}
	if err := <-errchan; err != nil {
		return err
	}
	sdk.LogDebug(a.logger, "finished fetching pull requests", "repo", reponame, "count", count)
	return nil
}

func (a *API) processPullRequests(raw []PullRequestResponse, reponame string, repoid string, updated time.Time) error {
	async := sdk.NewAsync(10)
	for _, _pr := range raw {
		pr := _pr
		async.Do(func() error {
			return a.fetchPullRequestComments(pr, reponame, repoid, updated)
		})
		async.Do(func() error {
			return a.fetchPullRequestCommits(pr, reponame, repoid, updated)
		})
		async.Do(func() error {
			return a.ExtractPullRequestReview(pr, repoid)
		})
	}
	if err := async.Wait(); err != nil {
		return err
	}
	// we need the first commit of every pr in the pr object, wait for the commits to be fetched before processing prs
	for _, pr := range raw {
		a.sendPullRequest(pr, repoid, updated)
	}
	return nil
}

func prReviewRequestKey(prID string) string {
	return fmt.Sprintf("review_requests:%s", prID)
}

func (a *API) syncPRReviewRequests(prID string, currentRequests map[string]bool) error {
	// check state to see if any requests have been removed
	var prevRequests []string
	key := prReviewRequestKey(prID)
	ok, err := a.state.Get(key, &prevRequests)
	if err != nil {
		return fmt.Errorf("error getting state for key %s: %w", key, err)
	}
	if ok {
		for _, prevRequestID := range prevRequests {
			if !currentRequests[prevRequestID] {
				// previous request is missing from current
				update := sdk.SourceCodePullRequestReviewRequestUpdate{}
				v := false
				update.Set.Active = &v
				if err := a.pipe.Write(sdk.NewSourceCodePullRequestReviewRequestUpdate(a.customerID, "", prevRequestID, a.refType, update)); err != nil {
					return fmt.Errorf("error writing review request update to pipe: %w", err)
				}
			}
		}
	}
	if len(currentRequests) > 0 {
		var ids []string
		for id := range currentRequests {
			ids = append(ids, id)
		}
		return a.state.Set(key, ids)
	}
	return nil
}

// ExtractPullRequestReview will pull out reviews and review requests from a pr and send them to the pipe
func (a *API) ExtractPullRequestReview(raw PullRequestResponse, repoID string) error {
	prID := sdk.NewSourceCodePullRequestID(a.customerID, strconv.FormatInt(raw.ID, 10), a.refType, repoID)
	requests := make(map[string]bool)
	for _, participant := range raw.Participants {
		if participant.Role == "REVIEWER" {
			if participant.Approved {
				if err := a.pipe.Write(&sdk.SourceCodePullRequestReview{
					Active:                true,
					IntegrationInstanceID: sdk.StringPointer(a.integrationInstanceID),
					CustomerID:            a.customerID,
					PullRequestID:         prID,
					RefID:                 sdk.Hash(raw.ID, participant.User.AccountID),
					RefType:               a.refType,
					RepoID:                repoID,
					UserRefID:             participant.User.UUID,
					State:                 sdk.SourceCodePullRequestReviewStateApproved,
				}); err != nil {
					return fmt.Errorf("error writing review to pipe: %w", err)
				}
			} else if participant.ParticipatedOn.IsZero() {
				// a non-participated reviewer is counted as a request
				id := sdk.NewSourceCodePullRequestReviewRequestID(a.customerID, a.refType, prID, participant.User.UUID)
				sdk.LogDebug(a.logger, "sending a pr review request", "_id", id)
				if err := a.pipe.Write(&sdk.SourceCodePullRequestReviewRequest{
					Active:                 true,
					CreatedDate:            sdk.SourceCodePullRequestReviewRequestCreatedDate(*sdk.NewDateWithTime(raw.UpdatedOn)),
					RequestedReviewerRefID: participant.User.UUID,
					RefType:                a.refType,
					PullRequestID:          prID,
					CustomerID:             a.customerID,
					IntegrationInstanceID:  sdk.StringPointer(a.integrationInstanceID),
					ID:                     id,
				}); err != nil {
					return fmt.Errorf("error writing review request to pipe: %w", err)
				}
				requests[id] = true
			}
		}
	}
	return a.syncPRReviewRequests(prID, requests)
}

// ConvertPullRequest converts from raw response to pinpoint object
func (a *API) ConvertPullRequest(raw PullRequestResponse, repoid, firstsha string) *sdk.SourceCodePullRequest {

	commitid := sdk.NewSourceCodeCommitID(a.customerID, firstsha, a.refType, repoid)
	pr := &sdk.SourceCodePullRequest{
		Active:                true,
		CustomerID:            a.customerID,
		IntegrationInstanceID: sdk.StringPointer(a.integrationInstanceID),
		RefType:               a.refType,
		RefID:                 fmt.Sprint(raw.ID),
		RepoID:                sdk.NewSourceCodeRepoID(a.customerID, repoid, a.refType),
		BranchID:              sdk.NewSourceCodeBranchID(a.customerID, repoid, a.refType, raw.Source.Branch.Name, commitid),
		BranchName:            raw.Source.Branch.Name,
		Title:                 raw.Title,
		Description:           `<div class="source-bitbucket">` + sdk.ConvertMarkdownToHTML(raw.Description) + "</div>",
		URL:                   raw.Links.HTML.Href,
		Identifier:            fmt.Sprintf("#%d", raw.ID), // in bitbucket looks like #1 is the format for PR identifiers in their UI
		CreatedByRefID:        raw.Author.UUID,
	}
	sdk.ConvertTimeToDateModel(raw.CreatedOn, &pr.CreatedDate)
	sdk.ConvertTimeToDateModel(raw.UpdatedOn, &pr.UpdatedDate)
	switch raw.State {
	case "OPEN":
		pr.Status = sdk.SourceCodePullRequestStatusOpen
	case "DECLINED":
		pr.Status = sdk.SourceCodePullRequestStatusClosed
		pr.ClosedByRefID = raw.ClosedBy.AccountID
		sdk.ConvertTimeToDateModel(raw.UpdatedOn, &pr.ClosedDate)
	case "MERGED":
		pr.MergeSha = raw.MergeCommit.Hash
		pr.MergeCommitID = sdk.NewSourceCodeCommitID(a.customerID, raw.MergeCommit.Hash, a.refType, pr.RepoID)
		pr.MergedByRefID = raw.ClosedBy.AccountID
		pr.Status = sdk.SourceCodePullRequestStatusMerged
		sdk.ConvertTimeToDateModel(raw.UpdatedOn, &pr.MergedDate)
	default:
		sdk.LogError(a.logger, "PR has an unknown state", "state", raw.State, "ref_id", pr.RefID)
	}
	return pr
}

func (a *API) sendPullRequest(raw PullRequestResponse, repoid string, updated time.Time) {
	if raw.UpdatedOn.Before(updated) {
		return
	}
	prid := fmt.Sprint(raw.ID)
	var firstsha string
	ok, _ := a.state.Get(FirstSha(repoid, prid), &firstsha)
	if !ok {
		sdk.LogInfo(a.logger, "no first commit sha found for pr", "pr", raw.ID, "repo", repoid)
	}
	a.pipe.Write(a.ConvertPullRequest(raw, repoid, firstsha))
}
