package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &WebhookResource{}
var _ resource.ResourceWithImportState = &WebhookResource{}

func NewWebhookResource() resource.Resource {
	return &WebhookResource{}
}

func (r *WebhookResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client, got: %T", req.ProviderData),
		)
		return
	}

	r.client = client
}

type WebhookResource struct {
	client *client
}

type WebhookResourceModel struct {
	ID           types.String `tfsdk:"id"`
	ClientID     types.String `tfsdk:"client_id"`
	URL          types.String `tfsdk:"url"`
	SharedSecret types.String `tfsdk:"shared_secret"`
	Types        []int64      `tfsdk:"types"`
}

func (r *WebhookResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_webhook"
}

func (r *WebhookResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Webhook resource for UrbanSportsClub API",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"client_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Client ID for the webhook",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"url": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "URL for the webhook endpoint",
			},
			"shared_secret": schema.StringAttribute{
				Required:            true,
				Sensitive:           true,
				MarkdownDescription: "Shared secret for webhook authentication",
			},
			"types": schema.ListAttribute{
				Required:            true,
				MarkdownDescription: "Types of webhooks to subscribe to",
				ElementType:         types.Int64Type,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *WebhookResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data WebhookResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	webhookReq := struct {
		ClientID     string  `json:"clientId"`
		URL          string  `json:"url"`
		SharedSecret string  `json:"sharedSecret"`
		Types        []int64 `json:"types"`
	}{
		ClientID:     data.ClientID.ValueString(),
		URL:          data.URL.ValueString(),
		SharedSecret: data.SharedSecret.ValueString(),
		Types:        data.Types,
	}

	jsonBody, err := json.Marshal(webhookReq)
	if err != nil {
		resp.Diagnostics.AddError("JSON Marshal Error", err.Error())
		return
	}

	httpReq, err := http.NewRequest("POST", r.client.baseURL+"/webhook-endpoints", bytes.NewBuffer(jsonBody))
	if err != nil {
		resp.Diagnostics.AddError("Request Creation Error", err.Error())
		return
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpResp, err := r.client.doRequest(httpReq)
	if err != nil {
		resp.Diagnostics.AddError("Request Error", err.Error())
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(httpResp.Body)
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Failed to create webhook. Status: %d, Body: %s", httpResp.StatusCode, string(respBody)))
		return
	}

	var webhookResp struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(httpResp.Body).Decode(&webhookResp); err != nil {
		resp.Diagnostics.AddError("Response Parse Error", err.Error())
		return
	}

	data.ID = types.StringValue(webhookResp.ID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *WebhookResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data WebhookResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpReq, err := http.NewRequest("GET", fmt.Sprintf("%s/webhook-endpoints/%s", r.client.baseURL, data.ID.ValueString()), nil)
	if err != nil {
		resp.Diagnostics.AddError("Request Creation Error", err.Error())
		return
	}

	httpResp, err := r.client.doRequest(httpReq)
	if err != nil {
		resp.Diagnostics.AddError("Request Error", err.Error())
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode == http.StatusNotFound {
		resp.State.RemoveResource(ctx)
		return
	}

	if httpResp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(httpResp.Body)
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Failed to read webhook. Status: %d, Body: %s", httpResp.StatusCode, string(respBody)))
		return
	}

	var webhookResp struct {
		ID       string  `json:"id"`
		ClientID string  `json:"clientId"`
		URL      string  `json:"url"`
		Types    []int64 `json:"types"`
	}
	if err := json.NewDecoder(httpResp.Body).Decode(&webhookResp); err != nil {
		resp.Diagnostics.AddError("Response Parse Error", err.Error())
		return
	}

	data.ID = types.StringValue(webhookResp.ID)
	data.ClientID = types.StringValue(webhookResp.ClientID)
	data.URL = types.StringValue(webhookResp.URL)
	data.Types = webhookResp.Types

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *WebhookResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data WebhookResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	webhookReq := struct {
		URL          string `json:"url"`
		SharedSecret string `json:"sharedSecret"`
	}{
		URL:          data.URL.ValueString(),
		SharedSecret: data.SharedSecret.ValueString(),
	}

	jsonBody, err := json.Marshal(webhookReq)
	if err != nil {
		resp.Diagnostics.AddError("JSON Marshal Error", err.Error())
		return
	}

	httpReq, err := http.NewRequest("PATCH", fmt.Sprintf("%s/webhook-endpoints/%s", r.client.baseURL, data.ID.ValueString()), bytes.NewBuffer(jsonBody))
	if err != nil {
		resp.Diagnostics.AddError("Request Creation Error", err.Error())
		return
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpResp, err := r.client.doRequest(httpReq)
	if err != nil {
		resp.Diagnostics.AddError("Request Error", err.Error())
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(httpResp.Body)
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Failed to update webhook. Status: %d, Body: %s", httpResp.StatusCode, string(respBody)))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *WebhookResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data WebhookResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpReq, err := http.NewRequest("DELETE", fmt.Sprintf("%s/webhook-endpoints/%s", r.client.baseURL, data.ID.ValueString()), nil)
	if err != nil {
		resp.Diagnostics.AddError("Request Creation Error", err.Error())
		return
	}

	httpResp, err := r.client.doRequest(httpReq)
	if err != nil {
		resp.Diagnostics.AddError("Request Error", err.Error())
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusNoContent && httpResp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(httpResp.Body)
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Failed to delete webhook. Status: %d, Body: %s", httpResp.StatusCode, string(respBody)))
		return
	}
}

func (r *WebhookResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
