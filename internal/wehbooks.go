package internal

import (
	"errors"
	"time"

	"github.com/pinpt/agent.next.bitbucket/internal/api"
	"github.com/pinpt/agent.next/sdk"
)

// WebHook is called when a webhook is received on behalf of the integration
func (g *BitBucketIntegration) WebHook(webhook sdk.WebHook) error {
	sdk.LogInfo(g.logger, "webhook not implemented")
	return nil
}

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

func (g *BitBucketIntegration) registerUnregisterWebhooks(instance sdk.Instance, register bool) error {
	customerID := instance.CustomerID()
	integrationID := instance.IntegrationInstanceID()
	var creds sdk.WithHTTPOption
	config := instance.Config()
	if config.BasicAuth == nil && config.OAuth2Auth == nil {
		return errors.New("missing auth")
	}
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
	a := api.New(g.logger, g.httpClient, customerID, g.refType, creds)
	var uuid string
	var err error
	if register {
		// only needed for registering webhooks
		uuid, err = a.FetchMyUser()
		if err != nil {
			return err
		}
	}
	teams, err := a.FetchWorkSpaces()
	if err != nil {
		return err
	}
	repochan := make(chan *sdk.SourceCodeRepo)
	errchan := make(chan error)

	go func() {
		for r := range repochan {
			client := g.manager.HTTPManager().New("https://bitbucket.org/!api/2.0", nil)
			a := api.New(g.logger, client, customerID, g.refType, creds)
			if register {
				url, err := g.manager.CreateWebHook(customerID, g.refType, integrationID, r.RefID)
				if err != nil {
					errchan <- err
					return
				}
				// webhooks use a different endpoint
				if _, err = a.CreateWebHook(r.Name, r.RefID, uuid, url, webhookEvents); err != nil {
					errchan <- err
					return
				}
				sdk.LogInfo(g.logger, "webhook created", "repo name", r.Name)
				sdk.LogDebug(g.logger, "webhook created", "url", url)
			} else {
				if err := a.DeleteExistingWebHooks(r.Name); err != nil {
					errchan <- err
					return
				}
				sdk.LogInfo(g.logger, "webhook deleted", "repo name", r.Name)
			}
		}
		errchan <- nil
	}()
	go func() {
		for _, team := range teams {
			if err := a.FetchRepos(team, time.Time{}, repochan); err != nil {
				errchan <- err
			}
		}
		close(repochan)
	}()

	err = <-errchan
	return err
}
