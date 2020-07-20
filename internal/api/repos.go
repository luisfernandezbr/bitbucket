package api

import (
	"fmt"
	"net/url"
	"time"

	"github.com/pinpt/agent.next/sdk"
)

// FetchRepos gets team names
func (a *API) FetchRepos(team string, updated time.Time, repo chan<- *sdk.SourceCodeRepo) error {
	sdk.LogDebug(a.logger, "fetching repos", "team", team)
	endpoint := sdk.JoinURL("repositories", team)
	params := url.Values{}
	out := make(chan objects)
	errchan := make(chan error)
	go func() {
		for obj := range out {
			rawRepos := []RepoResponse{}
			if err := obj.Unmarshal(&rawRepos); err != nil {
				errchan <- err
				return
			}
			a.sendRepos(rawRepos, updated, repo)
		}
		errchan <- nil
	}()
	if err := a.paginate(endpoint, params, out); err != nil {
		return fmt.Errorf("error fetching repos. err %v", err)
	}
	if err := <-errchan; err != nil {
		return err
	}
	sdk.LogDebug(a.logger, "finished fetching repos", "team", team)
	return nil
}

// ConvertRepo converts from raw response to pinpoint object
func (a *API) ConvertRepo(raw RepoResponse) *sdk.SourceCodeRepo {
	return &sdk.SourceCodeRepo{
		Active:        true,
		CustomerID:    a.customerID,
		DefaultBranch: raw.Mainbranch.Name,
		Description:   raw.Description,
		Language:      raw.Language,
		Name:          raw.FullName,
		RefID:         raw.UUID,
		RefType:       a.refType,
		URL:           raw.Links.HTML.Href,
	}
}

func (a *API) sendRepos(raw []RepoResponse, updated time.Time, repo chan<- *sdk.SourceCodeRepo) {
	for _, each := range raw {
		repo <- a.ConvertRepo(each)
	}
}
