package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
)

// AnsiblePlaybookConfig holds configuration for running an Ansible playbook
type AnsiblePlaybookConfig struct {
	InstanceIDs      []string
	PlaybookContent  string
	VarsFileContent  string
	ExtraVars        map[string]interface{}
	Timeout          int
	Interval         int
	SensitiveOutput  bool
}

// AnsiblePlaybookResult holds the result of playbook execution
type AnsiblePlaybookResult struct {
	CommandID string
	Status    string
	Outputs   map[string]string
}

// Logger interface for logging operations
type Logger interface {
	Info(ctx context.Context, msg string, fields map[string]interface{})
	Error(ctx context.Context, msg string, fields map[string]interface{})
}

// RunAnsiblePlaybook executes an Ansible playbook on instances via SSM
func RunAnsiblePlaybook(ctx context.Context, client *ssm.SSM, config AnsiblePlaybookConfig, logger Logger) (*AnsiblePlaybookResult, error) {
	// Convert instance IDs to AWS string pointers
	var instanceIDs []*string
	for _, id := range config.InstanceIDs {
		instanceIDs = append(instanceIDs, aws.String(id))
	}

	// Wait for instances to be ready in SSM first
	if logger != nil {
		logger.Info(ctx, "Waiting for instances to be ready in SSM", map[string]interface{}{
			"instance_count": len(instanceIDs),
		})
	}
	if err := waitForSSMReady(ctx, client, instanceIDs, config.Timeout, config.Interval); err != nil {
		return nil, err
	}

	// Build ansible-playbook command
	commands := []string{
		"#!/bin/bash",
		"set -e",
		"cd /tmp",
		fmt.Sprintf("cat > playbook.yml << 'PLAYBOOK_EOF'\n%s\nPLAYBOOK_EOF", config.PlaybookContent),
	}

	// Add vars file if provided
	varsFileProvided := config.VarsFileContent != ""
	if varsFileProvided {
		commands = append(commands, fmt.Sprintf("cat > vars.yml << 'VARS_EOF'\n%s\nVARS_EOF", config.VarsFileContent))
	}

	// Build ansible-playbook command with appropriate flags
	ansibleCmd := "ansible-playbook playbook.yml"

	if varsFileProvided {
		ansibleCmd += " -e @vars.yml"
	}

	// Add extra vars if provided
	if len(config.ExtraVars) > 0 {
		varsJSON, err := json.Marshal(config.ExtraVars)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal extra_vars: %w", err)
		}
		ansibleCmd += fmt.Sprintf(" --extra-vars '%s'", string(varsJSON))
	}

	commands = append(commands, ansibleCmd)

	if logger != nil {
		logger.Info(ctx, "Running Ansible playbook", map[string]interface{}{
			"instance_count": len(instanceIDs),
		})
	}

	// Send command via SSM
	var fullCommand string
	for _, cmd := range commands {
		fullCommand += cmd + "\n"
	}

	sendInput := &ssm.SendCommandInput{
		InstanceIds:  instanceIDs,
		DocumentName: aws.String("AWS-RunShellScript"),
		Parameters: map[string][]*string{
			"commands": {aws.String(fullCommand)},
		},
		TimeoutSeconds: aws.Int64(int64(config.Timeout)),
	}

	sendOutput, err := client.SendCommand(sendInput)
	if err != nil {
		return nil, fmt.Errorf("failed to send SSM command: %w", err)
	}

	commandID := *sendOutput.Command.CommandId
	if logger != nil {
		logger.Info(ctx, "SSM command sent", map[string]interface{}{
			"command_id": commandID,
		})
	}

	// Wait for command completion
	deadline := time.Now().Add(time.Duration(config.Timeout) * time.Second)
	outputs := make(map[string]string)

	for {
		if time.Now().After(deadline) {
			return nil, fmt.Errorf("timeout waiting for Ansible playbook execution")
		}

		allComplete := true
		for _, instanceID := range instanceIDs {
			invocation, err := client.GetCommandInvocation(&ssm.GetCommandInvocationInput{
				CommandId:  aws.String(commandID),
				InstanceId: instanceID,
			})

			if err != nil {
				time.Sleep(5 * time.Second)
				allComplete = false
				continue
			}

			status := *invocation.Status
			if status == "InProgress" || status == "Pending" {
				allComplete = false
				continue
			}

			if status == "Success" {
				outputs[*instanceID] = *invocation.StandardOutputContent
				if logger != nil {
					logger.Info(ctx, "Ansible playbook succeeded on instance", map[string]interface{}{
						"instance_id": *instanceID,
					})
				}
			} else {
				outputs[*instanceID] = fmt.Sprintf("Status: %s\nStdout: %s\nStderr: %s",
					status,
					aws.StringValue(invocation.StandardOutputContent),
					aws.StringValue(invocation.StandardErrorContent))
				return nil, fmt.Errorf("Ansible playbook failed on instance %s with status %s", *instanceID, status)
			}
		}

		if allComplete {
			break
		}

		time.Sleep(5 * time.Second)
	}

	if logger != nil {
		logger.Info(ctx, "Ansible playbook execution completed successfully", nil)
	}

	return &AnsiblePlaybookResult{
		CommandID: commandID,
		Status:    "Success",
		Outputs:   outputs,
	}, nil
}
