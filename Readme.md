# terraform-provider-ssmready

This Terrafor provider serves a single purpose.  From a given list of instance IDs, it waits for them to join SSH.  Only Terraform providers can read the environment variables for AWS credentials, so rather than a `local-exec`, we distilled down with some GPT help from [terraform-provider-ssm](https://github.com/arthurgustin/terraform-provider-ssm) to just contain what we need with simple enough syntax and brevity (109 lines) to ascertain that credentials and API calls are used strictly for the intended purpose.

## Building
First, we need to get a sense of all of our modules our dependencies rely on by initializing `go.mod`.
```
go mod init ssmready
```
Next, we need to popuate the list of module URLs in `go.mod`.
```
go mod tidy
```
Finally, we can build our binary.
```
go build -o terraform-provider-ssmready
```

## Including the Provider
> TODO: Upload to Terraform registry
```
terraform {
  required_providers {
    ssmready = {
      source  = "local/ssmready"
      version = "v0.0.1"
    }
  }
}
```

## Using the Resource
```
resource "ssmready_ssm_instance_ready" "example" {
  instance_ids = ["i-abc123"]
  timeout      = 120 #s
  interval     = 10  #s
}
```