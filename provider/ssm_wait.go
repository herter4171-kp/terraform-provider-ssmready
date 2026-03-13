package provider

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// waitForSSMReady waits for instances to be online and available in SSM Fleet Manager
func waitForSSMReady(ctx context.Context, client *ssm.SSM, instanceIDs []*string, timeout, interval int) error {
	tflog.Info(ctx, "Waiting for instances to become available in SSM", map[string]interface{}{
		"timeout_seconds": timeout,
		"instance_count":  len(instanceIDs),
	})

	deadline := time.Now().Add(time.Duration(timeout) * time.Second)

	// Wait for instances to be online
	for {
		if time.Now().After(deadline) {
			return fmt.Errorf("timeout exceeded while waiting for SSM instances")
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
			return fmt.Errorf("error describing SSM instance info: %w", err)
		}

		allReady := true
		for _, id := range instanceIDs {
			if !ready[*id] {
				allReady = false
				break
			}
		}

		if allReady {
			tflog.Info(ctx, "All instances ping as online, waiting for Fleet Manager")
			break
		}

		time.Sleep(time.Duration(interval) * time.Second)
	}

	// Wait for Fleet Manager inventory
	for _, id := range instanceIDs {
		if err := waitForInventoryPresence(ctx, client, *id, 10*time.Minute); err != nil {
			return err
		}
	}

	return nil
}

func waitForInventoryPresence(ctx context.Context, ssmClient *ssm.SSM, instanceID string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	for {
		if time.Now().After(deadline) {
			return fmt.Errorf("timeout waiting for instance %s to appear in SSM inventory", instanceID)
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
			return fmt.Errorf("error querying inventory for instance %s: %w", instanceID, err)
		}

		if len(resp.Entities) > 0 {
			tflog.Info(ctx, "Instance is in Fleet Manager", map[string]interface{}{
				"instance_id": instanceID,
			})
			return nil
		}

		time.Sleep(5 * time.Second)
	}
}
