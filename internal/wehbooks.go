package internal

import (
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

func (g *BitBucketIntegration) registerWebhooks(customerID, integrationID string, creds sdk.WithHTTPOption) error {

	a := api.New(g.logger, g.httpClient, customerID, g.refType, creds)
	uuid, err := a.FetchMyUser()
	if err != nil {
		return err
	}
	teams, err := a.FetchWorkSpaces()
	if err != nil {
		return err
	}
	repochan := make(chan *sdk.SourceCodeRepo)
	errchan := make(chan error)

	go func() {
		for r := range repochan {
			url, err := g.manager.CreateWebHook(customerID, g.refType, integrationID, r.RefID)
			if err != nil {
				errchan <- err
				return
			}
			// webhooks use a different endpoint
			client := g.manager.HTTPManager().New("https://bitbucket.org/!api/2.0", nil)
			a := api.New(g.logger, client, customerID, g.refType, creds)
			if uuid, err = a.CreateWebHook(r.Name, r.RefID, uuid, url, webhookEvents); err != nil {
				errchan <- err
				return
			}
			sdk.LogInfo(g.logger, "webhook created", "id", uuid)
			sdk.LogDebug(g.logger, "webhook created", "url", url)
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
