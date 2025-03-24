package cloudcanary

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces
var (
	_ provider.Provider = &cloudCanaryProvider{}
)

// CloudCanaryProvider defines the provider implementation.
type cloudCanaryProvider struct{}

// New creates a new provider instance
func New() provider.Provider {
	return &cloudCanaryProvider{}
}

// Metadata returns the provider type name.
func (p *cloudCanaryProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "cloudcanary"
}

// Schema defines the provider-level schema for configuration data.
func (p *cloudCanaryProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "CloudCanary provider for monitoring HTTP endpoints and APIs.",
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				Required:    true,
				Sensitive:   true,
				Description: "API key for CloudCanary service.",
			},
			"base_url": schema.StringAttribute{
				Optional:    true,
				Description: "Base URL for the CloudCanary API.",
			},
		},
	}
}

// Configure prepares a CloudCanary client for data sources and resources.
func (p *cloudCanaryProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	tflog.Info(ctx, "Configuring CloudCanary provider")

	// Extract configuration data
	var config providerConfig
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Set defaults
	baseURL := "https://api.cloudcanary.io/v1"
	if !config.BaseURL.IsNull() {
		baseURL = config.BaseURL.ValueString()
	}

	// Initialize the client
	apiKey := config.APIKey.ValueString()
	if apiKey == "" {
		resp.Diagnostics.AddError(
			"Missing API Key",
			"The API key is required to authenticate with CloudCanary.",
		)
		return
	}

	client := &cloudCanaryClient{
		apiKey:  apiKey,
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}

	// Verify authentication
	err := client.verifyAuth(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to authenticate with CloudCanary",
			fmt.Sprintf("Error verifying authentication: %s", err),
		)
		return
	}

	resp.ResourceData = client
	resp.DataSourceData = client

	tflog.Info(ctx, "Configured CloudCanary provider", map[string]any{
		"base_url": baseURL,
	})
}

// Resources defines the resources implemented in the provider.
func (p *cloudCanaryProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewHTTPCheckResource,
		NewAPICheckResource,
	}
}

// DataSources defines the data sources implemented in the provider.
func (p *cloudCanaryProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewCheckResultsDataSource,
	}
}

// providerConfig stores API configuration
type providerConfig struct {
	APIKey  types.String `tfsdk:"api_key"`
	BaseURL types.String `tfsdk:"base_url"`
}