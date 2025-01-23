package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &ProviderDataSource{}

func NewProviderDataSource() datasource.DataSource {
	return &ProviderDataSource{}
}

type ProviderDataSource struct {
	client *client
}

type ProviderDataSourceModel struct {
	ID       types.String `tfsdk:"id"`
	ClientID types.String `tfsdk:"client_id"`
}

func (d *ProviderDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ProviderDataSourceModel
	data.ID = types.StringValue("provider")
	data.ClientID = types.StringValue(d.client.clientID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (d *ProviderDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_provider"
}

func (d *ProviderDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Provider data source",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"client_id": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (d *ProviderDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			"Expected *client, got: %T",
		)
		return
	}

	d.client = client
}
