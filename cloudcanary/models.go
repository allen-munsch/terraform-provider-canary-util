package cloudcanary

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// HTTPCheck represents an HTTP check configuration
type HTTPCheck struct {
	ID               types.String            `tfsdk:"id"`
	Name             types.String            `tfsdk:"name"`
	URL              types.String            `tfsdk:"url"`
	Method           types.String            `tfsdk:"method"`
	Headers          types.Map               `tfsdk:"headers"`
	Body             types.String            `tfsdk:"body"`
	ExpectedStatus   types.Int64             `tfsdk:"expected_status"`
	ExpectedResponse types.String            `tfsdk:"expected_response"`
	Interval         types.Int64             `tfsdk:"interval"`
	Timeout          types.Int64             `tfsdk:"timeout"`
	FollowRedirects  types.Bool              `tfsdk:"follow_redirects"`
	Regions          types.List              `tfsdk:"regions"`
	Retries          types.Int64             `tfsdk:"retries"`
	LastResult       types.String            `tfsdk:"last_result"`
	LastCheckTime    types.String            `tfsdk:"last_check_time"`
}

// APICheck represents an API check configuration
type APICheck struct {
	ID                 types.String            `tfsdk:"id"`
	Name               types.String            `tfsdk:"name"`
	Endpoint           types.String            `tfsdk:"endpoint"`
	Method             types.String            `tfsdk:"method"`
	Headers            types.Map               `tfsdk:"headers"`
	Body               types.String            `tfsdk:"body"`
	ExpectedStatus     types.Int64             `tfsdk:"expected_status"`
	ResponseValidation types.List              `tfsdk:"response_validation"`
	Interval           types.Int64             `tfsdk:"interval"`
	Timeout            types.Int64             `tfsdk:"timeout"`
	AuthType           types.String            `tfsdk:"auth_type"`
	AuthValue          types.String            `tfsdk:"auth_value"`
	LastResult         types.String            `tfsdk:"last_result"`
	LastCheckTime      types.String            `tfsdk:"last_check_time"`
}

// CheckResult represents the result of a check execution
type CheckResult struct {
	ID            types.String `tfsdk:"id"`
	CheckID       types.String `tfsdk:"check_id"`
	Status        types.String `tfsdk:"status"`
	ResponseTime  types.Int64  `tfsdk:"response_time"`
	Message       types.String `tfsdk:"message"`
	Timestamp     types.String `tfsdk:"timestamp"`
	Region        types.String `tfsdk:"region"`
	ResponseBody  types.String `tfsdk:"response_body"`
	ResponseCode  types.Int64  `tfsdk:"response_code"`
	FailureReason types.String `tfsdk:"failure_reason"`
}

// CheckResultsDataModel represents the data source for check results
type CheckResultsDataModel struct {
	ID        types.String   `tfsdk:"id"`
	CheckID   types.String   `tfsdk:"check_id"`
	Limit     types.Int64    `tfsdk:"limit"`
	Results   []CheckResult  `tfsdk:"results"`
	StartTime types.String   `tfsdk:"start_time"`
	EndTime   types.String   `tfsdk:"end_time"`
}