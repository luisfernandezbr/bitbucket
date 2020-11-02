package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/pinpt/agent/v4/sdk"
)

// FetchMyUser returns the logged in user
func (a *API) FetchMyUser() (MyUser, error) {
	var out MyUser
	_, err := a.get("user", nil, &out)
	return out, err
}

// FetchUsers gets team members
func (a *API) FetchUsers(team string, updated time.Time) error {
	sdk.LogDebug(a.logger, "fetching users", "team", team)
	endpoint := sdk.JoinURL("workspaces", team, "members")
	params := url.Values{}
	var count int
	out := make(chan json.RawMessage)
	errchan := make(chan error)
	go func() {
		for obj := range out {
			rawUsers := []userResponse{}
			if err := json.Unmarshal(obj, &rawUsers); err != nil {
				errchan <- err
				return
			}
			if err := a.sendUsers(rawUsers, updated); err != nil {
				errchan <- err
				return
			}
			count += len(rawUsers)
		}
		errchan <- nil
	}()
	if err := a.paginate(endpoint, params, out); err != nil {
		rerr := err.(*sdk.HTTPError)
		if rerr.StatusCode != http.StatusForbidden {
			return fmt.Errorf("error fetching users. err %v", err)
		}
	}
	if err := <-errchan; err != nil {
		return err
	}
	sdk.LogDebug(a.logger, "finished fetching users", "team", team, "count")
	return nil
}

func (a *API) sendUsers(raw []userResponse, updated time.Time) error {
	for _, meta := range raw {
		user := meta.User
		var usertype sdk.SourceCodeUserType
		if user.Type == "user" {
			usertype = sdk.SourceCodeUserTypeHuman
		} else {
			usertype = sdk.SourceCodeUserTypeBot
		}
		if err := a.pipe.Write(&sdk.SourceCodeUser{
			AvatarURL:             sdk.StringPointer(user.Links.Avatar.Href),
			CustomerID:            a.customerID,
			RefID:                 user.RefID(),
			RefType:               a.refType,
			Member:                true,
			Name:                  user.DisplayName,
			Type:                  usertype,
			URL:                   sdk.StringPointer(user.Links.HTML.Href),
			IntegrationInstanceID: sdk.StringPointer(a.integrationInstanceID),
		}); err != nil {
			return fmt.Errorf("error sending user to pipe: %w", err)
		}
	}
	return nil
}

// RefID will return the the ref_id of the user
func (u *attlassianUser) RefID() string {
	return u.AccountID
}
