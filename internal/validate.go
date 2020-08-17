package internal

import "github.com/pinpt/agent.next/sdk"

// Validate is called when the integration is requesting a validation from the app
func (g *BitBucketIntegration) Validate(validate sdk.Validate) (map[string]interface{}, error) {
	return nil, nil
}
