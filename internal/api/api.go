package api

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/pinpt/agent/v4/sdk"
)

const updatedFormat = "2006-01-02T15:04:05.999999999-07:00"

// API the api object
type API struct {
	client                sdk.HTTPClient
	state                 sdk.State
	refType               string
	customerID            string
	integrationInstanceID string
	logger                sdk.Logger
	creds                 sdk.WithHTTPOption
	pipe                  sdk.Pipe
}

// New returns a new instance of API
func New(logger sdk.Logger, client sdk.HTTPClient, state sdk.State, pipe sdk.Pipe, customerID, integrationInstanceID, refType string, creds sdk.WithHTTPOption) *API {
	return &API{
		logger:                logger,
		client:                client,
		customerID:            customerID,
		integrationInstanceID: integrationInstanceID,
		refType:               refType,
		creds:                 creds,
		state:                 state,
		pipe:                  pipe,
	}
}

func (a *API) paginate(endpoint string, params url.Values, callback func(buf json.RawMessage) error) error {
	if params == nil {
		params = url.Values{}
	}
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
		if err := callback(res.Values); err != nil {
			return err
		}
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

// getCount will return the total number of records
func (a *API) getCount(endpoint string, params url.Values) (int64, error) {
	if params == nil {
		params = url.Values{}
	}
	var res paginationResponse
	_, err := a.get(endpoint, params, &res)
	if err != nil {
		return 0, err
	}
	return res.Size, nil
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
