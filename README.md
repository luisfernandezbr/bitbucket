## BitBucket Pinpoint Integration

### Overview

To run it locally, cd to `$GOPATH/src/github.com/pinpt/agent.next` and run the following

    go run -tags dev . dev ../agent.next.bitbucket \
	--set 'basic_auth={"username":USER_NAME,"password":PASSWORD}' \
	--set 'accounts={"bitbucket":{"login":"bitbucket", "type":"ORG", "public":true}, "microsoft":{"login":"microsoft", "type":"ORG", "public":true}}' \
	 --set 'exclusions={"bitbucket": "bitbucket/geordi"}'


The `--set basic_auth` is required, but the others are not.
The `--set accounts` is used to add public and open source repos
The `--set exclusions` is used to list a repos not to be exported

### Author

- Pinpoint
- https://pinpoint.com
- 2020-06-17 21:00:28.301105 +0200 CEST m=+51.002814245
