package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/pinpt/agent/v4/sdk"
	"github.com/pinpt/bitbucket/internal/api"
)

const webhookVersion = "1" // change this to have the webhook uninstalled and reinstalled new

const (
	// webHookRepoPush                  api.WebHookEventName = "repo:push"
	// webHookRepoFork                  api.WebHookEventName = "repo:fork"
	webHookRepoUpdated api.WebHookEventName = "repo:updated"

	// webHookRepoCommitCommentCreated  api.WebHookEventName = "repo:commit_comment_created"
	// webHookRepoCommitStatusCreated   api.WebHookEventName = "repo:commit_status_created"
	// webHookRepoCommitStatusUpdated   api.WebHookEventName = "repo:commit_status_updated"

	// webHookIssueCreated              api.WebHookEventName = "issue:created"
	// webHookIssueUpdated              api.WebHookEventName = "issue:updated"
	// webHookIssueCommentCreated       api.WebHookEventName = "issue:comment_created"

	webHookPullrequestCreated    api.WebHookEventName = "pullrequest:created"
	webHookPullrequestUpdated    api.WebHookEventName = "pullrequest:updated"
	webHookPullrequestApproved   api.WebHookEventName = "pullrequest:approved"
	webHookPullrequestUnapproved api.WebHookEventName = "pullrequest:unapproved"
	webHookPullrequestFulfilled  api.WebHookEventName = "pullrequest:fulfilled"
	webHookPullrequestRejected   api.WebHookEventName = "pullrequest:rejected"

	webHookPullrequestCommentCreated api.WebHookEventName = "pullrequest:comment_created"
	webHookPullrequestCommentUpdated api.WebHookEventName = "pullrequest:comment_updated"
	webHookPullrequestCommentDeleted api.WebHookEventName = "pullrequest:comment_deleted"
)

var webhookEvents = []api.WebHookEventName{
	// webHookRepoPush,
	// webHookRepoFork,
	// webHookRepoCommitCommentCreated,
	// webHookRepoCommitStatusCreated,
	// webHookRepoCommitStatusUpdated,
	// webHookIssueCreated,
	// webHookIssueUpdated,
	// webHookIssueCommentCreated,
	webHookRepoUpdated,
	webHookPullrequestCreated,
	webHookPullrequestUpdated,
	webHookPullrequestApproved,
	webHookPullrequestUnapproved,
	webHookPullrequestFulfilled,
	webHookPullrequestRejected,
	webHookPullrequestCommentCreated,
	webHookPullrequestCommentUpdated,
	webHookPullrequestCommentDeleted,
}

// WebHook is called when a webhook is received on behalf of the integration
func (g *BitBucketIntegration) WebHook(webhook sdk.WebHook) error {
	logger := webhook.Logger()
	vals, err := url.ParseQuery(webhook.URL())
	if err != nil {
		return err
	}
	data := webhook.Bytes()
	pipe := webhook.Pipe()
	customerID := webhook.CustomerID()
	integrationInstanceID := webhook.IntegrationInstanceID()
	config := webhook.Config()
	state := webhook.State()
	name := vals.Get("event")

	if name == "" {
		return errors.New("missing `event` from query")
	}
	var creds sdk.WithHTTPOption
	if config.BasicAuth != nil {
		sdk.LogInfo(logger, "using basic auth")
		creds = sdk.WithBasicAuth(
			config.BasicAuth.Username,
			config.BasicAuth.Password,
		)
	} else {
		sdk.LogInfo(logger, "using oauth2")
		creds = sdk.WithOAuth2Refresh(
			g.manager, g.refType,
			config.OAuth2Auth.AccessToken,
			*config.OAuth2Auth.RefreshToken,
		)
	}

	a := api.New(logger, g.httpClient, state, pipe, customerID, integrationInstanceID, g.refType, creds)

	eventname := api.WebHookEventName(name)
	switch eventname {
	case webHookRepoUpdated:
		var raw struct {
			Repository api.RepoResponse `json:"repository"`
		}
		if err := json.Unmarshal(data, &raw); err != nil {
			return err
		}
		repo := a.ConvertRepo(raw.Repository)
		if err := pipe.Write(repo); err != nil {
			return err
		}

	case webHookPullrequestCreated, webHookPullrequestUpdated, webHookPullrequestApproved,
		webHookPullrequestUnapproved, webHookPullrequestFulfilled, webHookPullrequestRejected:
		var raw struct {
			PullRequest api.PullRequestResponse `json:"pullrequest"`
			Repository  api.RepoResponse        `json:"repository"`
		}
		if err := json.Unmarshal(data, &raw); err != nil {
			return err
		}

		shas, err := a.FetchPullRequestCommits(raw.Repository.FullName, fmt.Sprint(raw.PullRequest.ID))
		if err != nil {
			return err
		}
		pr := a.ConvertPullRequest(raw.PullRequest, raw.Repository.UUID, shas)
		if err := pipe.Write(pr); err != nil {
			return err
		}

		if err := a.ExtractPullRequestReview(raw.PullRequest, pr.RepoID); err != nil {
			return fmt.Errorf("error getting reviews: %w", err)
		}

	case webHookPullrequestCommentCreated,
		webHookPullrequestCommentUpdated,
		webHookPullrequestCommentDeleted:
		var raw struct {
			PullRequest api.PullRequestResponse        `json:"pullrequest"`
			Comment     api.PullRequestCommentResponse `json:"comment"`
			Repository  api.RepoResponse               `json:"repository"`
		}
		if err := json.Unmarshal(data, &raw); err != nil {
			return err
		}
		prcomment := api.ConvertPullRequestComment(raw.Comment, raw.Repository.UUID, fmt.Sprint(raw.PullRequest.ID), customerID, integrationInstanceID, g.refType)
		if eventname == webHookPullrequestCommentDeleted {
			prcomment.Active = false
		}
		if err := pipe.Write(prcomment); err != nil {
			return err
		}
	}

	return nil
}

func (g *BitBucketIntegration) registerUnregisterWebhooks(instance sdk.Instance, register bool) error {
	logger := instance.Logger()
	customerID := instance.CustomerID()
	integrationID := instance.IntegrationInstanceID()
	state := instance.State()
	pipe := instance.Pipe()
	config := instance.Config()
	if config.BasicAuth == nil && config.OAuth2Auth == nil {
		return errors.New("missing auth")
	}
	var creds sdk.WithHTTPOption
	if config.BasicAuth != nil {
		sdk.LogInfo(logger, "using basic auth")
		creds = sdk.WithBasicAuth(
			config.BasicAuth.Username,
			config.BasicAuth.Password,
		)
	} else {
		sdk.LogInfo(logger, "using oauth2")
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
	a := api.New(logger, g.httpClient, state, pipe, customerID, integrationID, g.refType, creds)
	var userid string
	var err error
	if register {
		// only needed for registering webhooks
		user, err := a.FetchMyUser()
		if err != nil {
			return err
		}
		userid = user.UUID
	}
	workspaces, err := a.FetchWorkSpaces()
	if err != nil {
		return err
	}
	teams := api.ExtractWorkSpaceIDs(workspaces)
	repochan := make(chan *sdk.SourceCodeRepo, concurr)
	errchan := make(chan error)

	webhookManager := g.manager.WebHookManager()
	go func() {
		for r := range repochan {
			client := g.manager.HTTPManager().New("https://bitbucket.org/!api/2.0", nil)
			a := api.New(logger, client, state, pipe, customerID, integrationID, g.refType, creds)
			if register {
				if err := g.registerWebhooks(logger, r.Name, r.RefID, userid, customerID, integrationID, a, webhookManager); err != nil {
					webhookManager.Errored(customerID, integrationID, g.refType, r.RefID, sdk.WebHookScopeRepo, err)
				}
			} else {
				if err := g.unregisterWebhooks(logger, r.Name, r.RefID, customerID, integrationID, a, webhookManager); err != nil {
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

func (g *BitBucketIntegration) registerWebhooks(logger sdk.Logger, reponame, repoid, userid, customerID, integrationID string, a *api.API, webhookManager sdk.WebHookManager) error {

	if webhookManager.Exists(customerID, integrationID, g.refType, repoid, sdk.WebHookScopeRepo) {
		url, err := webhookManager.HookURL(customerID, integrationID, g.refType, repoid, sdk.WebHookScopeRepo)
		if err != nil {
			return err
		}
		// check and see if we need to upgrade our webhook
		if strings.Contains(url, "&version="+webhookVersion) {
			sdk.LogInfo(logger, "skipping web hook install since already installed")
			return nil
		}
		if err := g.unregisterWebhooks(logger, reponame, repoid, customerID, integrationID, a, webhookManager); err != nil {
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
	sdk.LogInfo(logger, "webhook created", "repo name", reponame, "url", url)
	return nil
}

func (g *BitBucketIntegration) unregisterWebhooks(logger sdk.Logger, reponame, repoid, customerID, integrationID string, a *api.API, webhookManager sdk.WebHookManager) error {
	if err := webhookManager.Delete(customerID, integrationID, g.refType, repoid, sdk.WebHookScopeRepo); err != nil {
		return err
	}
	if err := a.DeleteExistingWebHooks(reponame); err != nil {
		return err
	}
	sdk.LogInfo(logger, "webhook deleted", "repo name", reponame)
	return nil
}
