package api

import (
	"encoding/json"
	"net/url"

	"github.com/pinpt/agent/v4/sdk"
)

const webhookName = "pinpoint_webhooks"

// WebHookEventName the event name string type
type WebHookEventName string

// CreateWebHook creates a webhook, deleting existing ones if exist
func (a *API) CreateWebHook(reponame, repoid, userid, ur string, hooks []WebHookEventName) error {
	endpoint := sdk.JoinURL("repositories", reponame, "hooks")
	async := sdk.NewAsync(10)
	for _, _h := range hooks {
		h := string(_h)
		u := ur
		async.Do(func() error {
			params := url.Values{}
			params.Set("event", h)
			u += "&" + params.Encode()
			payload := webhookPayload{
				Active:      true,
				CreatorID:   "user:" + userid,
				Description: webhookName,
				Events:      []string{h},
				SubjectKey:  "repository:" + repoid,
				URL:         u,
			}
			var out struct {
				UUID string `json:"uuid"`
			}
			_, err := a.post(endpoint, payload, nil, &out)
			return err
		})
	}
	return async.Wait()
}

// DeleteWebhook deletes a webhook
func (a *API) DeleteWebhook(reponame, uuid string) error {
	endpoint := sdk.JoinURL("repositories", reponame, "hooks", uuid)
	var out interface{}
	_, err := a.delete(endpoint, &out)
	return err
}

// DeleteExistingWebHooks deletes all the pinpoint webhooks
func (a *API) DeleteExistingWebHooks(reponame string) error {
	endpoint := sdk.JoinURL("repositories", reponame, "hooks")
	return a.paginate(endpoint, nil, func(obj json.RawMessage) error {
		var resp []struct {
			Description string `json:"description"`
			UUID        string `json:"uuid"`
		}
		if err := json.Unmarshal(obj, &resp); err != nil {
			return err
		}
		for _, wh := range resp {
			if wh.Description == webhookName {
				if err := a.DeleteWebhook(reponame, wh.UUID); err != nil {
					return err
				}
			}
		}
		return nil
	})
}
