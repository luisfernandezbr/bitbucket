package api

import (
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/pinpt/agent.next/sdk"
)

func (a *API) fetchPullRequestComments(pr prResponse, reponame string, repoid string, updated time.Time, prcommentchan chan<- *sdk.SourceCodePullRequestComment) error {
	sdk.LogDebug(a.logger, "fetching pull requests comments", "repo", reponame)
	endpoint := sdk.JoinURL("repositories", reponame, "pullrequests", fmt.Sprint(pr.ID), "comments")
	params := url.Values{}
	params.Set("q", `updated_on > `+updated.Format(updatedFormat))
	params.Set("sort", "-updated_on")

	out := make(chan objects)
	errchan := make(chan error)
	var count int
	go func() {
		for obj := range out {
			rawResponse := []prCommentResponse{}
			if err := obj.Unmarshal(&rawResponse); err != nil {
				errchan <- err
				return
			}
			a.sendPullRequestComments(rawResponse, repoid, fmt.Sprint(pr.ID), prcommentchan)
			count += len(rawResponse)
		}
		errchan <- nil
	}()
	go func() {
		err := a.paginate(endpoint, params, out)
		if err != nil {
			errchan <- fmt.Errorf("error getting pr comments. err %v", err)
		}
	}()
	if err := <-errchan; err != nil {
		return err
	}
	sdk.LogDebug(a.logger, "finished fetching pull requests comments", "repo", reponame, "count", count)
	return nil
}

func (a *API) sendPullRequestComments(raw []prCommentResponse, repoid, prid string, prcommentchan chan<- *sdk.SourceCodePullRequestComment) {
	for _, rcomment := range raw {
		item := &sdk.SourceCodePullRequestComment{
			CustomerID:    a.customerID,
			RefType:       a.refType,
			RefID:         strconv.FormatInt(rcomment.ID, 10),
			URL:           rcomment.Links.HTML.Href,
			RepoID:        sdk.NewSourceCodeRepoID(a.customerID, repoid, a.refType),
			PullRequestID: sdk.NewSourceCodePullRequestID(a.customerID, prid, a.refType, repoid),
			Body:          rcomment.Content.Raw,
			UserRefID:     rcomment.User.UUID,
		}
		sdk.ConvertTimeToDateModel(rcomment.UpdatedOn, &item.UpdatedDate)
		sdk.ConvertTimeToDateModel(rcomment.CreatedOn, &item.CreatedDate)
		prcommentchan <- item
	}
}
