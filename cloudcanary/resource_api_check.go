package cloudcanary

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// apiCheckResource implements a CloudCanary API check resource
type apiCheckResource struct {
	client *cloudCanaryClient
}

// Ensure the implementation satisfies the expected interfaces
var _ resource.Resource = &apiCheckResource{}
var _ resource.ResourceWithImportState = &apiCheckResource{}

// NewAPICheckResource creates a new API check resource
func NewAPICheckResource() resource.Resource {
	return &apiCheckResource{}
}

// Metadata returns the resource type name
func (r *apiCheckResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_api_check"
}

// Schema defines the schema for the resource
func (r *apiCheckResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an API check for a web API endpoint.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The unique identifier for this check.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the check.",
			},
			"endpoint": schema.StringAttribute{
				Required:    true,
				Description: "The API endpoint URL to check.",
			},
			"method": schema.StringAttribute{
				Optional:    true,
				Description: "The HTTP method to use (GET, POST, etc.).",
			},
			"headers": schema.MapAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "HTTP headers to include in the request.",
			},
			"body": schema.StringAttribute{
				Optional:    true,
				Description: "HTTP request body, typically JSON for API requests.",
			},
			"expected_status": schema.Int64Attribute{
				Optional:    true,
				Description: "The expected HTTP status code.",
			},
			"response_validation": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "JSONPath validation expressions to validate the response.",
			},
			"interval": schema.Int64Attribute{
				Optional:    true,
				Description: "Check interval in seconds.",
			},
			"timeout": schema.Int64Attribute{
				Optional:    true,
				Description: "Timeout in seconds.",
			},
			"auth_type": schema.StringAttribute{
				Optional:    true,
				Description: "Authentication type (none, basic, bearer, api_key).",
			},
			"auth_value": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "Authentication value (token, API key, etc.).",
			},
			"last_result": schema.StringAttribute{
				Computed:    true,
				Description: "The result of the last check (SUCCESS, FAILURE).",
			},
			"last_check_time": schema.StringAttribute{
				Computed:    true,
				Description: "The time of the last check.",
			},
		},
	}
}

// Configure adds the provider configured client to the resource
func (r *apiCheckResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*cloudCanaryClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *cloudCanaryClient, got: %T", req.ProviderData),
		)
		return
	}

	r.client = client
}

// Create creates a new API check
func (r *apiCheckResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Get the plan
	var plan APICheck
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create a working copy for the API call
	// This allows us to use defaults for the API call without modifying the plan
	apiCheck := APICheck{
		Name:     plan.Name,
		Endpoint: plan.Endpoint,
	}
	
	// Copy all other fields directly from plan
	apiCheck.Method = plan.Method
	apiCheck.Headers = plan.Headers
	apiCheck.Body = plan.Body
	apiCheck.ExpectedStatus = plan.ExpectedStatus
	apiCheck.ResponseValidation = plan.ResponseValidation
	apiCheck.Interval = plan.Interval
	apiCheck.Timeout = plan.Timeout
	apiCheck.AuthType = plan.AuthType
	apiCheck.AuthValue = plan.AuthValue

	// Call the API using the working copy
	err := r.client.createAPICheck(ctx, &apiCheck)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating API check",
			fmt.Sprintf("Could not create API check: %s", err),
		)
		return
	}

	// Now update the original plan with only computed fields
	plan.ID = apiCheck.ID
	plan.LastResult = types.StringValue("PENDING")
	plan.LastCheckTime = types.StringValue(time.Now().Format(time.RFC3339))

	// Set state
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Read refreshes the Terraform state with the latest data
func (r *apiCheckResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state APICheck
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Call API to get the latest data
	apiCheck, err := r.client.readAPICheck(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading API check",
			fmt.Sprintf("Could not read API check ID %s: %s", state.ID.ValueString(), err),
		)
		return
	}

	// Preserve null values in the state - copy only non-null fields from API response
	if !apiCheck.ID.IsNull() {
		state.ID = apiCheck.ID
	}
	if !apiCheck.Name.IsNull() {
		state.Name = apiCheck.Name
	}
	if !apiCheck.Endpoint.IsNull() {
		state.Endpoint = apiCheck.Endpoint
	}
	if !apiCheck.Method.IsNull() {
		state.Method = apiCheck.Method
	}
	if !apiCheck.Headers.IsNull() {
		state.Headers = apiCheck.Headers
	}
	if !apiCheck.Body.IsNull() {
		state.Body = apiCheck.Body
	}
	if !apiCheck.ExpectedStatus.IsNull() {
		state.ExpectedStatus = apiCheck.ExpectedStatus
	}
	if !apiCheck.ResponseValidation.IsNull() {
		state.ResponseValidation = apiCheck.ResponseValidation
	}
	if !apiCheck.Interval.IsNull() {
		state.Interval = apiCheck.Interval
	}
	if !apiCheck.Timeout.IsNull() {
		state.Timeout = apiCheck.Timeout
	}
	if !apiCheck.AuthType.IsNull() {
		state.AuthType = apiCheck.AuthType
	}
	
	// Be extremely careful with sensitive values
	// Only update auth_value if the new value isn't null AND the state value is null
	if !apiCheck.AuthValue.IsNull() && state.AuthValue.IsNull() {
		state.AuthValue = apiCheck.AuthValue
	}
	
	// Always update computed fields
	state.LastResult = apiCheck.LastResult
	state.LastCheckTime = apiCheck.LastCheckTime

	// Set state
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

// Update updates the resource
func (r *apiCheckResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Get plan and current state
	var plan, state APICheck
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Preserve the ID from state
	plan.ID = state.ID

	// Call API to update the check
	err := r.client.updateAPICheck(ctx, &plan)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating API check",
			fmt.Sprintf("Could not update API check ID %s: %s", plan.ID.ValueString(), err),
		)
		return
	}

	// Update computed fields
	plan.LastCheckTime = types.StringValue(time.Now().Format(time.RFC3339))

	// Set state
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Delete deletes the resource
func (r *apiCheckResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Get current state
	var state APICheck
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Call API to delete the check
	err := r.client.deleteAPICheck(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting API check",
			fmt.Sprintf("Could not delete API check ID %s: %s", state.ID.ValueString(), err),
		)
		return
	}

	// Terraform will remove the resource from state
}

// ImportState imports an existing resource into Terraform
func (r *apiCheckResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}