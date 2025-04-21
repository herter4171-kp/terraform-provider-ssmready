# ssmready Provider

This provider waits for a list of EC2 instances to become ready in AWS Systems Manager Fleet Manager before continuing an apply.

## Example Usage

```hcl
terraform {
  required_providers {
    ssmready = {
      source  = "herter4171-kp/ssmready"
      version = "0.0.4"
    }
  }
}

provider "ssmready" {
  region = "us-gov-east-1"
}

resource "ssmready_ssm_instance_ready" "example" {
  instance_ids = ["i-0123456789abcdef0"]
  timeout      = 120
  interval     = 10
}
```