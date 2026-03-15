# CLI Test Example

This example demonstrates testing the `ansible-ssm` CLI with a safe, non-destructive playbook that only displays system information.

## Prerequisites

- AWS credentials configured (via environment variables or ~/.aws/credentials)
- EC2 instances with SSM agent installed and running
- Instances must be online in SSM Fleet Manager

## Instance IDs

This example uses the same instance IDs as the Terraform example:
- `i-0a2bcf2526ab300db`
- `i-0a90d2b400f72e5a2`

Region: `us-gov-east-1`

## Test Playbook

The `test-playbook.yml` is completely safe and only:
- Displays system information (OS, version, date/time)
- Echoes a test message
- Shows variables passed via vars file and extra vars

No system changes are made.

## Running the Test

### Basic test (no variables)

```bash
../../ansible-ssm -i i-0a2bcf2526ab300db,i-0a90d2b400f72e5a2 test-playbook.yml
```

### With vars file

```bash
../../ansible-ssm -i i-0a2bcf2526ab300db,i-0a90d2b400f72e5a2 -v test-vars.yml test-playbook.yml
```

### With vars file and extra vars

```bash
../../ansible-ssm \
  -i i-0a2bcf2526ab300db,i-0a90d2b400f72e5a2 \
  -v test-vars.yml \
  -e '{"environment":"production","app_version":"2.0.0"}' \
  test-playbook.yml
```

### Single instance test

```bash
../../ansible-ssm -i i-0a2bcf2526ab300db test-playbook.yml
```

## Expected Output

The CLI will:
1. Wait for instances to be ready in SSM
2. Send the playbook via SSM Run Command
3. Poll for completion
4. Display the output from each instance

You should see:
- System information (OS distribution and version)
- Current timestamp
- Echo message confirming successful execution
- Variable values (if provided)

## Troubleshooting

If instances don't appear ready:
- Verify SSM agent is running on the instances
- Check IAM role has required SSM permissions
- Confirm instances are online in AWS Systems Manager Fleet Manager console
