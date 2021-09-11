// launch-esc-task is a tool that expects to be passed a single repo that will be used to construct a command to
// invoke in/on an ECS container. It was originally part of the whosonfirst/go-webhookd-aws package but clone to
// this package since it's really Who's On First specific.
package main

import (
	"context"
	"encoding/base64"
	"errors"
	"flag"
	"fmt"
	"github.com/aaronland/go-aws-ecs"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/sfomuseum/go-flags/flagset"
	"github.com/sfomuseum/go-flags/multi"
	"log"
	"os"
	"regexp"
	"strings"
)

func main() {

	fs := flagset.NewFlagSet("webhookd-aws-launch-task")

	var ecs_dsn = fs.String("ecs-dsn", "", "A valid (go-whosonfirst-aws) ECS DSN.")

	var ecs_container = fs.String("ecs-container", "", "The name of your AWS ECS container.")
	var ecs_cluster = fs.String("ecs-cluster", "", "The name of your AWS ECS cluster.")
	var ecs_task = fs.String("ecs-task", "", "The name of your AWS ECS task (inclusive of its version number),")

	var ecs_launch_type = fs.String("ecs-launch-type", "FARGATE", "...")
	var ecs_public_ip = fs.String("ecs-public-ip", "ENABLED", "...")

	var ecs_subnets multi.MultiString
	fs.Var(&ecs_subnets, "ecs-subnet", "One or more AWS subnets in which your task will run.")

	var ecs_security_groups multi.MultiString
	fs.Var(&ecs_security_groups, "ecs-security-group", "One of more AWS security groups your task will assume.")

	var mode = fs.String("mode", "cli", "...")
	var command = fs.String("command", "", "...")

	flagset.Parse(fs)

	err := flagset.SetFlagsFromEnvVarsWithFeedback(fs, "WEBHOOKD", true)

	if err != nil {
		log.Fatal(err)
	}

	if *command == "" {
		log.Fatal("Missing command")
	}

	if *mode == "lambda" {

		expand := func(candidates []string, sep string) []string {

			expanded := make([]string, 0)

			for _, c := range candidates {

				for _, v := range strings.Split(c, sep) {
					expanded = append(expanded, v)
				}
			}

			return expanded
		}

		ecs_subnets = expand(ecs_subnets, ",")
		ecs_security_groups = expand(ecs_security_groups, ",")
	}

	task_opts := &ecs.TaskOptions{
		Task:           *ecs_task,
		Container:      *ecs_container,
		Cluster:        *ecs_cluster,
		Subnets:        ecs_subnets,
		SecurityGroups: ecs_security_groups,
		LaunchType:     *ecs_launch_type,
		PublicIP:       *ecs_public_ip,
	}

	launchTask := func(command string, args ...interface{}) (interface{}, error) {

		str_cmd := fmt.Sprintf(command, args...)
		cmd := strings.Split(str_cmd, " ")

		log.Printf("Launch ECS task with command '%s'\n", str_cmd)

		task_rsp, err := ecs.LaunchTaskWithDSN(*ecs_dsn, task_opts, cmd...)

		if err != nil {
			return nil, err
		}

		return task_rsp.Tasks, nil
	}

	switch *mode {

	case "cli":

		for _, repo := range flag.Args() {

			rsp, err := launchTask(*command, repo)

			if err != nil {
				log.Fatal(err)
			}

			log.Println(rsp)
		}

	case "lambda":

		re_target, err := regexp.Compile(`^[a-zA-Z0-9\-_]+$`)

		if err != nil {
			log.Fatal(err)
		}

		// this expects to be passed a value generated by
		// func (d *LambdaDispatcher) Dispatch in dispatcher.go
		// which is why the base64 decoding (20200528/thisisaaronland)

		// this is a nuisance when we are trying to test things so we
		// will first check the string to see if it looks like a base64
		// encoded value before we try to decode it (20200903/thisisaaronland)
		//
		// https://stackoverflow.com/questions/475074/regex-to-parse-or-validate-base64-data

		re_b64, err := regexp.Compile(`^(?:[A-Za-z0-9+/]{4})*(?:[A-Za-z0-9+/]{2}==|[A-Za-z0-9+/]{3}=)?$`)

		if err != nil {
			log.Fatal(err)
		}

		string_handler := func(ctx context.Context, payload string) (interface{}, error) {

			target := payload

			// see notes above

			if re_b64.MatchString(payload) {

				target_b, err := base64.StdEncoding.DecodeString(payload)

				if err != nil {
					log.Printf("Payload (%s) is not a base64 encoded string", payload)
					return nil, err
				}

				target = string(target_b)
			}

			if !re_target.MatchString(target) {
				return nil, errors.New("Invalid payload")
			}

			return launchTask(*command, target)
		}

		lambda.Start(string_handler)

	default:
		log.Fatal("Unknown mode")
	}

	os.Exit(0)
}
