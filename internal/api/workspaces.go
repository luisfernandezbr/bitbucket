package api

import (
	"fmt"
	"net/url"

	"github.com/pinpt/agent/v4/sdk"
)

// FetchWorkSpaces returns all workspaces
func (a *API) FetchWorkSpaces() ([]WorkSpacesResponse, error) {
	sdk.LogDebug(a.logger, "fetching workspaces")
	endpoint := "workspaces"
	params := url.Values{}
	params.Set("pagelen", "100")
	params.Set("role", "member")

	var workspaces []WorkSpacesResponse
	out := make(chan objects)
	errchan := make(chan error, 1)
	go func() {
		for obj := range out {
			res := []WorkSpacesResponse{}
			if err := obj.Unmarshal(&res); err != nil {
				errchan <- err
				return
			}
			workspaces = append(workspaces, res...)
		}
		errchan <- nil
	}()
	err := a.paginate(endpoint, params, out)
	if err != nil {
		return nil, fmt.Errorf("error fetching workspaces. err %v", err)
	}
	if err := <-errchan; err != nil {
		return nil, err
	}
	sdk.LogDebug(a.logger, "finished fetching workspaces")
	return workspaces, nil
}

// ExtractWorkSpaceIDs will return just the slugs of the give workspaces
func ExtractWorkSpaceIDs(ws []WorkSpacesResponse) []string {
	var ids []string
	for _, w := range ws {
		ids = append(ids, w.Slug)
	}
	return ids
}
