package provider

import (
    "context"

    "github.com/aws/aws-sdk-go/aws"
    "github.com/aws/aws-sdk-go/aws/session"
    "github.com/aws/aws-sdk-go/service/ssm"
    "github.com/hashicorp/terraform-plugin-sdk/v2/diag"
    "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Provider() *schema.Provider {
    return &schema.Provider{
        Schema: map[string]*schema.Schema{
            "region": {
                Type:     schema.TypeString,
                Required: true,
            },
        },
        ResourcesMap: map[string]*schema.Resource{
            "ssmready_ssm_instance_ready": resourceInstanceReady(),
        },
        ConfigureContextFunc: func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
            region := d.Get("region").(string)

            sess, err := session.NewSession(&aws.Config{
                Region: aws.String(region),
            })
            if err != nil {
                return nil, diag.FromErr(err)
            }

            return ssm.New(sess), nil
        },
    }
}
