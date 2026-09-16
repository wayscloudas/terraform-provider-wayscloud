// Copyright (c) WAYSCloud AS
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// Body shape of GET /v1/regions in wayscloud-provision-api (app/api_regions_public.py).
const testRegionsBody = `{
  "regions": [
    {"code": "no", "name": "Norge", "city": "Oslo", "country": "NO", "status": "active",
     "available_services": ["storage", "databases", "redis", "apps", "llm"]},
    {"code": "fi", "name": "Finland", "city": "Helsinki", "country": "FI", "status": "maintenance",
     "available_services": []},
    {"code": "de", "name": "Tyskland", "city": "Frankfurt", "country": "DE", "status": "maintenance"}
  ],
  "total": 3
}`

// TestRegionsDataSource_publicEndpoint reads the data source against a local
// server, so it runs without credentials or network access.
func TestRegionsDataSource_publicEndpoint(t *testing.T) {
	var mu sync.Mutex
	var requested []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		requested = append(requested, r.Method+" "+r.URL.Path)
		mu.Unlock()
		if r.Method != http.MethodGet || r.URL.Path != "/v1/regions" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, testRegionsBody)
	}))
	defer server.Close()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"wayscloud": providerserver.NewProtocol6WithError(New("test")()),
		},
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
provider "wayscloud" {
  api_key  = "wayscloud_api_test1234_fakesecretkey"
  endpoint = %q
}

data "wayscloud_regions" "all" {}

output "app_regions" {
  value = join(",", [for r in data.wayscloud_regions.all.regions : r.code if contains(r.available_services, "apps")])
}
`, server.URL),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.wayscloud_regions.all", "regions.#", "3"),

					resource.TestCheckResourceAttr("data.wayscloud_regions.all", "regions.0.code", "no"),
					resource.TestCheckResourceAttr("data.wayscloud_regions.all", "regions.0.name", "Norge"),
					resource.TestCheckResourceAttr("data.wayscloud_regions.all", "regions.0.city", "Oslo"),
					resource.TestCheckResourceAttr("data.wayscloud_regions.all", "regions.0.country", "NO"),
					resource.TestCheckResourceAttr("data.wayscloud_regions.all", "regions.0.status", "active"),
					resource.TestCheckResourceAttr("data.wayscloud_regions.all", "regions.0.available", "true"),
					resource.TestCheckResourceAttr("data.wayscloud_regions.all", "regions.0.available_services.#", "5"),
					resource.TestCheckResourceAttr("data.wayscloud_regions.all", "regions.0.available_services.3", "apps"),

					resource.TestCheckResourceAttr("data.wayscloud_regions.all", "regions.1.code", "fi"),
					resource.TestCheckResourceAttr("data.wayscloud_regions.all", "regions.1.status", "maintenance"),
					resource.TestCheckResourceAttr("data.wayscloud_regions.all", "regions.1.available", "false"),
					resource.TestCheckResourceAttr("data.wayscloud_regions.all", "regions.1.available_services.#", "0"),

					// A region without available_services still gets an empty list, not null.
					resource.TestCheckResourceAttr("data.wayscloud_regions.all", "regions.2.code", "de"),
					resource.TestCheckResourceAttr("data.wayscloud_regions.all", "regions.2.available", "false"),
					resource.TestCheckResourceAttr("data.wayscloud_regions.all", "regions.2.available_services.#", "0"),

					resource.TestCheckOutput("app_regions", "no"),
				),
			},
		},
	})

	mu.Lock()
	defer mu.Unlock()
	if len(requested) == 0 {
		t.Fatal("the data source made no request")
	}
	for _, r := range requested {
		if r != "GET /v1/regions" {
			t.Errorf("unexpected request %q; want only GET /v1/regions", r)
		}
	}
}
