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

// httpCheckResource implements a CloudCanary HTTP check resource
type httpCheckResource struct {
	client *cloudCanaryClient
}

// Ensure the implementation satisfies the expected interfaces
var _ resource.Resource = &httpCheckResource{}
var _ resource.ResourceWithImportState = &httpCheckResource{}

// NewHTTPCheckResource creates a new HTTP check resource
func NewHTTPCheckResource() resource.Resource {
	return &httpCheckResource{}
}

// Metadata returns the resource type name
func (r *httpCheckResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_http_check"
}

// Schema defines the schema for the resource
func (r *httpCheckResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an HTTP check for a website or endpoint.",
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
			"url": schema.StringAttribute{
				Required:    true,
				Description: "The URL to check.",
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
				Description: "HTTP request body for POST/PUT requests.",
			},
			"expected_status": schema.Int64Attribute{
				Optional:    true,
				Description: "The expected HTTP status code.",
			},
			"expected_response": schema.StringAttribute{
				Optional:    true,
				Description: "Text that should be present in the response body.",
			},
			"interval": schema.Int64Attribute{
				Optional:    true,
				Description: "Check interval in seconds.",
			},
			"timeout": schema.Int64Attribute{
				Optional:    true,
				Description: "Timeout in seconds.",
			},
			"follow_redirects": schema.BoolAttribute{
				Optional:    true,
				Description: "Whether to follow HTTP redirects.",
			},
			"regions": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "Regions to run the check from.",
			},
			"retries": schema.Int64Attribute{
				Optional:    true,
				Description: "Number of retries before marking as failed.",
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
func (r *httpCheckResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create creates a new HTTP check
func (r *httpCheckResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Get the plan
	var plan HTTPCheck
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Create a working copy for the API call
	// This allows us to use defaults for the API call without modifying the plan
	apiCheck := HTTPCheck{
		Name: plan.Name,
		URL:  plan.URL,
	}
	
	// Copy all other fields directly from plan
	apiCheck.Method = plan.Method
	apiCheck.Headers = plan.Headers
	apiCheck.Body = plan.Body
	apiCheck.ExpectedStatus = plan.ExpectedStatus
	apiCheck.ExpectedResponse = plan.ExpectedResponse
	apiCheck.Interval = plan.Interval
	apiCheck.Timeout = plan.Timeout
	apiCheck.FollowRedirects = plan.FollowRedirects
	apiCheck.Regions = plan.Regions
	apiCheck.Retries = plan.Retries

	// Call the API using the working copy
	err := r.client.createHTTPCheck(ctx, &apiCheck)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating HTTP check",
			fmt.Sprintf("Could not create HTTP check: %s", err),
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
func (r *httpCheckResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state HTTPCheck
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Call API to get the latest data
	apiCheck, err := r.client.readHTTPCheck(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading HTTP check",
			fmt.Sprintf("Could not read HTTP check ID %s: %s", state.ID.ValueString(), err),
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
	if !apiCheck.URL.IsNull() {
		state.URL = apiCheck.URL
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
	if !apiCheck.ExpectedResponse.IsNull() {
		state.ExpectedResponse = apiCheck.ExpectedResponse
	}
	if !apiCheck.Interval.IsNull() {
		state.Interval = apiCheck.Interval
	}
	if !apiCheck.Timeout.IsNull() {
		state.Timeout = apiCheck.Timeout
	}
	if !apiCheck.FollowRedirects.IsNull() {
		state.FollowRedirects = apiCheck.FollowRedirects
	}
	if !apiCheck.Regions.IsNull() {
		state.Regions = apiCheck.Regions
	}
	if !apiCheck.Retries.IsNull() {
		state.Retries = apiCheck.Retries
	}
	
	// Always update computed fields
	state.LastResult = apiCheck.LastResult
	state.LastCheckTime = apiCheck.LastCheckTime

	// Set state
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

// Update updates the resource
func (r *httpCheckResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Get plan and current state
	var plan, state HTTPCheck
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
	err := r.client.updateHTTPCheck(ctx, &plan)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating HTTP check",
			fmt.Sprintf("Could not update HTTP check ID %s: %s", plan.ID.ValueString(), err),
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
func (r *httpCheckResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Get current state
	var state HTTPCheck
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Call API to delete the check
	err := r.client.deleteHTTPCheck(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting HTTP check",
			fmt.Sprintf("Could not delete HTTP check ID %s: %s", state.ID.ValueString(), err),
		)
		return
	}

	// Terraform will remove the resource from state
}

// ImportState imports an existing resource into Terraform
func (r *httpCheckResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}