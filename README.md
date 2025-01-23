# Terraform Provider Urban Sports Club

This provider allows you to manage Urban Sports Club resources via Terraform.

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.19

## Development

### Building the Provider

1. Clone the repository
2. Enter the repository directory
3. Build the provider using `go install .`

The provider will be built and installed in your `GOBIN` directory.

### Local Provider Configuration

To use the provider locally, create a `.terraformrc` file in your home directory with:

```hcl
provider_installation {
  dev_overrides {
    "qourses/urbansportsclub" = "/Users/USERNAME/go/bin"
  }
  direct {}
}
```

Note: Adjust the path to match your `GOBIN` directory.

### Using the Provider

When using a local build of the provider, skip `terraform init`. You can directly run:

```bash
terraform plan
terraform apply
```

### Testing

```bash
go test ./...
```

## Example Usage

### Provider Configuration

```hcl
terraform {
  required_providers {
    urbansportsclub = {
      source = "qourses/urbansportsclub"
    }
  }
}

provider "urbansportsclub" {
  client_id = "CLIENT_ID"
  client_secret = "CLIENT_SECRET"
}
```

### Managing Webhooks

```hcl
data "urbansportsclub_provider" "current" {}

resource "random_password" "urbansportsclub_space_webhook_secret" {
  length  = 32
  special = false
  upper   = true
}

resource "urbansportsclub_webhook" "example" {
  client_id     = data.urbansportsclub_provider.current.client_id
  url           = "http://example.com/hooking"
  shared_secret = random_password.urbansportsclub_space_webhook_secret.result
  types         = [1]
}
```
