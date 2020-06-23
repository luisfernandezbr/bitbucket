package api

import (
	"fmt"
	"net/url"
	"time"

	"github.com/pinpt/agent.next/sdk"
)

// FetchRepos gets team names
func (a *API) FetchRepos(team string, updated time.Time, repo chan<- *sdk.SourceCodeRepo) error {

	endpoint := sdk.JoinURL("repositories", team)
	params := url.Values{}
	out := make(chan objects)
	errchan := make(chan error)
	go func() {
		for obj := range out {
			rawRepos := []repoResonse{}
			if err := obj.Unmarshal(&rawRepos); err != nil {
				errchan <- err
				return
			}
			a.sendRepos(rawRepos, updated, repo)
		}
		errchan <- nil
	}()
	go func() {
		err := a.paginate(endpoint, params, out)
		if err != nil {
			fmt.Println("ERROR", err)
			errchan <- nil
		}
	}()
	if err := <-errchan; err != nil {
		return err
	}

	return nil
}

func (a *API) sendRepos(raw []repoResonse, updated time.Time, repo chan<- *sdk.SourceCodeRepo) {
	for _, each := range raw {
		if each.UpdatedOn.After(updated) {
			repo <- &sdk.SourceCodeRepo{
				Active:        true,
				CustomerID:    a.customerID,
				DefaultBranch: each.Mainbranch.Name,
				Description:   each.Description,
				Language:      each.Language,
				Name:          each.FullName,
				RefID:         each.UUID,
				RefType:       a.refType,
				URL:           each.Links.HTML.Href,
			}
		}
	}
}
