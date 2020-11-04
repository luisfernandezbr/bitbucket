package api

import (
	"encoding/json"
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
	err := a.paginate(endpoint, params, func(obj json.RawMessage) error {
		res := []WorkSpacesResponse{}
		if err := json.Unmarshal(obj, &res); err != nil {
			return err
		}
		workspaces = append(workspaces, res...)

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("error fetching workspaces. err %v", err)
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
