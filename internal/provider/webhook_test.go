package provider

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccWebhookResource(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "POST":
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, `{"id": "webhook123"}`)
		case "GET":
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, `{
				"id": "webhook123",
				"clientId": "client123",
				"url": "https://example.com/webhook",
				"types": [1, 2]
			}`)
		case "PATCH":
			w.WriteHeader(http.StatusOK)
		case "DELETE":
			w.WriteHeader(http.StatusNoContent)
		}
	}))
	defer server.Close()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccWebhookResourceConfig(server.URL, "client123", "https://example.com/webhook", "secret123", []int{1, 2}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("urbansportsclub_webhook.test", "client_id", "client123"),
					resource.TestCheckResourceAttr("urbansportsclub_webhook.test", "url", "https://example.com/webhook"),
					resource.TestCheckResourceAttr("urbansportsclub_webhook.test", "shared_secret", "secret123"),
					resource.TestCheckResourceAttr("urbansportsclub_webhook.test", "types.0", "1"),
					resource.TestCheckResourceAttr("urbansportsclub_webhook.test", "types.1", "2"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "urbansportsclub_webhook.test",
				ImportState:       true,
				ImportStateVerify: true,
				// Ignore fields that aren't returned by the API
				ImportStateVerifyIgnore: []string{"shared_secret"},
			},
			// Update testing
			{
				Config: testAccWebhookResourceConfig(server.URL, "client123", "https://example.com/webhook2", "secret456", []int{1, 2}),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("urbansportsclub_webhook.test", "url", "https://example.com/webhook2"),
					resource.TestCheckResourceAttr("urbansportsclub_webhook.test", "shared_secret", "secret456"),
				),
			},
		},
	})
}

func testAccWebhookResourceConfig(baseURL, clientID, url, secret string, types []int) string {
	typeStr := "["
	for i, t := range types {
		if i > 0 {
			typeStr += ", "
		}
		typeStr += fmt.Sprintf("%d", t)
	}
	typeStr += "]"

	return fmt.Sprintf(`
provider "urbansportsclub" {
  base_url = %[1]q
}

resource "urbansportsclub_webhook" "test" {
  client_id     = %[2]q
  url           = %[3]q
  shared_secret = %[4]q
  types         = %[5]s
}
`, baseURL, clientID, url, secret, typeStr)
}

func testAccPreCheck(t *testing.T) {
	// Add any pre-check logic here
}
