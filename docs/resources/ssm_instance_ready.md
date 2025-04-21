# ssm_instance_ready Resource

Waits for EC2 instances to be fully registered and online in AWS Systems Manager Fleet Manager before allowing the apply to continue.

## Example Usage

```hcl
resource "ssmready_ssm_instance_ready" "example" {
  instance_ids = ["i-0123456789abcdef0"]
  timeout      = 300
  interval     = 10
}
```