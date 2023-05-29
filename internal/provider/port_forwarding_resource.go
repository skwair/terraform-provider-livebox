package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/skwair/terraform-provider-livebox/livebox"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource              = &portForwardingResource{}
	_ resource.ResourceWithConfigure = &portForwardingResource{}
)

// portForwardingResource is the resource implementation.
type portForwardingResource struct {
	client *livebox.Client
}

// NewPortForwardingResource is a helper function to simplify the provider implementation.
func NewPortForwardingResource() resource.Resource {
	return &portForwardingResource{}
}

// Configure adds the provider configured client to the resource.
func (r *portForwardingResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	r.client = req.ProviderData.(*livebox.Client)
}

// Metadata returns the resource type name.
func (r *portForwardingResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_port_forwarding"
}

// Schema defines the schema for the resource.
func (r *portForwardingResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Configure a port forwarding rule on a Livebox.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Arbitrary but unique name of the port forwarding rule to create.",
			},
			"protocol": schema.StringAttribute{
				Required:    true,
				Description: `Protocol of the port forwarding rule to create. Must be one of: "tcp", "udp" or "tcp/udp"`,
			},
			"external_port": schema.Int64Attribute{
				Required:    true,
				Description: "External port of the port forwarding rule to create.",
			},
			"internal_port": schema.Int64Attribute{
				Required:    true,
				Description: "Internal port of the port forwarding rule to create.",
			},
			"port_range": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "If set, forwards a range of consecutive ports instead of one.",
			},
			"destination": schema.StringAttribute{
				Required:    true,
				Description: "IP address to forward traffic to.",
			},
			"enabled": schema.BoolAttribute{
				Required:    true,
				Description: "Whether this port forwarding rule is enabled or not.",
			},
		},
	}
}

type portForwardingModel struct {
	Name         basetypes.StringValue `tfsdk:"name"`
	Protocol     basetypes.StringValue `tfsdk:"protocol"`
	ExternalPort basetypes.Int64Value  `tfsdk:"external_port"`
	InternalPort basetypes.Int64Value  `tfsdk:"internal_port"`
	PortRange    basetypes.Int64Value  `tfsdk:"port_range"`
	Destination  basetypes.StringValue `tfsdk:"destination"`
	Enabled      basetypes.BoolValue   `tfsdk:"enabled"`
}

// Create creates the resource and sets the initial Terraform state.
func (r *portForwardingResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan portForwardingModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	cfg := livebox.PortForwardingConfig{
		Name:         plan.Name.ValueString(),
		ExternalPort: int(plan.ExternalPort.ValueInt64()),
		InternalPort: int(plan.InternalPort.ValueInt64()),
		PortRange:    int(plan.PortRange.ValueInt64()),
		Protocol:     livebox.Protocol(plan.Protocol.ValueString()),
		Destination:  plan.Destination.ValueString(),
		Enabled:      plan.Enabled.ValueBool(),
	}

	err := r.client.UpsertPortForwarding(cfg)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating port forward",
			fmt.Sprintf("Could not create port forward %q, unexpected error: %v", cfg.Name, err),
		)
		return
	}

	plan.Name = types.StringValue(cfg.Name)
	plan.Protocol = types.StringValue(string(cfg.Protocol))
	plan.Destination = types.StringValue(cfg.Destination)
	plan.ExternalPort = types.Int64Value(int64(cfg.ExternalPort))
	plan.InternalPort = types.Int64Value(int64(cfg.InternalPort))
	plan.PortRange = types.Int64Value(int64(cfg.PortRange))
	plan.Enabled = types.BoolValue(cfg.Enabled)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *portForwardingResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state portForwardingModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := state.Name.ValueString()
	pf, err := r.client.GetPortForwarding(name)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error getting port forward start",
			fmt.Sprintf("Could not read state for port forward %q: %v", name, err),
		)
		return
	}

	state.Name = types.StringValue(pf.Name)
	state.Protocol = types.StringValue(string(pf.Protocol))
	state.ExternalPort = types.Int64Value(int64(pf.ExternalPort))
	state.InternalPort = types.Int64Value(int64(pf.InternalPort))
	state.PortRange = types.Int64Value(int64(pf.PortRange))
	state.Destination = types.StringValue(pf.Destination)
	state.Enabled = types.BoolValue(pf.Enabled)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *portForwardingResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan portForwardingModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	cfg := livebox.PortForwardingConfig{
		Name:         plan.Name.ValueString(),
		ExternalPort: int(plan.ExternalPort.ValueInt64()),
		InternalPort: int(plan.InternalPort.ValueInt64()),
		PortRange:    int(plan.PortRange.ValueInt64()),
		Protocol:     livebox.Protocol(plan.Protocol.ValueString()),
		Destination:  plan.Destination.ValueString(),
		Enabled:      plan.Enabled.ValueBool(),
	}

	err := r.client.UpsertPortForwarding(cfg)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating port forward",
			fmt.Sprintf("Could not update port forward %q: %v", cfg.Name, err),
		)
		return
	}

	plan.Name = types.StringValue(cfg.Name)
	plan.Protocol = types.StringValue(string(cfg.Protocol))
	plan.Destination = types.StringValue(cfg.Destination)
	plan.ExternalPort = types.Int64Value(int64(cfg.ExternalPort))
	plan.InternalPort = types.Int64Value(int64(cfg.InternalPort))
	plan.PortRange = types.Int64Value(int64(cfg.PortRange))
	plan.Enabled = types.BoolValue(cfg.Enabled)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *portForwardingResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state portForwardingModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := state.Name.ValueString()
	err := r.client.DeletePortForwarding(name)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting port forward",
			fmt.Sprintf("Could not delete port forward %q: %v", name, err),
		)
		return
	}
}
