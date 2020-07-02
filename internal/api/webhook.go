package api

import (
	"github.com/pinpt/agent.next/sdk"
)

const webhookName = "pinpoint_webhooks"

// CreateWebhook creates a webhook, deleting existing ones if exist
func (a *API) CreateWebhook(reponame, repoid, userid, url string, hooks []string) (string, error) {
	if err := a.DeleteExistingWebhooks(reponame); err != nil {
		return "", err
	}
	endpoint := sdk.JoinURL("repositories", reponame, "hooks")
	payload := webhookPayload{
		Active:      true,
		CreatorID:   "user:" + userid,
		Description: webhookName,
		Events:      hooks,
		SubjectKey:  "repository:" + repoid,
		URL:         url,
	}
	var out struct {
		UUID string `json:"uuid"`
	}
	_, err := a.post(endpoint, payload, nil, &out)
	if err != nil {
		return "", err
	}
	return out.UUID, nil
}

// DeleteWebhook deletes a webhook
func (a *API) DeleteWebhook(reponame, uuid string) error {
	endpoint := sdk.JoinURL("repositories", reponame, "hooks", uuid)
	var out interface{}
	_, err := a.delete(endpoint, &out)
	return err
}

// DeleteExistingWebhooks deletes all the pinpoint webhooks
func (a *API) DeleteExistingWebhooks(reponame string) error {
	endpoint := sdk.JoinURL("repositories", reponame, "hooks")
	out := make(chan objects)
	errchan := make(chan error)
	go func() {
		for obj := range out {
			var resp []struct {
				Description string `json:"description"`
				UUID        string `json:"uuid"`
			}
			if err := obj.Unmarshal(&resp); err != nil {
				errchan <- err
				return
			}
			for _, wh := range resp {
				if wh.Description == webhookName {
					if err := a.DeleteWebhook(reponame, wh.UUID); err != nil {
						errchan <- err
						return
					}
				}
			}
		}
		errchan <- nil
	}()
	go func() {
		if err := a.paginate(endpoint, nil, out); err != nil {
			errchan <- err
		}
	}()
	if err := <-errchan; err != nil {
		return err
	}
	return nil
}