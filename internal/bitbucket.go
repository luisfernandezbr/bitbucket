package internal

import (
	"errors"
	"sync"
	"time"

	"github.com/pinpt/agent.next.bitbucket/internal/api"
	"github.com/pinpt/agent.next/sdk"
)

// BitBucketIntegration is an integration for BitBucket
type BitBucketIntegration struct {
	logger  sdk.Logger
	config  sdk.Config
	manager sdk.Manager
	refType string

	httpClient sdk.HTTPClient
}

var _ sdk.Integration = (*BitBucketIntegration)(nil)

// Start is called when the integration is starting up
func (g *BitBucketIntegration) Start(logger sdk.Logger, config sdk.Config, manager sdk.Manager) error {
	g.logger = sdk.LogWith(logger, "pkg", "bitbucket")
	g.config = config
	g.manager = manager
	g.refType = "bitbucket"
	sdk.LogInfo(g.logger, "starting")
	return nil
}

// Enroll is called when a new integration instance is added
func (g *BitBucketIntegration) Enroll(instance sdk.Instance) error {
	sdk.LogInfo(g.logger, "enroll not implemented")
	return nil
}

// Dismiss is called when an existing integration instance is removed
func (g *BitBucketIntegration) Dismiss(instance sdk.Instance) error {
	sdk.LogInfo(g.logger, "dismiss not implemented")
	return nil
}

// WebHook is called when a webhook is received on behalf of the integration
func (g *BitBucketIntegration) WebHook(webhook sdk.WebHook) error {
	sdk.LogInfo(g.logger, "webhook not implemented")
	return nil
}

// Stop is called when the integration is shutting down for cleanup
func (g *BitBucketIntegration) Stop() error {
	sdk.LogInfo(g.logger, "stopping")
	return nil
}

// Export is called to tell the integration to run an export
func (g *BitBucketIntegration) Export(export sdk.Export) error {
	sdk.LogInfo(g.logger, "export started")

	// Pipe must be called to begin an export and receive a pipe for sending data
	pipe := export.Pipe()

	// State is a customer specific state object for this integration and customer
	state := export.State()

	// CustomerID will return the customer id for the export
	customerID := export.CustomerID()

	// Config is any customer specific configuration for this customer
	config := export.Config()
	ok, username := config.GetString("username")
	if !ok {
		return errors.New("missing username")
	}
	ok, password := config.GetString("password")
	if !ok {
		return errors.New("missing password")
	}

	g.httpClient = g.manager.HTTPManager().New("https://api.bitbucket.org/2.0", nil)

	sdk.LogDebug(g.logger, "export starting")

	client := g.httpClient
	creds := &api.BasicCreds{
		Username: username,
		Password: password,
	}

	var updated time.Time
	var strTime string
	if ok, _ := state.Get("updated", &strTime); ok {
		updated, _ = time.Parse(time.RFC3339Nano, strTime)
	}

	a := api.New(g.logger, client, creds, customerID, g.refType)
	teams, err := a.FetchTeams()
	if err != nil {
		return err
	}

	repochan := make(chan *sdk.SourceCodeRepo)
	userchan := make(chan *sdk.SourceCodeUser)
	prchan := make(chan *sdk.SourceCodePullRequest)
	prcommentchan := make(chan *sdk.SourceCodePullRequestComment)
	prcommitchan := make(chan *sdk.SourceCodePullRequestCommit)
	prreviewchan := make(chan *sdk.SourceCodePullRequestReview)

	wg := sync.WaitGroup{}
	// =========== user ============
	wg.Add(1)
	go func() {
		defer wg.Done()
		var count int
		for r := range userchan {
			pipe.Write(r)
			count++
		}
		sdk.LogInfo(g.logger, "finished sending users", "len", count)
	}()
	// =========== repo ============
	wg.Add(1)
	go func() {
		defer wg.Done()
		var count int
		for r := range repochan {
			pipe.Write(r)
			a.FetchPullRequests(r.Name, r.RefID, updated,
				prchan,
				prcommentchan,
				prcommitchan,
				prreviewchan,
			)
			count++
		}
		sdk.LogInfo(g.logger, "finished sending repos", "len", count)
	}()
	// =========== prs ============
	wg.Add(1)
	go func() {
		defer wg.Done()
		var count int
		for r := range prchan {
			pipe.Write(r)
			count++
		}
		sdk.LogInfo(g.logger, "finished sending prs", "len", count)
	}()
	// =========== pr comment ============
	wg.Add(1)
	go func() {
		defer wg.Done()
		var count int
		for r := range prcommentchan {
			pipe.Write(r)
			count++
		}
		sdk.LogInfo(g.logger, "finished sending pr comments", "len", count)
	}()
	// =========== pr commit ============
	wg.Add(1)
	go func() {
		defer wg.Done()
		var count int
		for r := range prcommitchan {
			pipe.Write(r)
			count++
		}
		sdk.LogInfo(g.logger, "finished sending commits", "len", count)
	}()
	// =========== pr review ============
	wg.Add(1)
	go func() {
		defer wg.Done()
		var count int
		for r := range prreviewchan {
			pipe.Write(r)
			count++
		}
		sdk.LogInfo(g.logger, "finished sending reviews", "len", count)
	}()

	for _, team := range teams {

		if err := a.FetchRepos(team, updated, repochan); err != nil {
			sdk.LogError(g.logger, "error fetching repos", "err", err)
			return err
		}
		if err := a.FetchUsers(team, updated, userchan); err != nil {
			sdk.LogError(g.logger, "error fetching repos", "err", err)
			return err
		}

	}
	close(repochan)
	close(userchan)
	close(prchan)
	close(prcommentchan)
	close(prcommitchan)
	close(prreviewchan)
	wg.Wait()
	state.Set("updated", time.Now().Format(time.RFC3339Nano))

	return nil
}
