// webhookd is an instance of the default `whosonfirst/go-webhookd` daemon with a variety of Who's On First
// specific packages enabled.
package main

import (
	// necessary for blob dispatcher
	_ "github.com/aaronland/gocloud-blob-s3"
	// necessary for blob dispatcher
	_ "gocloud.dev/blob/fileblob"
	// necessary for pubsub dispatcher
	_ "gocloud.dev/pubsub/awssnssqs"
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
	_ "gocloud.dev/runtimevar/constantvar"
	_ "gocloud.dev/runtimevar/filevar"
)

import (
	"context"
	"log"
	"os"
	"strings"

	aa_log "github.com/aaronland/go-log/v2"
	"github.com/sfomuseum/go-flags/flagset"
	"github.com/sfomuseum/runtimevar"
	"github.com/whosonfirst/go-webhookd/v3/config"
	"github.com/whosonfirst/go-webhookd/v3/daemon"
	
)

func main() {

	fs := flagset.NewFlagSet("webhooks")

	config_uri := fs.String("config-uri", "", "A valid Go Cloud runtimevar URI representing your webhookd config.")

	flagset.Parse(fs)

	ctx := context.Background()
	logger := log.Default()
	
	err := flagset.SetFlagsFromEnvVarsWithFeedback(fs, "WEBHOOKD", true)

	if err != nil {
		aa_log.Fatal(logger, "Failed to set flags from env vars, %v", err)
	}
	
	str_cfg, err := runtimevar.StringVar(ctx, *config_uri)

	if err != nil {
		aa_log.Fatal(logger, "Failed to open runtimevar, %v", err)
	}

	cfg_r := strings.NewReader(str_cfg)

	cfg, err := config.NewConfigFromReader(ctx, cfg_r)

	if err != nil {
		aa_log.Fatal(logger, "Failed to load config from reader, %v", err)
	}

	wh_daemon, err := daemon.NewWebhookDaemonFromConfig(ctx, cfg)

	if err != nil {
		aa_log.Fatal(logger, "Failed to create new webhookd, %v", err)
	}

	err = wh_daemon.StartWithLogger(ctx, logger)

	if err != nil {
		aa_log.Fatal(logger, "Failed to serve requests, %v", err)
	}

	os.Exit(0)
}
