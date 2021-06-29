package main

import (
	_ "github.com/aaronland/go-cloud-s3blob"			// necessary for blob dispatcher
	_ "gocloud.dev/blob/fileblob"					// necessary for blob dispatcher
	_ "github.com/whosonfirst/go-whosonfirst-findingaid/repo"	// necessary for findingaid-repo dispatcher
	_ "github.com/whosonfirst/go-cache-blob"			// necessary for findingaid-repo dispatcher
)

import (
	_ "github.com/whosonfirst/go-webhookd-aws/v2"			// defines the lambda dispatcher
	_ "github.com/whosonfirst/go-webhookd-github"			// defines the github* transformations
	_ "github.com/whosonfirst/go-webhookd-gocloud"			// defines the blob dispatcher
	_ "github.com/whosonfirst/go-whosonfirst-webhookd/dispatcher"	// defines the findingaid-repo dispatcher
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
