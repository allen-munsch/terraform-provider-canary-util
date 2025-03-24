# Terraform Provider: CloudCanary (Mock Implementation)

This is a mock implementation of a Terraform provider that simulates managing synthetic monitoring checks through a fictional "CloudCanary" platform. This provider is intended for demonstration, educational purposes, and as a starting template for developing real Terraform providers.

> **Note**: This provider doesn't connect to any real service. It simulates API interactions and returns mock data.

## Features (Simulated)

- **HTTP Checks**: Simulate monitoring websites and HTTP endpoints
- **API Checks**: Simulate monitoring API endpoints with JSON validation
- **Multi-region**: Simulate running checks from multiple geographic regions
- **Results Data Source**: Access mock monitoring results within Terraform

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.18

## Building The Provider

1. Clone the repository
```sh
git clone https://github.com/yourorg/terraform-provider-cloudcanary.git
cd terraform-provider-cloudcanary
```

2. Build the provider
```sh
make build
```

3. Install the provider for local use
```sh
make install
```

## Using the provider

Configure the provider:

```hcl
terraform {
  required_providers {
    cloudcanary = {
      source = "yourorg/cloudcanary"
      version = "~> 0.1.0"
    }
  }
}

provider "cloudcanary" {
  # Any non-empty string will work as an API key
  api_key = "test-api-key"
  
  # Optional: specify a custom API base URL (not used in mock implementation)
  # base_url = "https://api.custom-cloudcanary.example.com/v1"
}
```

### HTTP Check Example

```hcl
resource "cloudcanary_http_check" "website" {
  name             = "Company Website"
  url              = "https://example.com"
  method           = "GET"
  expected_status  = 200
  interval         = 60  # seconds
  timeout          = 10  # seconds
  follow_redirects = true
  
  regions = [
    "us-east-1",
    "eu-west-1"
  ]
  
  headers = {
    "User-Agent" = "CloudCanary/1.0"
  }
  
  expected_response = "Welcome to Example"
}
```

### API Check Example

```hcl
resource "cloudcanary_api_check" "api" {
  name            = "Status API"
  endpoint        = "https://api.example.com/v1/status"
  method          = "POST"
  expected_status = 200
  interval        = 300  # seconds
  
  headers = {
    "Content-Type" = "application/json"
  }
  
  body = jsonencode({
    query = "status"
  })
  
  auth_type  = "bearer"
  auth_value = "sample-token"
  
  response_validation = [
    "$.status == 'up'",
    "$.version != null"
  ]
}
```

### Check Results Data Source

```hcl
data "cloudcanary_check_results" "website_results" {
  check_id = cloudcanary_http_check.website.id
  limit    = 10
}

output "latest_status" {
  value = data.cloudcanary_check_results.website_results.results[0].status
}
```

## Resources

### `cloudcanary_http_check`

#### Arguments

- `name` - (Required) Name of the check
- `url` - (Required) URL to check
- `method` - (Optional) HTTP method (GET, POST, etc.). Default: GET
- `headers` - (Optional) Map of HTTP headers
- `body` - (Optional) HTTP request body for POST/PUT requests
- `expected_status` - (Optional) Expected HTTP status code. Default: 200
- `expected_response` - (Optional) Text that should be in the response body
- `interval` - (Optional) Check interval in seconds. Default: 60
- `timeout` - (Optional) Request timeout in seconds. Default: 10
- `follow_redirects` - (Optional) Whether to follow redirects. Default: true
- `regions` - (Optional) List of regions to run the check from
- `retries` - (Optional) Number of retry attempts. Default: 0

#### Attributes

- `id` - Generated unique identifier for the check
- `last_result` - Result of the most recent check (always "PENDING" initially, then "SUCCESS" for mock reads)
- `last_check_time` - Time of the most recent check

### `cloudcanary_api_check`

#### Arguments

- `name` - (Required) Name of the check
- `endpoint` - (Required) API endpoint URL
- `method` - (Optional) HTTP method. Default: GET
- `headers` - (Optional) Map of HTTP headers
- `body` - (Optional) HTTP request body (typically JSON)
- `expected_status` - (Optional) Expected HTTP status code. Default: 200
- `response_validation` - (Optional) List of JSONPath validations
- `interval` - (Optional) Check interval in seconds. Default: 300
- `timeout` - (Optional) Request timeout in seconds. Default: 30
- `auth_type` - (Optional) Authentication type (none, basic, bearer, api_key)
- `auth_value` - (Optional) Authentication value (token, API key, etc.)

#### Attributes

- `id` - Generated unique identifier for the check
- `last_result` - Result of the most recent check (always "PENDING" initially, then "SUCCESS" for mock reads)
- `last_check_time` - Time of the most recent check

### Data Source: `cloudcanary_check_results`

#### Arguments

- `check_id` - (Required) ID of the check to get results for
- `limit` - (Optional) Maximum number of results to return. Default: 10
- `start_time` - (Optional) Start time for results (RFC3339 format, not actually used in the mock)
- `end_time` - (Optional) End time for results (RFC3339 format, not actually used in the mock)

#### Attributes

- `id` - Generated unique identifier for this data source instance
- `results` - List of simulated check results, with the following fields:
  - `id` - Generated unique identifier for the result
  - `check_id` - ID of the check this result belongs to
  - `status` - Alternating result status (SUCCESS, FAILURE)
  - `response_time` - Simulated response time in milliseconds
  - `message` - Generic message associated with the result
  - `timestamp` - Simulated timestamp when the check was executed (RFC3339 format)
  - `region` - Region where the check was executed (if applicable)
  - `response_body` - Response body (if available)
  - `response_code` - HTTP response code (if available)
  - `failure_reason` - Reason for failure (if applicable)

## How This Mock Implementation Works

This provider implements a simulated/mock version of a monitoring service:

1. **No actual HTTP requests are made** - All API operations are simulated within the provider
2. **Resources persist only in Terraform state** - No actual checks are created on any remote system
3. **Generated IDs** - Check IDs are deterministically generated based on names and endpoints
4. **Simulated results** - The data source returns mock check results with alternating success/failure patterns

## Development

### Requirements

- [Go](https://golang.org/doc/install) >= 1.18
- [Terraform](https://www.terraform.io/downloads.html) >= 1.0

### Building

```sh
make build
```

### Testing

```sh
make test
```

### Installing Locally

```sh
make install
```

### Important Implementation Notes

#### Handling Null Values

This provider carefully preserves null values in Terraform state to avoid inconsistencies. When creating or updating resources, it's important to understand that:

- Fields not specified in your configuration will remain `null`
- Default values are only used internally for API calls but not imposed on Terraform state
- This ensures that Terraform's plan and apply mechanisms work correctly and don't detect false changes

#### Sensitive Values

The `auth_value` field for API checks is marked as sensitive and will be stored securely in Terraform state. Its value will not be displayed in logs or console output.