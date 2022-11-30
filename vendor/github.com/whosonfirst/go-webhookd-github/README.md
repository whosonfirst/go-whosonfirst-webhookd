# go-webhookd-github

Go package to implement the `whosonfirst/go-webhookd` interfaces for receiving and transforming webhooks originating from GitHub.

## Documentation

[![Go Reference](https://pkg.go.dev/badge/github.com/whosonfirst/go-webhookd-github.svg)](https://pkg.go.dev/github.com/whosonfirst/go-webhookd-github)

### go-webhookd

Before you begin please [read the go-webhookd documentation](https://github.com/whosonfirst/go-webhookd/blob/master/README.md) for an overview of concepts and principles.

## Usage

```
import (
	_ "github.com/go-webhookd-github"
)
```

## Receivers

### GitHub

The `GitHub` receiver handles Webhooks sent from [GitHub](https://developer.github.com/webhooks/). It validates that the message sent is actually from GitHub (by way of the `X-Hub-Signature` header) but performs no other processing. It is defined as a URI string in the form of:

```
github://?secret={SECRET}&ref={REF}
```

#### Properties

| Name | Value | Description | Required |
| --- | --- | --- | --- |
| secret | string | The secret used to generate [the HMAC hex digest](https://developer.github.com/webhooks/#delivery-headers) of the message payload. | yes |
| ref | string | An optional Git `ref` to filter by. If present and a WebHook is sent with a different ref then the daemon will return a `666` error response. | no |

## Transformations

### GitHubCommits

The `GitHubCommits` transformation will extract all the commits (added, modified, removed) from a `push` event and return a CSV encoded list of rows consisting of: commit hash, repository name, path. For example:

```
e3a18d4de60a5e50ca78ca1733238735ddfaef4c,sfomuseum-data-flights-2020-05,data/171/316/450/9/1713164509.geojson
e3a18d4de60a5e50ca78ca1733238735ddfaef4c,sfomuseum-data-flights-2020-05,data/171/316/451/9/1713164519.geojson
e3a18d4de60a5e50ca78ca1733238735ddfaef4c,sfomuseum-data-flights-2020-05,data/171/316/483/5/1713164835.geojson
````

It is defined as a URI string in the form of:

```
githubcommits://?exclude_additions={EXCLUDE_ADDITIONS}&exclude_modification={EXCLUDE_MODIFICATIONS}&exclude_deletions={EXCLUDE_DELETIONS}
```

#### Properties

| Name | Value | Description | Required |
| --- | --- | --- | --- |
| exclude_additions| boolean | A flag to indicate that new additions in a commit should be ignored. | no |
| exclude_modifications| boolean | A flag to indicate that modifications in a commit should be ignored. | no |
| exclude_deletions | boolean | A flag to indicate that deletions in a commit should be ignored. | no |
| prepend_message | boolean | An optional boolean value to prepend the commit message to the final output. This takes the form of '#message,{COMMIT_MESSAGE},' | no |
| prepend_author | boolean | An optional boolean value to prepend the name of the commit author to the final output. This takes the form of '#author,{COMMIT_AUTHOR},' | no |
| halt_on_message | string | An optional regular expression that will be compared to the commit message; if it matches the transformer will return an error with code `webhookd.HaltEvent` | no | 
| halt_on_author | string | An optional regular expression that will be compared to the commit author; if it matches the transformer will return an error with code `webhookd.HaltEvent` | no |

### GitHubRepo

The `GitHubRepo` transformation will extract the reporsitory name for all the commits matching (added, modified, removed) criteria. It is defined as a URI string in the form of:

```
githubrepo://?exclude_additions={EXCLUDE_ADDITIONS}&exclude_modification={EXCLUDE_MODIFICATIONS}&exclude_deletions={EXCLUDE_DELETIONS}
```

#### Properties

| Name | Value | Description | Required |
| --- | --- | --- | --- |
| exclude_additions| boolean | A flag to indicate that new additions in a commit should be ignored. | no |
| exclude_modifications| boolean | A flag to indicate that modifications in a commit should be ignored. | no |
| exclude_deletions | boolean | A flag to indicate that deletions in a commit should be ignored. | no |
| prepend_message | boolean | An optional boolean value to prepend the commit message to the final output. This takes the form of '#message,{COMMIT_MESSAGE},' | no |
| prepend_author | boolean | An optional boolean value to prepend the name of the commit author to the final output. This takes the form of '#author,{COMMIT_AUTHOR},' | no |
| halt_on_message | string | An optional regular expression that will be compared to the commit message; if it matches the transformer will return an error with code `webhookd.HaltEvent` | no | 
| halt_on_author | string | An optional regular expression that will be compared to the commit author; if it matches the transformer will return an error with code `webhookd.HaltEvent` | no |

## See also

* https://github.com/whosonfirst/go-webhookd