package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/pinpt/agent/v4/sdk"
)

// FetchPullRequestCommits fetches the first commit in the pr
func (a *API) FetchPullRequestCommits(reponame, prid string) ([]string, error) {
	var shas []string
	endpoint := sdk.JoinURL("repositories", reponame, "pullrequests", prid, "commits")
	if err := a.paginate(endpoint, nil, func(obj json.RawMessage) error {
		rawResponse := []prCommitResponse{}
		if err := json.Unmarshal(obj, &rawResponse); err != nil {
			return err
		}
		for _, each := range rawResponse {
			shas = append(shas, each.Hash)
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return shas, nil
}

// fetchPullRequestCommits will fetch all the commits for the pr after updated
func (a *API) fetchPullRequestCommits(pr PullRequestResponse, reponame string, repoRefID string, updated time.Time) ([]string, error) {
	sdk.LogDebug(a.logger, "fetching pull requests commits", "repo", reponame)
	endpoint := sdk.JoinURL("repositories", reponame, "pullrequests", fmt.Sprint(pr.ID), "commits")
	params := url.Values{}
	if !updated.IsZero() {
		params.Set("q", `updated_on > `+updated.Format(updatedFormat))
	}
	params.Set("sort", "-updated_on")
	var count int
	var shas []string
	err := a.paginate(endpoint, params, func(obj json.RawMessage) error {
		rawResponse := []prCommitResponse{}
		if err := json.Unmarshal(obj, &rawResponse); err != nil {
			return err
		}
		commitShas, err := a.sendPullRequestCommits(rawResponse, repoRefID, fmt.Sprint(pr.ID))
		if err != nil {
			return fmt.Errorf("error sending pr commits: %w", err)
		}
		shas = append(shas, commitShas...)
		count += len(rawResponse)
		return nil
	})
	if err != nil {
		rerr := err.(*sdk.HTTPError)
		// not found means no commits
		if rerr.StatusCode == http.StatusNotFound {
			sdk.LogDebug(a.logger, "no commits found for this PR", "repo", reponame, "pr", pr.ID)
		} else {
			return nil, fmt.Errorf("error fetching pr commits. err %v", err)
		}
	}
	sdk.LogDebug(a.logger, "finished fetching pull request commits", "repo", reponame, "count", count)
	return shas, nil
}

func (a *API) sendPullRequestCommits(raw []prCommitResponse, repoRefID, prRefID string) ([]string, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	// size the array to save a little mem/cpu
	shas := make([]string, len(raw))
	for i, rccommit := range raw {
		shas[i] = rccommit.Hash
		item := &sdk.SourceCodePullRequestCommit{
			Active:                true,
			CustomerID:            a.customerID,
			RefType:               a.refType,
			RefID:                 rccommit.Hash,
			URL:                   rccommit.Links.HTML.Href,
			RepoID:                sdk.NewSourceCodeRepoID(a.customerID, repoRefID, a.refType),
			PullRequestID:         sdk.NewSourceCodePullRequestID(a.customerID, prRefID, a.refType, repoRefID),
			Sha:                   rccommit.Hash,
			Message:               rccommit.Message,
			AuthorRefID:           rccommit.Author.User.RefID(),
			CommitterRefID:        rccommit.Author.User.RefID(),
			IntegrationInstanceID: sdk.StringPointer(a.integrationInstanceID),
		}
		sdk.ConvertTimeToDateModel(rccommit.Date, &item.CreatedDate)
		if err := a.pipe.Write(item); err != nil {
			return nil, fmt.Errorf("error writing pr commit to pipe: %w", err)
		}
	}
	return shas, nil
}
