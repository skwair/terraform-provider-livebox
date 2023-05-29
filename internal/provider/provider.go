package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/skwair/terraform-provider-livebox/livebox"
)

var _ provider.Provider = &Livebox{}

type Livebox struct {
	version string
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &Livebox{
			version: version,
		}
	}
}

func (l *Livebox) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "livebox"
	resp.Version = l.version
}

func (l *Livebox) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A terraform provider to interact with a Livebox. It currently only supports configuring port forwarding rules.",
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				Optional:    true,
				Description: "URI exposing the Livebox API. May also be provided via LIVEBOX_HOST environment variable.",
			},
			"password": schema.StringAttribute{
				Optional:    true,
				Sensitive:   true,
				Description: "Password for accessing the Livebox API. May also be provided via LIVEBOX_PASSWORD environment variable.",
			},
		},
	}
}

type liveboxProviderModel struct {
	Host     types.String `tfsdk:"host"`
	Password types.String `tfsdk:"password"`
}

func (l *Livebox) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	tflog.Info(ctx, "Configuring Livebox client")

	var config liveboxProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if config.Host.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Unknown Livebox Host",
			"The provider cannot create the Livebox API client as there is an unknown configuration value for the Livebox host. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the LIVEBOX_HOST environment variable.",
		)
	}

	if config.Password.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("password"),
			"Unknown Livebox Password",
			"The provider cannot create the Livebox client as there is an unknown configuration value for the Livebox password. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the LIVEBOX_PASSWORD environment variable.",
		)
	}

	host := os.Getenv("LIVEBOX_HOST")
	password := os.Getenv("LIVEBOX_PASSWORD")

	if !config.Host.IsNull() {
		host = config.Host.ValueString()
	}

	if !config.Password.IsNull() {
		password = config.Password.ValueString()
	}

	if host == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Missing Livebox Host",
			"The provider cannot create the Livebox client as there is a missing or empty value for the Livebox host. "+
				"Set the host value in the configuration or use the LIVEBOX_HOST environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if password == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("password"),
			"Missing Livebox Password",
			"The provider cannot create the Livebox client as there is a missing or empty value for the Livebox password. "+
				"Set the password value in the configuration or use the LIVEBOX_PASSWORD environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	ctx = tflog.SetField(ctx, "livebox_host", host)
	ctx = tflog.SetField(ctx, "livebox_password", password)
	ctx = tflog.MaskFieldValuesWithFieldKeys(ctx, "livebox_password")

	tflog.Debug(ctx, "Creating Livebox client")

	client, err := livebox.NewClient(host, password)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Livebox Client",
			"An unexpected error occurred when creating the Livebox API client. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"Livebox Client Error: "+err.Error(),
		)
		return
	}

	resp.DataSourceData = client
	resp.ResourceData = client

	tflog.Info(ctx, "Configured Livebox client", map[string]any{"success": true})
}

func (l *Livebox) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{}
}

func (l *Livebox) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewPortForwardingResource,
	}
}
