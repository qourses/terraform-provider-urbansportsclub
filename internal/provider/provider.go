package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ provider.Provider = &UrbanSportsClubProvider{}

type client struct {
	httpClient   *http.Client
	clientID     string
	clientSecret string
	baseURL      string
	accessToken  string
}

type UrbanSportsClubProvider struct {
	version string
}

type UrbanSportsClubProviderModel struct {
	ClientID     types.String `tfsdk:"client_id"`
	ClientSecret types.String `tfsdk:"client_secret"`
}

func newClient(clientID, clientSecret string) (*client, error) {
	c := &client{
		httpClient:   &http.Client{},
		clientID:     clientID,
		clientSecret: clientSecret,
		baseURL:      "https://connect.fitogram.pro",
	}

	if err := c.getAccessToken(); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *client) getAccessToken() error {
	data := url.Values{}
	data.Set("client_id", c.clientID)
	data.Set("client_secret", c.clientSecret)
	data.Set("grant_type", "client_credentials")

	req, err := http.NewRequest("POST", c.baseURL+"/auth", strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("auth failed: %s", body)
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return err
	}

	c.accessToken = tokenResp.AccessToken
	return nil
}

func (c *client) doRequest(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	return c.httpClient.Do(req)
}

func (p *UrbanSportsClubProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config UrbanSportsClubProviderModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	client, err := newClient(
		config.ClientID.ValueString(),
		config.ClientSecret.ValueString(),
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to create client",
			err.Error(),
		)
		return
	}

	resp.DataSourceData = client
	resp.ResourceData = client // This client will be available in Configure of WebhookResource
}

func (p *UrbanSportsClubProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "urbansportsclub"
	resp.Version = p.version
}

func (p *UrbanSportsClubProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"client_id": schema.StringAttribute{
				MarkdownDescription: "The client ID for API authentication",
				Required:            true,
				Sensitive:           true,
			},
			"client_secret": schema.StringAttribute{
				MarkdownDescription: "The client secret for API authentication",
				Required:            true,
				Sensitive:           true,
			},
		},
	}
}

func (p *UrbanSportsClubProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		func() resource.Resource {
			return NewWebhookResource() // Remove client parameter
		},
	}
}

func (p *UrbanSportsClubProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewProviderDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &UrbanSportsClubProvider{
			version: version,
		}
	}
}
