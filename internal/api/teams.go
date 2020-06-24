package api

import (
	"fmt"
	"net/url"

	"github.com/pinpt/agent.next/sdk"
)

// FetchTeams gets team names
func (a *API) FetchTeams() ([]string, error) {
	sdk.LogDebug(a.logger, "fetching teams")
	endpoint := "teams"
	params := url.Values{}
	params.Set("pagelen", "100")
	params.Set("role", "member")
	var names []string

	out := make(chan objects)
	errchan := make(chan error, 1)
	go func() {
		for obj := range out {
			var res []struct {
				Name string `json:"username"`
			}
			if err := obj.Unmarshal(&res); err != nil {
				errchan <- err
				return
			}
			for _, n := range res {
				names = append(names, n.Name)
			}
		}
		errchan <- nil
	}()
	go func() {
		err := a.paginate(endpoint, params, out)
		if err != nil {
			errchan <- fmt.Errorf("error fetching teams. err %v", err)
		}
	}()
	if err := <-errchan; err != nil {
		return nil, err
	}
	sdk.LogDebug(a.logger, "finished fetching teams")
	return names, nil
}
