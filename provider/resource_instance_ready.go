package provider

import (
    "context"
    "fmt"
    "os"
    "time"

    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/service/ssm"
    "github.com/hashicorp/terraform-plugin-sdk/v2/diag"
    "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

const ssmreadyPrefix = "\033[36m[ssmready]\033[0m" // cyan

func printStatus(format string, a ...interface{}) {
    fmt.Fprintf(os.Stderr, "%s %s\n", ssmreadyPrefix, fmt.Sprintf(format, a...))
}

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

    printStatus("Waiting up to %d seconds for instances to become available in SSM", timeout)

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
            printStatus("All instances ping as online. Waiting for Fleet Manager")
            break
        }

        time.Sleep(time.Duration(interval) * time.Second)
    }

    for _, id := range instanceIDs {
        waitForInventoryPresence(client, *id, 10*time.Minute)
    }

    d.SetId(fmt.Sprintf("ssm-ready-%d", time.Now().Unix()))
    return nil
}

func waitForInventoryPresence(ssmClient *ssm.SSM, instanceID string, timeout time.Duration) error {
    deadline := time.Now().Add(timeout)

    for {
        if time.Now().After(deadline) {
            return fmt.Errorf("Timeout waiting for instance %s to appear in SSM inventory", instanceID)
        }

        resp, err := ssmClient.GetInventory(&ssm.GetInventoryInput{
            Filters: []*ssm.InventoryFilter{
                {
                    Key:    aws.String("AWS:InstanceInformation.InstanceId"),
                    Values: []*string{aws.String(instanceID)},
                    Type:   aws.String("Equal"),
                },
            },
        })

        if err != nil {
            return fmt.Errorf("Error querying inventory for instance %s: %w", instanceID, err)
        }

        if len(resp.Entities) > 0 {
            printStatus("Instance %s is in Fleet Manager", instanceID)
            return nil
        }

        time.Sleep(5 * time.Second)
    }
}
