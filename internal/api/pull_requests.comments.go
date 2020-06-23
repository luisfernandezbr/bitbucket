package api

import (
	"fmt"
	"net/url"
	"strconv"

	"github.com/pinpt/agent.next/sdk"
)

func (a *API) fetchPullRequestComments(pr prResponse, reponame string, repoid string, prcommentchan chan<- *sdk.SourceCodePullRequestComment) error {
	sdk.LogInfo(a.logger, "processing pr comments", "repo", reponame)
	endpoint := sdk.JoinURL("repositories", reponame, "pullrequests", fmt.Sprint(pr.ID), "comments")
	params := url.Values{}

	out := make(chan objects)
	errchan := make(chan error)
	go func() {
		for obj := range out {
			rawResponse := []prCommentResponse{}
			if err := obj.Unmarshal(&rawResponse); err != nil {
				errchan <- err
				return
			}
			a.sendPullRequestComments(rawResponse, repoid, fmt.Sprint(pr.ID), prcommentchan)
		}
		errchan <- nil
	}()
	go func() {
		err := a.paginate(endpoint, params, out)
		if err != nil {
			fmt.Println("ERROR", err)
			errchan <- nil
		}
	}()
	if err := <-errchan; err != nil {
		return err
	}
	return nil
}

func (a *API) sendPullRequestComments(raw []prCommentResponse, repoid, prid string, prcommentchan chan<- *sdk.SourceCodePullRequestComment) {
	for _, rcomment := range raw {
		item := &sdk.SourceCodePullRequestComment{
			CustomerID:    a.customerID,
			RefType:       a.refType,
			RefID:         strconv.FormatInt(rcomment.ID, 10),
			URL:           rcomment.Links.HTML.Href,
			RepoID:        sdk.NewSourceCodeRepoID(a.customerID, repoid, a.refType),
			PullRequestID: sdk.NewSourceCodePullRequestID(a.customerID, prid, a.refType, repoid),
			Body:          rcomment.Content.Raw,
			UserRefID:     rcomment.User.UUID,
		}
		sdk.ConvertTimeToDateModel(rcomment.UpdatedOn, &item.UpdatedDate)
		sdk.ConvertTimeToDateModel(rcomment.CreatedOn, &item.CreatedDate)
		prcommentchan <- item
	}
}
