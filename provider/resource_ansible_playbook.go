package provider

import (
	"context"

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

	// Extract configuration from Terraform resource data
	instanceIDsRaw := d.Get("instance_ids").([]interface{})
	var instanceIDs []string
	for _, id := range instanceIDsRaw {
		instanceIDs = append(instanceIDs, id.(string))
	}

	config := AnsiblePlaybookConfig{
		InstanceIDs:     instanceIDs,
		PlaybookContent: d.Get("playbook_content").(string),
		Timeout:         d.Get("timeout").(int),
		Interval:        d.Get("interval").(int),
		SensitiveOutput: d.Get("sensitive_output").(bool),
	}

	if varsFileContent, ok := d.GetOk("vars_file_content"); ok {
		config.VarsFileContent = varsFileContent.(string)
	}

	if extraVarsRaw, ok := d.GetOk("extra_vars"); ok {
		config.ExtraVars = extraVarsRaw.(map[string]interface{})
	}

	// Use shared implementation
	result, err := RunAnsiblePlaybook(ctx, client, config, &tfLogger{ctx: ctx})
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(result.CommandID)
	d.Set("command_id", result.CommandID)
	d.Set("status", result.Status)
	d.Set("output", result.Outputs)

	return nil
}

// tfLogger adapts tflog to the Logger interface
type tfLogger struct {
	ctx context.Context
}

func (l *tfLogger) Info(ctx context.Context, msg string, fields map[string]interface{}) {
	tflog.Info(ctx, msg, fields)
}

func (l *tfLogger) Error(ctx context.Context, msg string, fields map[string]interface{}) {
	tflog.Error(ctx, msg, fields)
}

func resourceAnsiblePlaybookRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Command execution is ephemeral, so we just keep the stored state
	return nil
}
