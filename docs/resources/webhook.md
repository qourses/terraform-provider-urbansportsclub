---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "urbansportsclub_webhook Resource - terraform-provider-urbansportsclub"
subcategory: ""
description: |-
  Webhook resource for UrbanSportsClub API
---

# urbansportsclub_webhook (Resource)

Webhook resource for UrbanSportsClub API

See: https://docs.urbansportsclub.io/endpoint/webhooks
for full reference of types

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `client_id` (String) Client ID for the webhook
- `shared_secret` (String, Sensitive) Shared secret for webhook authentication
- `types` (List of Number) Types of webhooks to subscribe to
- `url` (String) URL for the webhook endpoint

### Read-Only

- `id` (String) The ID of this resource.
