package api

import (
	"fmt"
	"io/ioutil"
	"net/url"

	"github.com/pinpt/agent.next/sdk"
)

func (a *API) fetchPullRequestCommits(pr prResponse, reponame string, repoid string, prcommitchan chan<- *sdk.SourceCodePullRequestCommit) error {
	sdk.LogInfo(a.logger, "processing pr comments", "repo", reponame)
	endpoint := sdk.JoinURL("repositories", reponame, "pullrequests", fmt.Sprint(pr.ID), "commits")
	params := url.Values{}
	// params.Set("fields", "values.hash,values.message,values.date,values.author.raw")

	out := make(chan objects)
	errchan := make(chan error)
	go func() {
		for obj := range out {
			rawResponse := []prCommitResponse{}
			if err := obj.Unmarshal(&rawResponse); err != nil {
				errchan <- err
				return
			}
			a.sendPullRequestCommits(rawResponse, repoid, fmt.Sprint(pr.ID), prcommitchan)
		}
		errchan <- nil
	}()
	go func() {
		err := a.paginate(endpoint, params, out)
		if err != nil {
			rerr := err.(*sdk.HTTPError)
			b, _ := ioutil.ReadAll(rerr.Body)
			fmt.Println("ERROR", string(b))
			errchan <- nil
		}
	}()
	if err := <-errchan; err != nil {
		return err
	}
	return nil
}

func (a *API) sendPullRequestCommits(raw []prCommitResponse, repoid, prid string, prcommitchan chan<- *sdk.SourceCodePullRequestCommit) {
	for _, rccommit := range raw {
		item := &sdk.SourceCodePullRequestCommit{
			CustomerID:     a.customerID,
			RefType:        a.refType,
			RefID:          rccommit.Hash,
			URL:            rccommit.Links.HTML.Href,
			RepoID:         sdk.NewSourceCodeRepoID(a.customerID, repoid, a.refType),
			PullRequestID:  sdk.NewSourceCodePullRequestID(a.customerID, prid, a.refType, repoid),
			Sha:            rccommit.Hash,
			Message:        rccommit.Message,
			AuthorRefID:    rccommit.Author.User.UUID,
			CommitterRefID: rccommit.Author.User.UUID,
		}
		sdk.ConvertTimeToDateModel(rccommit.Date, &item.CreatedDate)
		prcommitchan <- item
	}
}
