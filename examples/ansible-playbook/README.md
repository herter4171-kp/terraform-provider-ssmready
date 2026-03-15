# Ansible Playbook Example

This example demonstrates using the `ssmready_ansible_playbook` resource to configure EC2 instances via AWS Systems Manager.

## Using Local Development Build

To test with your local build instead of the published version:

1. Build the provider:
   ```bash
   cd ../..
   make build-provider
   cd examples/ansible-playbook
   ```

2. Use the Terraform CLI config override:
   ```bash
   export TF_CLI_CONFIG_FILE="$(pwd)/.terraformrc"
   ```

3. Run Terraform:
   ```bash
   terraform init
   terraform plan
   terraform apply
   ```

The `.terraformrc` file tells Terraform to use the locally built provider from `../../terraform-provider-ssmready` instead of downloading from the registry.

## Using Published Version

To use the published version from the Terraform Registry, simply don't set the `TF_CLI_CONFIG_FILE` environment variable:

```bash
unset TF_CLI_CONFIG_FILE
terraform init
terraform plan
terraform apply
```

## Configuration

1. Copy the example vars file:
   ```bash
   cp terraform.tfvars.example terraform.tfvars
   ```

2. Edit `terraform.tfvars` with your instance IDs and region

3. Run Terraform as shown above

## What This Example Does

- Waits for instances to be ready in SSM Fleet Manager
- Uploads and executes the Ansible playbook via SSM Run Command
- Uses both a vars file and extra vars
- Returns the command ID, status, and output from each instance
