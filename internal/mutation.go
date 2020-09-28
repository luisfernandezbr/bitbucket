package internal

import "github.com/pinpt/agent/v4/sdk"

// Mutation is called when a mutation is received on behalf of the integration
func (g *BitBucketIntegration) Mutation(mutation sdk.Mutation) error {
	sdk.LogInfo(g.logger, "mutation not implemented")
	return nil
}
