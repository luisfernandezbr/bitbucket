package internal

import "github.com/pinpt/agent/v4/sdk"

// Mutation is called when a mutation is received on behalf of the integration
func (g *BitBucketIntegration) Mutation(mutation sdk.Mutation) (*sdk.MutationResponse, error) {
	sdk.LogInfo(mutation.Logger(), "mutation not implemented")
	return nil, nil
}
