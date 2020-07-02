package api

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/pinpt/agent.next/sdk"
)

const updatedFormat = "2006-01-02T15:04:05.999999999-07:00"

// API the api object
type API struct {
	client     sdk.HTTPClient
	refType    string
	customerID string
	logger     sdk.Logger
	creds      sdk.WithHTTPOption
}

func New(logger sdk.Logger, client sdk.HTTPClient, customerID, refType string, creds sdk.WithHTTPOption) *API {
	return &API{
		logger:     logger,
		client:     client,
		customerID: customerID,
		refType:    refType,
		creds:      creds,
	}
}

func (a *API) paginate(endpoint string, params url.Values, out chan<- objects) error {
	defer close(out)
	var page string
	for {
		var res paginationResponse
		if page != "" {
			params.Set("page", page)
		}
		_, err := a.get(endpoint, params, &res)
		if err != nil {
			return err
		}
		out <- res.Values
		if res.Next == "" {
			return nil
		}
		u, _ := url.Parse(res.Next)
		page = u.Query().Get("page")
		if page == "" {
			return fmt.Errorf("no `page` in next. %v", u.String())
		}
	}
}

func (a *API) get(endpoint string, params url.Values, out interface{}) (*sdk.HTTPResponse, error) {
	if params == nil {
		params = url.Values{}
	}
	return a.client.Get(out, sdk.WithEndpoint(endpoint), sdk.WithGetQueryParameters(params), a.creds)
}

func (a *API) delete(endpoint string, out interface{}) (*sdk.HTTPResponse, error) {
	return a.client.Delete(out, sdk.WithEndpoint(endpoint), a.creds)
}

func (a *API) post(endpoint string, data interface{}, params url.Values, out interface{}) (*sdk.HTTPResponse, error) {
	if params == nil {
		params = url.Values{}
	}
	return a.client.Post(strings.NewReader(sdk.Stringify(data)), out, sdk.WithEndpoint(endpoint), sdk.WithGetQueryParameters(params), a.creds)
}

type objects []map[string]interface{}

func (o objects) Unmarshal(out interface{}) error {
	b, err := json.Marshal(o)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, out)
}
