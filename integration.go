package main

import (
	"github.com/pinpt/agent/v4/runner"
	"github.com/pinpt/bitbucket/internal"
)

// Integration is used to export the integration
var Integration internal.BitBucketIntegration

func main() {
	runner.Main(&Integration)
}
