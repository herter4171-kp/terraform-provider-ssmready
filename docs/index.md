# ssmready Provider

This provider enables instance configuration within Terraform workflows using AWS Systems Manager.

## Resources

1. **ssmready_ssm_instance_ready** - Waits for instances to join SSM Fleet Manager
2. **ssmready_ansible_playbook** - Runs Ansible playbooks on instances via SSM Run Command

The Ansible resource automatically waits for SSM readiness before executing, allowing you to configure operating systems and install software as part of your infrastructure deployment. This eliminates the need for separate configuration management runs or SSH access - instances only need the SSM agent and appropriate IAM permissions.

## Example Usage

### Basic SSM Wait

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

resource "ssmready_ssm_instance_ready" "example" {
  instance_ids = ["i-0123456789abcdef0"]
  timeout      = 120
  interval     = 10
}
```

### Ansible Playbook Execution

```hcl
provider "ssmready" {
  region = "us-east-1"
}

resource "ssmready_ansible_playbook" "configure" {
  instance_ids = ["i-0123456789abcdef0"]
  
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
  
  timeout  = 600
  interval = 10
}
```