# ssmready Provider

This provider enables SSM-based automation for EC2 instances with two resources:

1. **ssmready_ssm_instance_ready** - Waits for instances to join SSM Fleet Manager
2. **ssmready_ansible_playbook** - Executes Ansible playbooks via SSM Run Command (includes automatic SSM wait)

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