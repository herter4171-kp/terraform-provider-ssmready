# terraform-provider-ssmready

[![Release](https://github.com/herter4171-kp/terraform-provider-ssmready/actions/workflows/release.yml/badge.svg)](https://github.com/herter4171-kp/terraform-provider-ssmready/actions/workflows/release.yml)

This Terraform provider enables instance configuration as part of your infrastructure deployment. It provides two resources:

1. **ssmready_ssm_instance_ready** - Waits for instances to join SSM Fleet Manager
2. **ssmready_ansible_playbook** - Runs Ansible playbooks on instances via SSM Run Command

This allows you to configure operating systems and install software directly within your Terraform workflow, without requiring separate configuration management runs or manual SSH access. The provider uses AWS Systems Manager to execute commands, so instances only need the SSM agent and appropriate IAM permissions.

## Why This Approach

Terraform providers can access AWS credentials from the environment, making this more secure than using `local-exec` provisioners. The implementation is focused and straightforward - derived from [terraform-provider-ssm](https://github.com/arthurgustin/terraform-provider-ssm) with help from GPT to keep the code minimal and auditable.

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

1. Create a PR with your changes
2. Merge the PR to `main`
3. Create and push a version tag on the main branch:
   ```bash
   git checkout main
   git pull
   git tag v0.0.11
   git push origin v0.0.11
   ```

GitHub Actions will automatically:
1. Build binaries for all platforms
2. Sign the release
3. Publish to GitHub releases
4. Update the Terraform Registry

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