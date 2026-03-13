package provider

import (
    "context"
    "fmt"
    "time"

    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/service/ssm"
    "github.com/hashicorp/terraform-plugin-sdk/v2/diag"
    "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceInstanceReady() *schema.Resource {
    return &schema.Resource{
        CreateContext: resourceInstanceReadyCreate,
        ReadContext:   schema.NoopContext,
        DeleteContext: schema.NoopContext,
        Schema: map[string]*schema.Schema{
            "instance_ids": {
                Type:     schema.TypeList,
                Required: true,
                Elem:     &schema.Schema{Type: schema.TypeString},
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
        },
    }
}

func resourceInstanceReadyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
    client := meta.(*ssm.SSM)

    instanceIDsRaw := d.Get("instance_ids").([]interface{})
    var instanceIDs []*string
    for _, id := range instanceIDsRaw {
        instanceIDs = append(instanceIDs, aws.String(id.(string)))
    }

    timeout := d.Get("timeout").(int)
    interval := d.Get("interval").(int)

    if err := waitForSSMReady(ctx, client, instanceIDs, timeout, interval); err != nil {
        return diag.FromErr(err)
    }

    d.SetId(fmt.Sprintf("ssm-ready-%d", time.Now().Unix()))
    return nil
}
