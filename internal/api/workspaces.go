package api

import (
	"fmt"
	"net/url"
)

// FetchWorkSpaces returns all workspaces
func (a *API) FetchWorkSpaces() ([]string, error) {
	endpoint := "workspaces"
	params := url.Values{}
	params.Set("pagelen", "100")
	params.Set("role", "member")

	var ids []string
	out := make(chan objects)
	errchan := make(chan error, 1)
	go func() {
		for obj := range out {
			res := []workSpacesResponse{}
			if err := obj.Unmarshal(&res); err != nil {
				fmt.Println("ERROR", err)
				errchan <- err
			}
			ids = append(ids, workSpaceIDs(res)...)
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
		return nil, err
	}
	return ids, nil
}

func workSpaceIDs(ws []workSpacesResponse) []string {
	var ids []string
	for _, w := range ws {
		ids = append(ids, w.UUID)
	}
	return ids
}
