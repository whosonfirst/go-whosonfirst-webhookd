// webhookd is an instance of the default `whosonfirst/go-webhookd` daemon with a variety of Who's On First
// specific packages enabled.
package main

import (
	// necessary for blob dispatcher and the findingaid-repo dispatcher (by way of the go-whosonfirst-findingaid cache)
	_ "github.com/aaronland/gocloud-blob-s3"
	// necessary for findingaid-repo dispatcher
	_ "github.com/whosonfirst/go-cache-blob"
	// necessary for blob dispatcher
	_ "gocloud.dev/blob/fileblob"
)

import (
	// defines the lambda dispatcher
	_ "github.com/whosonfirst/go-webhookd-aws/v2"
	// defines the github* transformations
	_ "github.com/whosonfirst/go-webhookd-github"
	// defines the blob dispatcher
	_ "github.com/whosonfirst/go-webhookd-gocloud"
)

import (
	_ "gocloud.dev/runtimevar/awsparamstore"
)

import (
	"context"
	"github.com/sfomuseum/go-flags/flagset"
	"github.com/whosonfirst/go-webhookd/v3/config"
	"github.com/whosonfirst/go-webhookd/v3/daemon"
	"log"
	"os"
)

func main() {

	fs := flagset.NewFlagSet("webhooks")

	config_uri := fs.String("config-uri", "", "A valid Go Cloud runtimevar URI representing your webhookd config.")

	flagset.Parse(fs)

	err := flagset.SetFlagsFromEnvVarsWithFeedback(fs, "WEBHOOKD", true)

	if err != nil {
		log.Fatalf("Failed to set flags from env vars, %v", err)
	}

	ctx := context.Background()

	cfg, err := config.NewConfigFromURI(ctx, *config_uri)

	if err != nil {
		log.Fatalf("Failed to load config %s, %v", *config_uri, err)
	}

	wh_daemon, err := daemon.NewWebhookDaemonFromConfig(ctx, cfg)

	if err != nil {
		log.Fatal(err)
	}

	err = wh_daemon.Start(ctx)

	if err != nil {
		log.Fatal(err)
	}

	os.Exit(0)
}
