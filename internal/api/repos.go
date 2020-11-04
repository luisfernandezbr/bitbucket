package api

import (
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/pinpt/agent/v4/sdk"
)

// FetchRepos gets team names
func (a *API) FetchRepos(team string, updated time.Time, repo chan<- *sdk.SourceCodeRepo) error {
	sdk.LogDebug(a.logger, "fetching repos", "team", team, "since", updated)
	endpoint := sdk.JoinURL("repositories", team)
	params := url.Values{}
	if !updated.IsZero() {
		params.Set("q", `updated_on > `+updated.Format(updatedFormat))
	}
	params.Set("sort", "-updated_on")
	var count int
	if err := a.paginate(endpoint, params, func(obj json.RawMessage) error {
		rawRepos := []RepoResponse{}
		if err := json.Unmarshal(obj, &rawRepos); err != nil {
			return err
		}
		count += len(rawRepos)
		for _, each := range rawRepos {
			ts := time.Now()
			repo <- a.ConvertRepo(each)
			sdk.LogDebug(a.logger, "processed repo", "updated_on", each.UpdatedOn, "since", updated, "waited", time.Since(ts))
		}
		return nil
	}); err != nil {
		return fmt.Errorf("error fetching repos. err %v", err)
	}
	sdk.LogDebug(a.logger, "finished fetching repos", "team", team, "count", count)
	return nil
}

// FetchRepoCount will return the number of repos for a workspace
func (a *API) FetchRepoCount(workspaceSlug string) (int64, error) {
	endpoint := sdk.JoinURL("repositories", workspaceSlug)
	return a.getCount(endpoint, nil)
}

// ConvertRepo converts from raw response to pinpoint object
func (a *API) ConvertRepo(raw RepoResponse) *sdk.SourceCodeRepo {
	var visibility sdk.SourceCodeRepoVisibility
	if raw.IsPrivate {
		visibility = sdk.SourceCodeRepoVisibilityPublic
	} else {
		visibility = sdk.SourceCodeRepoVisibilityPrivate
	}
	// .Affiliation is set in the main bitbucket.go file
	return &sdk.SourceCodeRepo{
		Active:                true,
		CustomerID:            a.customerID,
		DefaultBranch:         raw.Mainbranch.Name,
		Description:           raw.Description,
		Language:              raw.Language,
		Name:                  raw.FullName,
		RefID:                 raw.UUID,
		RefType:               a.refType,
		URL:                   raw.Links.HTML.Href,
		Visibility:            visibility,
		IntegrationInstanceID: sdk.StringPointer(a.integrationInstanceID),
	}
}
