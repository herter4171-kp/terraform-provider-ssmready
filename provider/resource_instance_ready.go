
package provider

import (
    "context"
    "fmt"
    "log"
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
                Default:  300,
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

    log.Printf("[INFO] Waiting up to %d seconds for instances to become available in SSM", timeout)

    deadline := time.Now().Add(time.Duration(timeout) * time.Second)

    for {
        if time.Now().After(deadline) {
            return diag.Errorf("Timeout exceeded while waiting for SSM instances")
        }

        ready := make(map[string]bool)
        for _, id := range instanceIDs {
            ready[*id] = false
        }

        input := &ssm.DescribeInstanceInformationInput{}
        err := client.DescribeInstanceInformationPages(input,
            func(page *ssm.DescribeInstanceInformationOutput, lastPage bool) bool {
                for _, info := range page.InstanceInformationList {
                    if info.InstanceId != nil && info.PingStatus != nil {
                        if _, ok := ready[*info.InstanceId]; ok && *info.PingStatus == "Online" {
                            ready[*info.InstanceId] = true
                        }
                    }
                }
                return !lastPage
            })
        if err != nil {
            return diag.FromErr(fmt.Errorf("Error describing SSM instance info: %w", err))
        }

        allReady := true
        for _, id := range instanceIDs {
            if !ready[*id] {
                allReady = false
                break
            }
        }

        if allReady {
            log.Printf("[INFO] All instances ready: %v", instanceIDs)
            break
        }

        time.Sleep(time.Duration(interval) * time.Second)
    }

    // Just use a deterministic ID from the list
    d.SetId(fmt.Sprintf("ssm-ready-%d", time.Now().Unix()))
    return nil
}
