# terraform-provider-ssmready

[![Release](https://github.com/herter4171-kp/terraform-provider-ssmready/actions/workflows/release.yml/badge.svg)](https://github.com/herter4171-kp/terraform-provider-ssmready/actions/workflows/release.yml)

This Terraform provider enables SSM-based automation for EC2 instances. It provides two resources:

1. **ssmready_ssm_instance_ready** - Waits for instances to join SSM Fleet Manager
2. **ssmready_ansible_playbook** - Executes Ansible playbooks on instances via SSM Run Command

Only Terraform providers can read the environment variables for AWS credentials, so rather than a `local-exec`, we distilled down with some GPT help from [terraform-provider-ssm](https://github.com/arthurgustin/terraform-provider-ssm) to contain what we need with simple enough syntax to ascertain that credentials and API calls are used strictly for the intended purpose.

## Building

First, we need to get a sense of all of our modules our dependencies rely on by initializing `go.mod`.
```bash
go mod init ssmready
```
Next, we need to populate the list of module URLs in `go.mod`.
```bash
go mod tidy
```
Finally, we can build our binary.
```bash
go build -o terraform-provider-ssmready
```

## Releasing

This provider uses [GoReleaser](https://goreleaser.com/) with GitHub Actions for releases. The version is determined by git tags.

To create a new release:

```bash
# Create and push a version tag (e.g., v0.0.11 for the next version)
git tag v0.0.11
git push origin v0.0.11

# GitHub Actions will automatically:
# 1. Build binaries for all platforms
# 2. Sign the release
# 3. Publish to GitHub releases
# 4. Update the Terraform Registry
```

The version number comes from the git tag, not from any file in the repository.

## Logging

The provider uses structured logging via Terraform's standard logging mechanism. To view detailed logs during execution:

```bash
# Set log level (TRACE, DEBUG, INFO, WARN, ERROR)
export TF_LOG=INFO
terraform apply

# Or for more verbose output
export TF_LOG=DEBUG
terraform apply

# To log to a file
export TF_LOG_PATH=./terraform.log
terraform apply
```

The provider logs important events like:
- Waiting for SSM readiness
- Instance status changes
- Ansible playbook execution progress
- Command completion

## Including the Provider
```hcl
terraform {
  required_providers {
    ssmready = {
      source  = "herter4171-kp/ssmready"
      version = "~> 0.0.10"
    }
  }
}

provider "ssmready" {
  region = "us-east-1"
}
```

## Using the Resources

### Wait for SSM Readiness
```hcl
resource "ssmready_ssm_instance_ready" "example" {
  instance_ids = ["i-abc123"]
  timeout      = 3600 #s default
  interval     = 10   #s default
}
```

### Run Ansible Playbook
```hcl
resource "ssmready_ansible_playbook" "configure" {
  instance_ids = ["i-abc123"]
  
  playbook_content = <<-EOT
    ---
    - hosts: localhost
      connection: local
      tasks:
        - name: Install nginx
          yum:
            name: nginx
            state: present
          become: yes
  EOT
  
  extra_vars = {
    environment = "production"
  }
  
  timeout  = 1800
  interval = 10
}
```

Note: The Ansible playbook resource automatically waits for instances to be ready in SSM, so no separate `ssmready_ssm_instance_ready` resource is needed.