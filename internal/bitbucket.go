package internal

import (
	"errors"
	"strings"
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
	g.httpClient = g.manager.HTTPManager().New("https://api.bitbucket.org/2.0", nil)
	sdk.LogInfo(g.logger, "starting")
	return nil
}

// Enroll is called when a new integration instance is added
func (g *BitBucketIntegration) Enroll(instance sdk.Instance) error {
	sdk.LogInfo(g.logger, "enrolling agent")
	return g.registerUnregisterWebhooks(instance, true)
}

// Dismiss is called when an existing integration instance is removed
func (g *BitBucketIntegration) Dismiss(instance sdk.Instance) error {
	sdk.LogInfo(g.logger, "dismiss not implemented")
	return g.registerUnregisterWebhooks(instance, false)
}

// Stop is called when the integration is shutting down for cleanup
func (g *BitBucketIntegration) Stop() error {
	sdk.LogInfo(g.logger, "stopping")
	return nil
}

func (g *BitBucketIntegration) getHTTPCredOpts(config sdk.Config) sdk.WithHTTPOption {
	if config.BasicAuth != nil {
		sdk.LogInfo(g.logger, "using basic auth")
		return sdk.WithBasicAuth(
			config.BasicAuth.Username,
			config.BasicAuth.Password,
		)
	}
	sdk.LogInfo(g.logger, "using oauth2")
	return sdk.WithOAuth2Refresh(
		g.manager, g.refType,
		config.OAuth2Auth.AccessToken,
		*config.OAuth2Auth.RefreshToken,
	)
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
	if config.BasicAuth == nil && config.OAuth2Auth == nil {
		return errors.New("missing authentication")
	}

	// inst := sdk.NewInstance(config, state, pipe, customerID, export.IntegrationInstanceID())
	// if err := g.Enroll(*inst); err != nil {
	// 	return err
	// }
	// os.Exit(1)

	hasInclusions := config.Inclusions != nil
	hasExclusions := config.Exclusions != nil
	accounts := config.Accounts
	if accounts == nil {
		sdk.LogInfo(g.logger, "no accounts configured, will do only customer's account")
	}
	sdk.LogInfo(g.logger, "export starting", "customer", customerID)

	client := g.httpClient
	creds := g.getHTTPCredOpts(config)
	var updated time.Time
	if !export.Historical() {
		var strTime string
		if ok, _ := state.Get("updated", &strTime); ok {
			updated, _ = time.Parse(time.RFC3339Nano, strTime)
		}
	}
	a := api.New(g.logger, client, state, customerID, g.refType, creds)
	wss, err := a.FetchWorkSpaces()
	if err != nil {
		return err
	}
	teams := api.ExtractWorkSpaceIDs(wss)
	var thirdparty []string
	if accounts != nil {
		for name := range *accounts {
			thirdparty = append(thirdparty, name)
			teams = append(teams, name)
		}
	}

	errchan := make(chan error)
	repochan := make(chan *sdk.SourceCodeRepo)
	userchan := make(chan *sdk.SourceCodeUser)
	prchan := make(chan *sdk.SourceCodePullRequest)
	prcommentchan := make(chan *sdk.SourceCodePullRequestComment)
	prcommitchan := make(chan *sdk.SourceCodePullRequestCommit)
	prreviewchan := make(chan *sdk.SourceCodePullRequestReview)
	prreviewrequestchan := make(chan *sdk.SourceCodePullRequestReviewRequest)

	// =========== repo ============
	go func() {
		var count int
		for r := range repochan {
			if hasInclusions || hasExclusions {
				name := strings.Split(r.Name, "/")
				if hasInclusions && !config.Inclusions.Matches(name[0], r.Name) {
					continue
				}
				if hasExclusions && config.Exclusions.Matches(name[0], r.Name) {
					continue
				}
			}
			team := strings.Split(r.Name, "/")[0]
			if thirdparty != nil && inslice(team, thirdparty) {
				r.Affiliation = sdk.SourceCodeRepoAffiliationThirdparty
			} else {
				r.Affiliation = sdk.SourceCodeRepoAffiliationOrganization
			}
			if err := pipe.Write(r); err != nil {
				errchan <- err
				return
			}
			if err := a.FetchPullRequests(r.Name, r.RefID, updated, prchan, prcommentchan, prcommitchan, prreviewchan, prreviewrequestchan); err != nil {
				errchan <- err
				return
			}
			count++
		}
		sdk.LogDebug(g.logger, "finished sending repos", "len", count)
	}()
	// =========== prs ============
	go func() {
		var count int
		for r := range prchan {
			if err := pipe.Write(r); err != nil {
				errchan <- err
				return
			}
			count++
		}
		sdk.LogDebug(g.logger, "finished sending prs", "len", count)
	}()
	// =========== pr comment ============
	go func() {
		var count int
		for r := range prcommentchan {
			if err := pipe.Write(r); err != nil {
				errchan <- err
				return
			}
			count++
		}
		sdk.LogDebug(g.logger, "finished sending pr comments", "len", count)
	}()
	// =========== pr commit ============
	go func() {
		var count int
		for r := range prcommitchan {
			if err := pipe.Write(r); err != nil {
				errchan <- err
				return
			}
			count++
		}
		sdk.LogDebug(g.logger, "finished sending commits", "len", count)
	}()
	// =========== pr review ============
	go func() {
		var count int
		for r := range prreviewchan {
			if err := pipe.Write(r); err != nil {
				errchan <- err
				return
			}
			count++
		}
		sdk.LogDebug(g.logger, "finished sending reviews", "len", count)
	}()
	// =========== pr review request ===========
	go func() {
		var count int
		for r := range prreviewrequestchan {
			if err := pipe.Write(r); err != nil {
				errchan <- err
				return
			}
			count++
		}
		sdk.LogDebug(g.logger, "finished sending review requests", "len", count)
	}()
	// =========== user ============
	go func() {
		var count int
		for r := range userchan {
			if err := pipe.Write(r); err != nil {
				errchan <- err
				return
			}
			count++
		}
		sdk.LogDebug(g.logger, "finished sending users", "len", count)
	}()
	go func() {
		for _, team := range teams {
			if err := a.FetchRepos(team, updated, repochan); err != nil {
				sdk.LogError(g.logger, "error fetching repos", "err", err)
				errchan <- err
				return
			}
			if err := a.FetchUsers(team, updated, userchan); err != nil {
				sdk.LogError(g.logger, "error fetching repos", "err", err)
				errchan <- err
				return
			}
		}
		errchan <- nil
	}()

	if err := <-errchan; err != nil {
		sdk.LogError(g.logger, "export finished with error", "err", err)
		return err
	}
	state.Set("updated", time.Now().Format(time.RFC3339Nano))

	close(repochan)
	close(userchan)
	close(prchan)
	close(prcommentchan)
	close(prcommitchan)
	close(prreviewchan)
	close(prreviewrequestchan)

	sdk.LogInfo(g.logger, "export finished")

	return nil
}

func inslice(word string, slice []string) bool {
	for _, w := range slice {
		if word == w {
			return true
		}
	}
	return false
}
