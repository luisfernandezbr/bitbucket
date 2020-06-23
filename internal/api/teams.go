package api

import (
	"fmt"
	"net/url"
)

// FetchTeams gets team names
func (a *API) FetchTeams() ([]string, error) {

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
				fmt.Println("ERROR", err)
				errchan <- err
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
			fmt.Println("ERROR", err)
			errchan <- nil
		}
	}()
	if err := <-errchan; err != nil {
		return nil, err
	}

	return names, nil
}
