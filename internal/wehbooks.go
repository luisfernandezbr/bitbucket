package internal

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/pinpt/agent.next.bitbucket/internal/api"
	"github.com/pinpt/agent.next/sdk"
)

const webhookVersion = "12" // change this to have the webhook uninstalled and reinstalled new

var webhookEvents = []string{
	"repo:push",
	"repo:fork",
	"repo:updated",
	"repo:commit_comment_created",
	"repo:commit_status_created",
	"repo:commit_status_updated",
	"issue:created",
	"issue:updated",
	"issue:comment_created",
	"pullrequest:created",
	"pullrequest:updated",
	"pullrequest:approved",
	"pullrequest:unapproved",
	"pullrequest:fulfilled",
	"pullrequest:rejected",
	"pullrequest:comment_created",
	"pullrequest:comment_updated",
	"pullrequest:comment_deleted",
}

// WebHook is called when a webhook is received on behalf of the integration
func (g *BitBucketIntegration) WebHook(webhook sdk.WebHook) error {

	vals, err := url.ParseQuery(webhook.URL())
	if err != nil {
		fmt.Println("ERROR", err)
	}
	data, _ := webhook.Data()
	pipe := webhook.Pipe()
	customerID := webhook.CustomerID()
	config := webhook.Config()
	state := webhook.State()
	eventname := vals.Get("event")

	var creds sdk.WithHTTPOption
	if config.BasicAuth != nil {
		sdk.LogInfo(g.logger, "using basic auth")
		creds = sdk.WithBasicAuth(
			config.BasicAuth.Username,
			config.BasicAuth.Password,
		)
	} else {
		sdk.LogInfo(g.logger, "using oauth2")
		creds = sdk.WithOAuth2Refresh(
			g.manager, g.refType,
			config.OAuth2Auth.AccessToken,
			*config.OAuth2Auth.RefreshToken,
		)
	}

	a := api.New(g.logger, g.httpClient, state, customerID, g.refType, creds)

	switch eventname {
	case "repo:push", "repo:fork", "repo:commit_comment_created",
		"repo:commit_status_created", "repo:commit_status_updated",
		"issue:created", "issue:updated", "issue:comment_created":
		// do nothing, not used
	case "repo:updated":
		var raw struct {
			Repository api.RepoResponse `json:"repository"`
		}
		if err := sdk.MapToStruct(data, &raw); err != nil {
			return err
		}
		repo := a.ConvertRepo(raw.Repository)
		if err := pipe.Write(repo); err != nil {
			return err
		}

	case "pullrequest:created",
		"pullrequest:updated",
		"pullrequest:approved",
		"pullrequest:unapproved",
		"pullrequest:rejected":
		var raw struct {
			PullRequest api.PullRequestResponse `json:"pullrequest"`
			Repository  api.RepoResponse        `json:"repository"`
		}
		if err := sdk.MapToStruct(data, &raw); err != nil {
			return err
		}

		var firstsha string
		ok, _ := state.Get("prsha."+raw.Repository.UUID+"."+fmt.Sprint(raw.PullRequest.ID), &firstsha)
		if !ok {
			var err error
			firstsha, err = a.FetchFirstPullRequestCommit(raw.Repository.FullName, fmt.Sprint(raw.PullRequest.ID))
			if err != nil {
				return err
			}
		}
		pr := a.ConvertPullRequest(raw.PullRequest, raw.Repository.UUID, firstsha)
		if err := pipe.Write(pr); err != nil {
			return err
		}

	case "pullrequest:comment_created",
		"pullrequest:comment_updated",
		"pullrequest:comment_deleted":
		var raw struct {
			PullRequest api.PullRequestResponse        `json:"pullrequest"`
			Comment     api.PullRequestCommentResponse `json:"comment"`
			Repository  api.RepoResponse               `json:"repository"`
		}
		if err := sdk.MapToStruct(data, &raw); err != nil {
			return err
		}
		prcomment := api.ConvertPullRequestComment(raw.Comment, raw.Repository.UUID, fmt.Sprint(raw.PullRequest.ID), customerID, g.refType)
		if eventname == "pullrequest:comment_deleted" {
			prcomment.Active = false
		}
		if err := pipe.Write(prcomment); err != nil {
			return err
		}
	}

	return nil
}

func (g *BitBucketIntegration) registerUnregisterWebhooks(instance sdk.Instance, register bool) error {
	customerID := instance.CustomerID()
	integrationID := instance.IntegrationInstanceID()
	state := instance.State()
	config := instance.Config()
	if config.BasicAuth == nil && config.OAuth2Auth == nil {
		return errors.New("missing auth")
	}
	var creds sdk.WithHTTPOption
	if config.BasicAuth != nil {
		sdk.LogInfo(g.logger, "using basic auth")
		creds = sdk.WithBasicAuth(
			config.BasicAuth.Username,
			config.BasicAuth.Password,
		)
	} else {
		sdk.LogInfo(g.logger, "using oauth2")
		creds = sdk.WithOAuth2Refresh(
			g.manager, g.refType,
			config.OAuth2Auth.AccessToken,
			*config.OAuth2Auth.RefreshToken,
		)
	}
	var concurr int64
	var ok bool
	if ok, concurr = config.GetInt("concurrency"); !ok {
		concurr = 10
	}
	a := api.New(g.logger, g.httpClient, state, customerID, g.refType, creds)
	var userid string
	var err error
	if register {
		// only needed for registering webhooks
		userid, err = a.FetchMyUser()
		if err != nil {
			return err
		}
	}
	teams, err := a.FetchWorkSpaces()
	if err != nil {
		return err
	}
	repochan := make(chan *sdk.SourceCodeRepo, concurr)
	errchan := make(chan error)

	webhookManager := g.manager.WebHookManager()
	go func() {
		for r := range repochan {
			client := g.manager.HTTPManager().New("https://bitbucket.org/!api/2.0", nil)
			a := api.New(g.logger, client, state, customerID, g.refType, creds)
			if register {
				if err := g.registerWebhooks(r.Name, r.RefID, userid, customerID, integrationID, a, webhookManager); err != nil {
					webhookManager.Errored(customerID, integrationID, g.refType, r.RefID, sdk.WebHookScopeRepo, err)
				}
			} else {
				if err := g.unregisterWebhooks(r.Name, r.RefID, customerID, integrationID, a, webhookManager); err != nil {
					webhookManager.Errored(customerID, integrationID, g.refType, r.RefID, sdk.WebHookScopeRepo, err)
				}
			}
		}
		errchan <- nil
	}()
	for _, team := range teams {
		if err := a.FetchRepos(team, time.Time{}, repochan); err != nil {
			return err
		}
	}
	close(repochan)
	return <-errchan
}

func (g *BitBucketIntegration) registerWebhooks(reponame, repoid, userid, customerID, integrationID string, a *api.API, webhookManager sdk.WebHookManager) error {

	if webhookManager.Exists(customerID, integrationID, g.refType, repoid, sdk.WebHookScopeRepo) {
		url, err := webhookManager.HookURL(customerID, integrationID, g.refType, repoid, sdk.WebHookScopeRepo)
		if err != nil {
			return err
		}
		// check and see if we need to upgrade our webhook
		if strings.Contains(url, "&version="+webhookVersion) {
			sdk.LogInfo(g.logger, "skipping web hook install since already installed")
			return nil
		}
		if err := g.unregisterWebhooks(reponame, repoid, customerID, integrationID, a, webhookManager); err != nil {
			return err
		}
	}
	url, err := webhookManager.Create(customerID, integrationID, g.refType, repoid, sdk.WebHookScopeRepo, "version="+webhookVersion)
	if err != nil {
		return err
	}
	if err = a.CreateWebHook(reponame, repoid, userid, url, webhookEvents); err != nil {
		return err
	}
	sdk.LogInfo(g.logger, "webhook created", "repo name", reponame, "url", url)
	return nil
}

func (g *BitBucketIntegration) unregisterWebhooks(reponame, repoid, customerID, integrationID string, a *api.API, webhookManager sdk.WebHookManager) error {
	if err := webhookManager.Delete(customerID, integrationID, g.refType, repoid, sdk.WebHookScopeRepo); err != nil {
		return err
	}
	if err := a.DeleteExistingWebHooks(reponame); err != nil {
		return err
	}
	sdk.LogInfo(g.logger, "webhook deleted", "repo name", reponame)
	return nil
}
