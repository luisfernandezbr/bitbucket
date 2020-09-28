package internal

import (
	"errors"
	"fmt"

	"github.com/pinpt/bitbucket/internal/api"
	"github.com/pinpt/agent/sdk"
)

// toAccount converts a WorkSpacesResponse to a sdk ConfigAccount
func toAccount(ws api.WorkSpacesResponse, accountType sdk.ConfigAccountType, repoCount int64) *sdk.ConfigAccount {
	return &sdk.ConfigAccount{
		ID:          ws.UUID,
		Type:        accountType,
		Public:      !ws.IsPrivate,
		Name:        &ws.Name,
		Description: &ws.Slug,
		TotalCount:  &repoCount,
		Selected:    sdk.BoolPointer(true),
	}
}

func toConfigAccounts(accounts []*sdk.ConfigAccount) *sdk.ConfigAccounts {
	res := make(sdk.ConfigAccounts)
	for _, account := range accounts {
		res[account.ID] = account
	}
	return &res
}

// isUserWorkspace will determine if a workspace is the user's default one
func isUserWorkspace(ws api.WorkSpacesResponse, user api.MyUser) bool {
	return ws.Name == user.DisplayName
}

func separateUserWorkSpace(user api.MyUser, workspaces []api.WorkSpacesResponse) (userWorkspace api.WorkSpacesResponse, otherWorkspaces []api.WorkSpacesResponse) {
	var foundUser bool
	for _, ws := range workspaces {
		if isUserWorkspace(ws, user) && !foundUser {
			userWorkspace = ws
			foundUser = true
		} else {
			otherWorkspaces = append(otherWorkspaces, ws)
		}
	}
	return
}

// AutoConfigure is called during the onboard process and it will add accounts to the config using the provided auth
func (g *BitBucketIntegration) AutoConfigure(autoconfig sdk.AutoConfigure) (*sdk.Config, error) {
	config := autoconfig.Config()
	if config.Scope == nil {
		return nil, errors.New("no config scope given for autoconfig")
	}
	sdk.LogInfo(g.logger, "autoconfiguring", "scope", config.Scope, "customer_id", autoconfig.CustomerID())
	a := api.New(g.logger, g.httpClient, autoconfig.State(), autoconfig.Pipe(), autoconfig.CustomerID(), autoconfig.IntegrationInstanceID(), g.refType, g.getHTTPCredOpts(config))
	workspaces, err := a.FetchWorkSpaces()
	if err != nil {
		return nil, fmt.Errorf("error fetching user workspaces: %w", err)
	}
	var accounts []*sdk.ConfigAccount
	// if theres just one then it's the user's workspace
	if len(workspaces) == 1 {
		switch *config.Scope {
		case sdk.OrgScope:
			sdk.LogWarn(g.logger, "org scope autoconfig only included one workspace", "workspace_name", workspaces[0].Name)
		case sdk.UserScope:
			repoCount, err := a.FetchRepoCount(workspaces[0].Slug)
			if err != nil {
				return nil, fmt.Errorf("error getting repo count: %w", err)
			}
			accounts = append(accounts, toAccount(workspaces[0], sdk.ConfigAccountTypeUser, repoCount))
		default:
			sdk.LogWarn(g.logger, "unexpected auto config scope", "scope", config.Scope)
		}
	} else {
		// need to sort out user workspace from others
		user, err := a.FetchMyUser()
		if err != nil {
			return nil, fmt.Errorf("error fetching user: %w", err)
		}
		userWS, otherWS := separateUserWorkSpace(user, workspaces)
		switch *config.Scope {
		case sdk.OrgScope:
			for _, ws := range otherWS {
				repoCount, err := a.FetchRepoCount(ws.Slug)
				if err != nil {
					return nil, fmt.Errorf("error getting repo count: %w", err)
				}
				accounts = append(accounts, toAccount(ws, sdk.ConfigAccountTypeOrg, repoCount))
			}
		case sdk.UserScope:
			repoCount, err := a.FetchRepoCount(userWS.Slug)
			if err != nil {
				return nil, fmt.Errorf("error getting repo count: %w", err)
			}
			accounts = append(accounts, toAccount(userWS, sdk.ConfigAccountTypeUser, repoCount))
		default:
			sdk.LogWarn(g.logger, "unexpected auto config scope", "scope", config.Scope)
		}
	}
	config.Accounts = toConfigAccounts(accounts)
	return &config, nil
}
