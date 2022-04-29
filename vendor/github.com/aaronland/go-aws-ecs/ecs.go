package ecs

import (
	"errors"
	"github.com/aaronland/go-aws-session"
	"github.com/aws/aws-sdk-go/aws"
	aws_session "github.com/aws/aws-sdk-go/aws/session"
	aws_cloudwatchlogs "github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	aws_ecs "github.com/aws/aws-sdk-go/service/ecs"
	_ "log"
)

type TaskResponse struct {
	Tasks      []string
	TaskOutput *aws_ecs.RunTaskOutput
}

type TaskOptions struct {
	Task            string
	Container       string
	Cluster         string
	LaunchType      string
	PlatformVersion string
	PublicIP        string
	Subnets         []string
	SecurityGroups  []string
}

type MonitorTaskResultSet map[string]*MonitorTaskResult

type MonitorTaskResult struct {
	ARN    string
	Errors []error
	Logs   []*aws_cloudwatchlogs.OutputLogEvent
}

type MonitorTaskOptions struct {
	DSN       string
	Container string
	Cluster   string
	WithLogs  bool
	LogsDSN   string
}

func LaunchTaskWithDSN(dsn string, task_opts *TaskOptions, cmd ...string) (*TaskResponse, error) {

	sess, err := session.NewSessionWithDSN(dsn)

	if err != nil {
		return nil, err
	}

	return LaunchTaskWithSession(sess, task_opts, cmd...)
}

func LaunchTaskWithSession(sess *aws_session.Session, task_opts *TaskOptions, cmd ...string) (*TaskResponse, error) {

	ecs_svc := aws_ecs.New(sess)

	cluster := aws.String(task_opts.Cluster)
	task := aws.String(task_opts.Task)

	launch_type := aws.String(task_opts.LaunchType)
	platform_version := aws.String(task_opts.PlatformVersion)
	public_ip := aws.String(task_opts.PublicIP)

	subnets := make([]*string, len(task_opts.Subnets))
	security_groups := make([]*string, len(task_opts.SecurityGroups))

	for i, sn := range task_opts.Subnets {
		subnets[i] = aws.String(sn)
	}

	for i, sg := range task_opts.SecurityGroups {
		security_groups[i] = aws.String(sg)
	}

	aws_cmd := make([]*string, len(cmd))

	for i, str := range cmd {
		aws_cmd[i] = aws.String(str)
	}

	network := &aws_ecs.NetworkConfiguration{
		AwsvpcConfiguration: &aws_ecs.AwsVpcConfiguration{
			AssignPublicIp: public_ip,
			SecurityGroups: security_groups,
			Subnets:        subnets,
		},
	}

	process_override := &aws_ecs.ContainerOverride{
		Name:    aws.String(task_opts.Container),
		Command: aws_cmd,
	}

	overrides := &aws_ecs.TaskOverride{
		ContainerOverrides: []*aws_ecs.ContainerOverride{
			process_override,
		},
	}

	input := &aws_ecs.RunTaskInput{
		Cluster:              cluster,
		TaskDefinition:       task,
		LaunchType:           launch_type,
		PlatformVersion:      platform_version,
		NetworkConfiguration: network,
		Overrides:            overrides,
	}

	task_output, err := ecs_svc.RunTask(input)

	if err != nil {
		return nil, err
	}

	if len(task_output.Tasks) == 0 {
		return nil, errors.New("run task returned no errors... but no tasks")
	}

	task_arns := make([]string, len(task_output.Tasks))

	for i, t := range task_output.Tasks {
		task_arns[i] = *t.TaskArn
	}

	task_rsp := &TaskResponse{
		Tasks:      task_arns,
		TaskOutput: task_output,
	}

	return task_rsp, nil
}
