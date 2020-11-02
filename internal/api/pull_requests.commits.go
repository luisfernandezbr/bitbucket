package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/pinpt/agent/v4/sdk"
)

// FetchFirstPullRequestCommit fetches the first commit in the pr
func (a *API) FetchFirstPullRequestCommit(reponame, prid string) (string, error) {
	var hash string
	out := make(chan json.RawMessage)
	errchan := make(chan error)
	go func() {
		for obj := range out {
			rawResponse := []prCommitResponse{}
			if err := json.Unmarshal(obj, &rawResponse); err != nil {
				errchan <- err
				return
			}
			for _, each := range rawResponse {
				hash = each.Hash
			}
		}
		errchan <- nil
	}()
	endpoint := sdk.JoinURL("repositories", reponame, "pullrequests", prid, "commits")
	if err := a.paginate(endpoint, nil, out); err != nil {
		return "", err
	}
	if err := <-errchan; err != nil {
		return "", err
	}
	return hash, nil
}

func (a *API) fetchPullRequestCommits(pr PullRequestResponse, reponame string, repoid string, updated time.Time) error {
	sdk.LogDebug(a.logger, "fetching pull requests commits", "repo", reponame)
	endpoint := sdk.JoinURL("repositories", reponame, "pullrequests", fmt.Sprint(pr.ID), "commits")
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
			rawResponse := []prCommitResponse{}
			if err := json.Unmarshal(obj, &rawResponse); err != nil {
				errchan <- err
				return
			}
			a.sendPullRequestCommits(rawResponse, repoid, fmt.Sprint(pr.ID))
			count += len(rawResponse)
		}
		errchan <- nil
	}()
	if err := a.paginate(endpoint, params, out); err != nil {
		rerr := err.(*sdk.HTTPError)
		// not found means no commits
		if rerr.StatusCode == http.StatusNotFound {
			sdk.LogDebug(a.logger, "no commits found for this PR", "repo", reponame, "pr", pr.ID)
		} else {
			return fmt.Errorf("error fetching pr commits. err %v", err)
		}
	}
	if err := <-errchan; err != nil {
		return err
	}
	sdk.LogDebug(a.logger, "finished fetching pull requests commits", "repo", reponame, "count", count)
	return nil
}

func (a *API) sendPullRequestCommits(raw []prCommitResponse, repoid, prid string) error {
	if len(raw) == 0 {
		return nil
	}

	// we need the first id of the pr in the pr object
	key := FirstSha(repoid, prid)
	if !a.state.Exists(key) {
		if err := a.state.Set(key, raw[0].Hash); err != nil {
			return fmt.Errorf("error setting first commit sha: %w", err)
		}
	}
	for _, rccommit := range raw {
		item := &sdk.SourceCodePullRequestCommit{
			Active:                true,
			CustomerID:            a.customerID,
			RefType:               a.refType,
			RefID:                 rccommit.Hash,
			URL:                   rccommit.Links.HTML.Href,
			RepoID:                sdk.NewSourceCodeRepoID(a.customerID, repoid, a.refType),
			PullRequestID:         sdk.NewSourceCodePullRequestID(a.customerID, prid, a.refType, repoid),
			Sha:                   rccommit.Hash,
			Message:               rccommit.Message,
			AuthorRefID:           rccommit.Author.User.UUID,
			CommitterRefID:        rccommit.Author.User.UUID,
			IntegrationInstanceID: sdk.StringPointer(a.integrationInstanceID),
		}
		sdk.ConvertTimeToDateModel(rccommit.Date, &item.CreatedDate)
		if err := a.pipe.Write(item); err != nil {
			return fmt.Errorf("error writing pr commit to pipe: %w", err)
		}
	}
	return nil
}
