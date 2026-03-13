package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceAnsiblePlaybook() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceAnsiblePlaybookCreate,
		ReadContext:   resourceAnsiblePlaybookRead,
		DeleteContext: schema.NoopContext,
		Schema: map[string]*schema.Schema{
			"instance_ids": {
				Type:     schema.TypeList,
				Required: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				ForceNew: true,
			},
			"playbook_content": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"extra_vars": {
				Type:     schema.TypeMap,
				Optional: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				ForceNew: true,
			},
			"vars_file_content": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"timeout": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  3600,
				ForceNew: true,
			},
			"interval": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  10,
				ForceNew: true,
			},
			"sensitive_output": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
				ForceNew: true,
			},
			"command_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"output": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceAnsiblePlaybookCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client := meta.(*ssm.SSM)

	instanceIDsRaw := d.Get("instance_ids").([]interface{})
	var instanceIDs []*string
	for _, id := range instanceIDsRaw {
		instanceIDs = append(instanceIDs, aws.String(id.(string)))
	}

	playbookContent := d.Get("playbook_content").(string)
	timeout := d.Get("timeout").(int)
	interval := d.Get("interval").(int)

	// Wait for instances to be ready in SSM first
	if err := waitForSSMReady(ctx, client, instanceIDs, timeout, interval); err != nil {
		return diag.FromErr(err)
	}

	// Build ansible-playbook command
	commands := []string{
		"#!/bin/bash",
		"set -e",
		"cd /tmp",
		fmt.Sprintf("cat > playbook.yml << 'PLAYBOOK_EOF'\n%s\nPLAYBOOK_EOF", playbookContent),
	}

	// Add vars file if provided
	varsFileProvided := false
	if varsFileContent, ok := d.GetOk("vars_file_content"); ok {
		varsFileProvided = true
		commands = append(commands, fmt.Sprintf("cat > vars.yml << 'VARS_EOF'\n%s\nVARS_EOF", varsFileContent.(string)))
	}

	// Build ansible-playbook command with appropriate flags
	ansibleCmd := "ansible-playbook playbook.yml"
	
	if varsFileProvided {
		ansibleCmd += " -e @vars.yml"
	}

	// Add extra vars if provided
	if extraVarsRaw, ok := d.GetOk("extra_vars"); ok {
		extraVars := extraVarsRaw.(map[string]interface{})
		if len(extraVars) > 0 {
			varsJSON, err := json.Marshal(extraVars)
			if err != nil {
				return diag.FromErr(fmt.Errorf("failed to marshal extra_vars: %w", err))
			}
			ansibleCmd += fmt.Sprintf(" --extra-vars '%s'", string(varsJSON))
		}
	}

	commands = append(commands, ansibleCmd)

	tflog.Info(ctx, "Running Ansible playbook", map[string]interface{}{
		"instance_count": len(instanceIDs),
	})

	// Send command via SSM
	sendInput := &ssm.SendCommandInput{
		InstanceIds:  instanceIDs,
		DocumentName: aws.String("AWS-RunShellScript"),
		Parameters: map[string][]*string{
			"commands": {aws.String(commands[0])},
		},
		TimeoutSeconds: aws.Int64(int64(timeout)),
	}

	// Flatten commands into single parameter
	var fullCommand string
	for _, cmd := range commands {
		fullCommand += cmd + "\n"
	}
	sendInput.Parameters["commands"] = []*string{aws.String(fullCommand)}

	sendOutput, err := client.SendCommand(sendInput)
	if err != nil {
		return diag.FromErr(fmt.Errorf("failed to send SSM command: %w", err))
	}

	commandID := *sendOutput.Command.CommandId
	d.Set("command_id", commandID)
	d.SetId(commandID)

	tflog.Info(ctx, "SSM command sent", map[string]interface{}{
		"command_id": commandID,
	})

	// Wait for command completion
	deadline := time.Now().Add(time.Duration(timeout) * time.Second)
	outputs := make(map[string]string)

	for {
		if time.Now().After(deadline) {
			return diag.Errorf("timeout waiting for Ansible playbook execution")
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
				tflog.Info(ctx, "Ansible playbook succeeded on instance", map[string]interface{}{
					"instance_id": *instanceID,
				})
			} else {
				outputs[*instanceID] = fmt.Sprintf("Status: %s\nStdout: %s\nStderr: %s",
					status,
					aws.StringValue(invocation.StandardOutputContent),
					aws.StringValue(invocation.StandardErrorContent))
				return diag.Errorf("Ansible playbook failed on instance %s with status %s", *instanceID, status)
			}
		}

		if allComplete {
			break
		}

		time.Sleep(5 * time.Second)
	}

	d.Set("status", "Success")
	d.Set("output", outputs)

	tflog.Info(ctx, "Ansible playbook execution completed successfully")

	return nil
}

func resourceAnsiblePlaybookRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Command execution is ephemeral, so we just keep the stored state
	return nil
}
