terraform {
  required_providers {
    ssmready = {
      source  = "herter4171-kp/ssmready"
      version = "0.0.11"
    }
  }
}

provider "ssmready" {
  region = var.aws_region
}

variable "aws_region" {
  default = "us-east-1"
}

variable "instance_ids" {
  type        = list(string)
  description = "List of EC2 instance IDs to configure"
}

# Run Ansible playbook to configure instances
# Note: This resource automatically waits for SSM readiness
resource "ssmready_ansible_playbook" "configure" {
  instance_ids = var.instance_ids

  playbook_content = file("${path.module}/playbook.yml")

  # Test vars file with complex data structures
  vars_file_content = file("${path.module}/vars.yml")

  # extra_vars override vars from the file
  extra_vars = {
    environment = "testing"
    app_version = "1.0.0"
  }

  timeout  = 1800
  interval = 10

  # Set to true if playbook output contains sensitive data
  sensitive_output = false
}

output "command_id" {
  value       = ssmready_ansible_playbook.configure.command_id
  description = "SSM command ID for the Ansible playbook execution"
}

output "playbook_status" {
  value       = ssmready_ansible_playbook.configure.status
  description = "Status of the Ansible playbook execution"
}

output "playbook_output" {
  value       = ssmready_ansible_playbook.configure.output
  description = "Output from each instance"
  sensitive   = false
}
