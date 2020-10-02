package main

import (
	"github.com/pinpt/bitbucket/internal"
	"github.com/pinpt/agent/v4/runner"
)

// Integration is used to export the integration
var Integration internal.BitBucketIntegration

func main() {
	runner.Main(&Integration)
}
