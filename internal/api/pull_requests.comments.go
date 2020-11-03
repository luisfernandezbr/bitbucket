package api

import (
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/pinpt/agent/v4/sdk"
)

func (a *API) fetchPullRequestComments(pr PullRequestResponse, reponame string, repoRefID string, updated time.Time) error {
	sdk.LogDebug(a.logger, "fetching pull requests comments", "repo", reponame)
	endpoint := sdk.JoinURL("repositories", reponame, "pullrequests", fmt.Sprint(pr.ID), "comments")
	params := url.Values{}
	if !updated.IsZero() {
		params.Set("q", `updated_on > `+updated.Format(updatedFormat))
	}
	params.Set("sort", "-updated_on")

	out := make(chan json.RawMessage)
	errchan := make(chan error)
	var count int
	go func() {
		for obj := range out {
			rawResponse := []PullRequestCommentResponse{}
			if err := json.Unmarshal(obj, &rawResponse); err != nil {
				errchan <- err
				return
			}
			for _, rcomment := range rawResponse {
				if err := a.pipe.Write(ConvertPullRequestComment(rcomment, repoRefID, fmt.Sprint(pr.ID), a.customerID, a.integrationInstanceID, a.refType)); err != nil {
					errchan <- fmt.Errorf("error writing pr comment to pipe: %w", err)
					return
				}
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
	sdk.LogDebug(a.logger, "finished fetching pull request comments", "repo", reponame, "count", count)
	return nil
}

// ConvertPullRequestComment converts from raw response to pinpoint object
func ConvertPullRequestComment(raw PullRequestCommentResponse, repoRefID, prid, customerID, integrationInstanceID, refType string) *sdk.SourceCodePullRequestComment {
	item := &sdk.SourceCodePullRequestComment{
		Active:                true,
		CustomerID:            customerID,
		RefType:               refType,
		IntegrationInstanceID: sdk.StringPointer(integrationInstanceID),
		RefID:                 fmt.Sprint(raw.ID),
		URL:                   raw.Links.HTML.Href,
		RepoID:                sdk.NewSourceCodeRepoID(customerID, repoRefID, refType),
		PullRequestID:         sdk.NewSourceCodePullRequestID(customerID, prid, refType, repoRefID),
		Body:                  `<div class="source-bitbucket">` + sdk.ConvertMarkdownToHTML(raw.Content.Raw) + "</div>",
		UserRefID:             raw.User.RefID(),
	}
	sdk.ConvertTimeToDateModel(raw.UpdatedOn, &item.UpdatedDate)
	sdk.ConvertTimeToDateModel(raw.CreatedOn, &item.CreatedDate)
	return item
}
