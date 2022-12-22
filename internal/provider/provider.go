package provider

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	lol "github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// Ensure ScaffoldingProvider satisfies various provider interfaces.
var _ provider.Provider = &neon{}

// neon defines the provider implementation.
type neon struct {
	version string
	client  *resty.Client
}

func (p *neon) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "neon"
	resp.Version = p.version
}

func (p *neon) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = lol.Schema{}
}

func (provider *neon) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var key string
	var ok bool
	if key, ok = os.LookupEnv("NEON_API_KEY"); !ok {
		resp.Diagnostics.AddError(
			"Unable to find token",
			"token cannot be an empty string",
		)
		return
	}
	provider.client = resty.New().
		SetBaseURL("https://console.neon.tech/api/v2").
		SetAuthToken(key).
		SetHeader("Content-Type", "application/json").
		AddRetryCondition(func(r *resty.Response, err error) bool {
			return err == nil && r.StatusCode() == 423
		}).
		SetTimeout(30 * time.Second).
		SetRetryCount(3).
		SetRetryWaitTime(10 * time.Second).
		SetError(fmt.Errorf("generic error"))
}

func (p neon) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		func() resource.Resource {
			return &branchResource{
				client: p.client,
			}
		},
		func() resource.Resource {
			return &projectResource{
				client: p.client,
			}
		},
		func() resource.Resource {
			return &endpointResource{
				client: p.client,
			}
		},
		func() resource.Resource {
			return &databaseResource{
				client: p.client,
			}
		},
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
