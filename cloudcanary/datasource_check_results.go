package cloudcanary

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// checkResultsDataSource implements a CloudCanary check results data source
type checkResultsDataSource struct {
	client *cloudCanaryClient
}

// Ensure the implementation satisfies the expected interfaces
var _ datasource.DataSource = &checkResultsDataSource{}

// NewCheckResultsDataSource creates a new check results data source
func NewCheckResultsDataSource() datasource.DataSource {
	return &checkResultsDataSource{}
}

// Metadata returns the data source type name
func (d *checkResultsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_check_results"
}

// Schema defines the schema for the data source
func (d *checkResultsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves the results for a specific check.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Unique identifier for this data source instance.",
			},
			"check_id": schema.StringAttribute{
				Required:    true,
				Description: "The ID of the check to retrieve results for.",
			},
			"limit": schema.Int64Attribute{
				Optional:    true,
				Description: "Maximum number of results to return.",
			},
			"start_time": schema.StringAttribute{
				Optional:    true,
				Description: "Start time for results (RFC3339 format).",
			},
			"end_time": schema.StringAttribute{
				Optional:    true,
				Description: "End time for results (RFC3339 format).",
			},
			"results": schema.ListNestedAttribute{
				Computed:    true,
				Description: "The check results.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "Unique identifier for this result.",
						},
						"check_id": schema.StringAttribute{
							Computed:    true,
							Description: "The ID of the check this result belongs to.",
						},
						"status": schema.StringAttribute{
							Computed:    true,
							Description: "The status of the check (SUCCESS, FAILURE).",
						},
						"response_time": schema.Int64Attribute{
							Computed:    true,
							Description: "Response time in milliseconds.",
						},
						"message": schema.StringAttribute{
							Computed:    true,
							Description: "Message associated with the result.",
						},
						"timestamp": schema.StringAttribute{
							Computed:    true,
							Description: "When the check was executed.",
						},
						"region": schema.StringAttribute{
							Computed:    true,
							Description: "Region where the check was executed.",
						},
						"response_body": schema.StringAttribute{
							Computed:    true,
							Description: "Response body (if available).",
						},
						"response_code": schema.Int64Attribute{
							Computed:    true,
							Description: "HTTP response code.",
						},
						"failure_reason": schema.StringAttribute{
							Computed:    true,
							Description: "Reason for failure (if failed).",
						},
					},
				},
			},
		},
	}
}

// Configure adds the provider configured client to the data source
func (d *checkResultsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*cloudCanaryClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *cloudCanaryClient, got: %T", req.ProviderData),
		)
		return
	}

	d.client = client
}

// Read refreshes the Terraform state with the latest data
func (d *checkResultsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config CheckResultsDataModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set default limit if not provided
	limit := 10
	if !config.Limit.IsNull() {
		limit = int(config.Limit.ValueInt64())
	}

	// Call API to get check results
	results, err := d.client.getCheckResults(ctx, config.CheckID.ValueString(), limit)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error retrieving check results",
			fmt.Sprintf("Could not retrieve results for check ID %s: %s", config.CheckID.ValueString(), err),
		)
		return
	}

	// Generate a unique ID for this data source instance
	config.ID = types.StringValue(fmt.Sprintf("results-%s-%d", config.CheckID.ValueString(), time.Now().Unix()))
	
	// Set the results
	config.Results = results

	// Set state
	diags = resp.State.Set(ctx, &config)
	resp.Diagnostics.Append(diags...)
}