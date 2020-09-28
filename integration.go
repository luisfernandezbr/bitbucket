package main

import (
	"github.com/pinpt/bitbucket/internal"
	"github.com/pinpt/agent/runner"
)

// Integration is used to export the integration
var Integration internal.BitBucketIntegration

func main() {
	runner.Main(&Integration)
}
