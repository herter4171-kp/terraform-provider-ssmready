# Ansible Playbook Example

This example demonstrates server configuration using the `ssmready_ansible_playbook` resource.

## What it does

- Waits for EC2 instances to be ready in SSM
- Displays environment and version variables
- Shows system information
- Creates a test file in /tmp with the variables
- Verifies variables were passed correctly

This is a safe, non-destructive test that only creates a file in /tmp.

## Testing Locally

1. Build the provider:
   ```bash
   cd ../..
   go build -o terraform-provider-ssmready
   cd examples/ansible-playbook
   ```

2. Create terraform.tfvars with your instance IDs:
   ```hcl
   instance_ids = ["i-0123456789abcdef0"]
   aws_region   = "us-east-1"
   ```

3. Set AWS credentials (from SSO):
   ```bash
   export AWS_ACCESS_KEY_ID=...
   export AWS_SECRET_ACCESS_KEY=...
   export AWS_SESSION_TOKEN=...
   ```

4. Run with local provider override (skip init):
   ```bash
   export TF_CLI_CONFIG_FILE=$(pwd)/.terraform.rc
   export TF_LOG=INFO
   terraform plan   # No init needed with dev overrides!
   terraform apply
   ```

The `.terraform.rc` file tells Terraform to use the locally built provider at `../../terraform-provider-ssmready` instead of downloading from the registry. With dev overrides active, you skip `terraform init` entirely.

## Files

- `main.tf` - Terraform configuration
- `playbook.yml` - Ansible playbook for server configuration
- `terraform.tfvars.example` - Example variable values
- `.terraform.rc` - Local provider override for testing

## Requirements

- EC2 instances must have the SSM agent installed
- Instances must have Ansible installed
- Appropriate IAM role attached to instances for SSM
