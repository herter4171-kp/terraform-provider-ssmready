# ssmready_ansible_playbook

Executes an Ansible playbook on EC2 instances via AWS Systems Manager (SSM) Run Command. This resource automatically waits for instances to be ready in SSM before executing the playbook, so no separate `ssmready_ssm_instance_ready` resource is needed.

## Example Usage

```hcl
resource "ssmready_ansible_playbook" "configure_servers" {
  instance_ids = ["i-abc123", "i-def456"]
  
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
        
        - name: Start nginx
          service:
            name: nginx
            state: started
            enabled: yes
          become: yes
  EOT
  
  # Option 1: Use extra_vars for simple key-value pairs
  extra_vars = {
    environment = "production"
    app_version = "1.2.3"
  }
  
  timeout  = 1800
  interval = 10
}
```

### Using a Variables File

For complex variables, you can provide a vars file:

```hcl
resource "ssmready_ansible_playbook" "configure_servers" {
  instance_ids = ["i-abc123", "i-def456"]
  
  playbook_content = file("${path.module}/playbook.yml")
  
  # Option 2: Use vars_file_content for complex data structures
  vars_file_content = file("${path.module}/vars.yml")
  
  # Can combine with extra_vars (extra_vars take precedence)
  extra_vars = {
    environment = "production"
  }
  
  timeout  = 1800
  interval = 10
}
```

## Argument Reference

- `instance_ids` - (Required) List of EC2 instance IDs to run the playbook on. The resource will wait for these instances to be registered with SSM before executing.
- `playbook_content` - (Required) The Ansible playbook YAML content as a string.
- `extra_vars` - (Optional) Map of extra variables to pass to ansible-playbook via `--extra-vars`. Only supports string values.
- `vars_file_content` - (Optional) Content of an Ansible variables file (YAML). Supports complex data structures like lists and nested objects. Passed via `-e @vars.yml`.
- `timeout` - (Optional) Maximum time in seconds to wait for SSM readiness and playbook execution. Default: 3600.
- `interval` - (Optional) Polling interval in seconds while waiting for SSM readiness. Default: 10.
- `sensitive_output` - (Optional) Whether to mark the output as sensitive in Terraform state. Default: false.

Note: If both `vars_file_content` and `extra_vars` are provided, `extra_vars` takes precedence for overlapping keys.

## Attributes Reference

- `command_id` - The SSM command ID for this playbook execution.
- `status` - The final status of the playbook execution.
- `output` - Map of instance IDs to their command output.

## Notes

- Instances must have Ansible installed for this resource to work.
- The playbook runs with `hosts: localhost` and `connection: local` on each instance.
- The resource automatically waits for instances to be online and available in SSM Fleet Manager before executing the playbook.
