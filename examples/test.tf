terraform {
  required_providers {
    cloudcanary = {
      source  = "yourorg/cloudcanary"
      version = "~> 0.1.0"
    }
  }
}

provider "cloudcanary" {
  api_key  = "test-api-key"
  base_url = "http://localhost:8080"
}

# HTTP Check for httpbin's status endpoint
resource "cloudcanary_http_check" "httpbin_status" {
  name             = "Httpbin Status Check"
  url              = "http://localhost:8080/status/200"
  method           = "GET"
  expected_status  = 200
  interval         = 30
  timeout          = 5
  follow_redirects = true
  retries          = 1
  
  headers = {
    "User-Agent" = "CloudCanary/1.0"
  }
}

# HTTP Check for httpbin's delayed response
resource "cloudcanary_http_check" "httpbin_delay" {
  name             = "Httpbin Delay Test"
  url              = "http://localhost:8080/delay/1"
  method           = "GET"
  expected_status  = 200
  interval         = 60
  timeout          = 10
  follow_redirects = true
  
  headers = {
    "User-Agent" = "CloudCanary/1.0"
  }
}

# API Check for httpbin's JSON response
resource "cloudcanary_api_check" "httpbin_json" {
  name            = "Httpbin JSON API"
  endpoint        = "http://localhost:8080/json"
  method          = "GET"
  expected_status = 200
  interval        = 60
  timeout         = 5
  
  headers = {
    "Accept" = "application/json"
  }
  
  auth_type  = "none"
  
  response_validation = [
    "$.slideshow != null",
    "$.slideshow.title != null"
  ]
}

# API Check for httpbin's POST endpoint
resource "cloudcanary_api_check" "httpbin_post" {
  name            = "Httpbin POST Test"
  endpoint        = "http://localhost:8080/post"
  method          = "POST"
  expected_status = 200
  interval        = 120
  timeout         = 5
  
  headers = {
    "Content-Type" = "application/json"
    "Accept"       = "application/json"
  }
  
  body = jsonencode({
    test = "value"
    nested = {
      key = "data"
    }
  })
  
  auth_type  = "none"
  
  response_validation = [
    "$.json.test == 'value'",
    "$.json.nested.key == 'data'"
  ]
}

# Check results data source
data "cloudcanary_check_results" "status_results" {
  check_id = cloudcanary_http_check.httpbin_status.id
  limit    = 5
}

# Outputs
output "status_check_id" {
  value = cloudcanary_http_check.httpbin_status.id
}

output "status_check_last_result" {
  value = cloudcanary_http_check.httpbin_status.last_result
}

output "delay_check_id" {
  value = cloudcanary_http_check.httpbin_delay.id
}

output "json_api_check_id" {
  value = cloudcanary_api_check.httpbin_json.id
}

output "latest_results" {
  value = data.cloudcanary_check_results.status_results.results
}