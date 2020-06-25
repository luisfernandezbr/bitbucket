package api

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/pinpt/agent.next/sdk"
)

// FetchUsers gets team members
func (a *API) FetchUsers(team string, updated time.Time, userchan chan<- *sdk.SourceCodeUser) error {
	sdk.LogDebug(a.logger, "fetching users", "team", team)
	endpoint := sdk.JoinURL("workspaces", team, "members")
	params := url.Values{}
	out := make(chan objects)
	errchan := make(chan error)
	go func() {
		for obj := range out {
			rawUsers := []userResponse{}
			if err := obj.Unmarshal(&rawUsers); err != nil {
				errchan <- err
				return
			}
			a.sendUsers(rawUsers, updated, userchan)
		}
		errchan <- nil
	}()
	go func() {
		err := a.paginate(endpoint, params, out)
		if err != nil {
			rerr := err.(*sdk.HTTPError)
			if rerr.StatusCode == http.StatusForbidden {
				errchan <- nil
			} else {
				errchan <- fmt.Errorf("error fetching users. err %v", err)
			}
		}
	}()
	if err := <-errchan; err != nil {
		return err
	}
	sdk.LogDebug(a.logger, "finished fetching users", "team", team)
	return nil
}

func (a *API) sendUsers(raw []userResponse, updated time.Time, userchan chan<- *sdk.SourceCodeUser) {
	for _, each := range raw {
		var usertype sdk.SourceCodeUserType
		if each.Type == "user" {
			usertype = sdk.SourceCodeUserTypeHuman
		} else {
			usertype = sdk.SourceCodeUserTypeBot
		}
		userchan <- &sdk.SourceCodeUser{
			AvatarURL:  sdk.StringPointer(each.Links.Avatar.Href),
			CustomerID: a.customerID,
			RefID:      each.UUID,
			RefType:    a.refType,
			Member:     true,
			Name:       each.DisplayName,
			Type:       usertype,
			URL:        sdk.StringPointer(each.Links.HTML.Href),
		}
	}
}
