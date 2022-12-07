package provider

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	lol "github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Ensure ScaffoldingProvider satisfies various provider interfaces.
var _ provider.Provider = neon{}
var _ provider.ProviderWithMetadata = neon{}
var _ provider.ProviderWithSchema = neon{}

// neon defines the provider implementation.
type neon struct {
	version    string
	httpClient http.Client
}

type neonData struct{}

func (p neon) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "neon"
	resp.Version = p.version
}

func (p neon) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = lol.Schema{}
}

func (provider neon) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	if _, ok := os.LookupEnv("NEON_API_KEY"); !ok {
		resp.Diagnostics.AddError(
			"Unable to find token",
			"token cannot be an empty string",
		)
		return
	}
	provider.httpClient = http.Client{
		Timeout: 60 * time.Second,
	}
	//resp.DataSourceData = client
	//resp.ResourceData = client
}

func (p neon) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewEndpointResource,
	}
}

func (p neon) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &neon{
			version: version,
		}
	}
}
