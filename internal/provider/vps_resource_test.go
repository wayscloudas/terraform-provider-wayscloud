// Copyright (c) WAYSCloud AS
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestNormalizeRegion(t *testing.T) {
	tests := []struct {
		input, want string
	}{
		// Every entry in regionNormalizeMap, in any case.
		{"oslo", "NO"}, {"Norway", "NO"}, {"no", "NO"}, {"NO-OSLO-1", "NO"},
		{"sweden", "SE"}, {"se", "SE"}, {"Stockholm", "SE"},
		{"france", "FR"}, {"fr", "FR"}, {"paris", "FR"},
		{"germany", "DE"}, {"DE", "DE"}, {"frankfurt", "DE"},
		{"eu", "EU"}, {"us", "US"},
		// Anything else is trimmed and upper-cased, as the API does.
		{"NO", "NO"}, {"nl", "NL"}, {"dk", "DK"}, {" ee ", "EE"}, {"Fi", "FI"},
	}
	for _, tt := range tests {
		if got := normalizeRegion(tt.input); got != tt.want {
			t.Errorf("normalizeRegion(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

// fakeVPSRegionAliases mirrors VPS_REGION_ALIASES in wayscloud-provision-api
// (app/region_codes.py), which POST /v1/vps applies after upper-casing and
// trimming the region.
var fakeVPSRegionAliases = map[string]string{
	"OSLO": "NO", "NORWAY": "NO", "NO-OSLO-1": "NO",
	"STOCKHOLM": "SE", "SWEDEN": "SE",
	"FRANKFURT": "DE", "GERMANY": "DE",
	"PARIS": "FR", "FRANCE": "FR",
}

func fakeVPSRegionFromInput(value string) string {
	region := strings.TrimSpace(strings.ToUpper(value))
	if code, ok := fakeVPSRegionAliases[region]; ok {
		return code
	}
	return region
}

// fakeVPSAPI serves the /v1/vps endpoints the VPS resource uses. Created
// instances are active at once, so the resource does not poll.
type fakeVPSAPI struct {
	*httptest.Server

	mu        sync.Mutex
	instances map[string]map[string]interface{}
	next      int
	// sentRegions holds the region of each create request, as sent.
	sentRegions []string
}

func newFakeVPSAPI(t *testing.T) *fakeVPSAPI {
	t.Helper()
	api := &fakeVPSAPI{instances: map[string]map[string]interface{}{}}
	api.Server = httptest.NewServer(http.HandlerFunc(api.serve))
	t.Cleanup(api.Close)
	return api
}

func (api *fakeVPSAPI) serve(w http.ResponseWriter, r *http.Request) {
	api.mu.Lock()
	defer api.mu.Unlock()
	w.Header().Set("Content-Type", "application/json")

	switch {
	case r.Method == http.MethodPost && r.URL.Path == "/v1/vps/":
		var body vpsCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}
		api.sentRegions = append(api.sentRegions, body.Region)
		api.next++
		id := fmt.Sprintf("vps-%d", api.next)
		displayName := body.DisplayName
		if displayName == "" {
			displayName = body.Hostname
		}
		api.instances[id] = map[string]interface{}{
			"id":             id,
			"provider_vm_id": "vm-" + id,
			"hostname":       body.Hostname,
			"display_name":   displayName,
			"plan_code":      body.PlanCode,
			"region":         fakeVPSRegionFromInput(body.Region),
			"os_template":    body.OSTemplate,
			"status":         "active",
			"power_state":    "on",
			"created_at":     "2026-09-16T12:00:00Z",
		}
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(api.instances[id])
	case strings.HasPrefix(r.URL.Path, "/v1/vps/"):
		id := strings.TrimPrefix(r.URL.Path, "/v1/vps/")
		instance, ok := api.instances[id]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprint(w, `{"detail": "VPS not found"}`)
			return
		}
		switch r.Method {
		case http.MethodGet:
			_ = json.NewEncoder(w).Encode(instance)
		case http.MethodDelete:
			delete(api.instances, id)
			fmt.Fprintf(w, `{"ok": true, "action": "delete", "vps_id": %q}`, id)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	default:
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"detail": "Not Found"}`)
	}
}

// checkStoredRegion checks the region the fake API stored for the VPS in state,
// and that every create request sent the normalized code.
func (api *fakeVPSAPI) checkStoredRegion(want string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources["wayscloud_vps.test"]
		if !ok {
			return fmt.Errorf("wayscloud_vps.test not in state")
		}
		api.mu.Lock()
		defer api.mu.Unlock()
		instance, ok := api.instances[rs.Primary.ID]
		if !ok {
			return fmt.Errorf("API has no VPS %q", rs.Primary.ID)
		}
		if got := instance["region"]; got != want {
			return fmt.Errorf("API stored region %q, want %q", got, want)
		}
		for _, sent := range api.sentRegions {
			if sent != normalizeRegion(sent) {
				return fmt.Errorf("create request sent region %q, want the normalized code", sent)
			}
		}
		return nil
	}
}

func testVPSRegionConfig(endpoint, region string) string {
	return fmt.Sprintf(`
provider "wayscloud" {
  api_key  = "wayscloud_api_test1234_fakesecretkey"
  endpoint = %q
}

resource "wayscloud_vps" "test" {
  hostname    = "web01.example.com"
  plan_code   = "NO-Start-Linux-2cpu-4096mb-30gb"
  region      = %q
  os_template = "ubuntu-24.04"
}
`, endpoint, region)
}

func testVPSProviderFactories() map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		"wayscloud": providerserver.NewProtocol6WithError(New("test")()),
	}
}

// TestVPSResource_regionSpellings creates a VPS with each spelling against a
// local API that normalizes the region the way the real one does. The test
// framework fails a step on "Provider produced invalid plan", on "inconsistent
// result after apply", and when the plan after apply is not empty.
func TestVPSResource_regionSpellings(t *testing.T) {
	for _, tt := range []struct {
		config, stored string
	}{
		{"NO", "NO"},
		{"no", "NO"},
		{"oslo", "NO"},
		{"no-oslo-1", "NO"},
		{"stockholm", "SE"},
		{"frankfurt", "DE"},
		{"paris", "FR"},
		{"NL", "NL"},
		{"nl", "NL"},
		{"dk", "DK"},
		{" ee ", "EE"},
	} {
		t.Run(strings.TrimSpace(tt.config), func(t *testing.T) {
			api := newFakeVPSAPI(t)
			resource.UnitTest(t, resource.TestCase{
				ProtoV6ProviderFactories: testVPSProviderFactories(),
				Steps: []resource.TestStep{
					{
						Config: testVPSRegionConfig(api.URL, tt.config),
						Check: resource.ComposeAggregateTestCheckFunc(
							// State keeps the region as written.
							resource.TestCheckResourceAttr("wayscloud_vps.test", "region", tt.config),
							api.checkStoredRegion(tt.stored),
						),
					},
				},
			})
		})
	}
}

// TestVPSResource_regionRespellingKeepsInstance rewrites the region of an
// existing VPS in another spelling. The plan must be empty: no replacement and
// no in-place update.
func TestVPSResource_regionRespellingKeepsInstance(t *testing.T) {
	for _, tt := range []struct {
		before, after string
	}{
		{"NO", "no"},
		{"no", "NO"},
		{"oslo", "no"},
		{"NL", "nl"},
	} {
		t.Run(tt.before+"_to_"+tt.after, func(t *testing.T) {
			api := newFakeVPSAPI(t)
			resource.UnitTest(t, resource.TestCase{
				ProtoV6ProviderFactories: testVPSProviderFactories(),
				Steps: []resource.TestStep{
					{Config: testVPSRegionConfig(api.URL, tt.before)},
					// PlanOnly fails the step if the plan is not empty.
					{Config: testVPSRegionConfig(api.URL, tt.after), PlanOnly: true},
				},
			})
		})
	}
}

// TestVPSResource_regionChangeReplacesInstance checks that a different region
// still forces a new VPS.
func TestVPSResource_regionChangeReplacesInstance(t *testing.T) {
	api := newFakeVPSAPI(t)
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testVPSProviderFactories(),
		Steps: []resource.TestStep{
			{
				Config: testVPSRegionConfig(api.URL, "NO"),
				Check:  resource.TestCheckResourceAttr("wayscloud_vps.test", "id", "vps-1"),
			},
			{
				Config: testVPSRegionConfig(api.URL, "se"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("wayscloud_vps.test", plancheck.ResourceActionReplace),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("wayscloud_vps.test", "id", "vps-2"),
					resource.TestCheckResourceAttr("wayscloud_vps.test", "region", "se"),
					api.checkStoredRegion("SE"),
				),
			},
		},
	})
}
