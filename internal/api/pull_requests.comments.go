package api

import (
	"fmt"
	"net/url"
	"time"

	"github.com/pinpt/agent.next/sdk"
)

func (a *API) fetchPullRequestComments(pr PullRequestResponse, reponame string, repoid string, updated time.Time, prcommentchan chan<- *sdk.SourceCodePullRequestComment) error {
	sdk.LogDebug(a.logger, "fetching pull requests comments", "repo", reponame)
	endpoint := sdk.JoinURL("repositories", reponame, "pullrequests", fmt.Sprint(pr.ID), "comments")
	params := url.Values{}
	if !updated.IsZero() {
		params.Set("q", `updated_on > `+updated.Format(updatedFormat))
	}
	params.Set("sort", "-updated_on")

	out := make(chan objects)
	errchan := make(chan error)
	var count int
	go func() {
		for obj := range out {
			rawResponse := []PullRequestCommentResponse{}
			if err := obj.Unmarshal(&rawResponse); err != nil {
				errchan <- err
				return
			}
			for _, rcomment := range rawResponse {
				prcommentchan <- ConvertPullRequestComment(rcomment, repoid, fmt.Sprint(pr.ID), a.customerID, a.refType)
			}
			count += len(rawResponse)
		}
		errchan <- nil
	}()
	err := a.paginate(endpoint, params, out)
	if err != nil {
		return fmt.Errorf("error getting pr comments. err %v", err)
	}
	if err := <-errchan; err != nil {
		return err
	}
	sdk.LogDebug(a.logger, "finished fetching pull requests comments", "repo", reponame, "count", count)
	return nil
}

// ConvertPullRequestComment converts from raw response to pinpoint object
func ConvertPullRequestComment(raw PullRequestCommentResponse, repoid, prid, customerID, refType string) *sdk.SourceCodePullRequestComment {
	item := &sdk.SourceCodePullRequestComment{
		Active:        true,
		CustomerID:    customerID,
		RefType:       refType,
		RefID:         fmt.Sprint(raw.ID),
		URL:           raw.Links.HTML.Href,
		RepoID:        sdk.NewSourceCodeRepoID(customerID, repoid, refType),
		PullRequestID: sdk.NewSourceCodePullRequestID(customerID, prid, refType, repoid),
		Body:          `<div class="source-bitbucket">` + sdk.ConvertMarkdownToHTML(raw.Content.Raw) + "</div>",
		UserRefID:     raw.User.UUID,
	}
	sdk.ConvertTimeToDateModel(raw.UpdatedOn, &item.UpdatedDate)
	sdk.ConvertTimeToDateModel(raw.CreatedOn, &item.CreatedDate)
	return item
}
