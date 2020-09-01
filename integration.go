package main

import (
	"github.com/pinpt/agent.next.bitbucket/internal"
	"github.com/pinpt/agent.next/runner"
)

// Integration is used to export the integration
var Integration internal.BitBucketIntegration

func main() {
	runner.Main(&Integration)
}
