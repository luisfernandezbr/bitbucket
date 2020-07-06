package internal

import "github.com/pinpt/agent.next/sdk"

// Mutation is called when a mutation is received on behalf of the integration
func (g *BitBucketIntegration) Mutation(mutation sdk.Mutation) error {
	sdk.LogInfo(g.logger, "mutation not implemented")
	return nil
}
