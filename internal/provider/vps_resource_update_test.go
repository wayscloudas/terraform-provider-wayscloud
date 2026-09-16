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

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

// vpsUpdateTestAPI serves the /v1/vps endpoints the VPS resource uses,
// including PATCH /v1/vps/{id}, which updates display_name and returns the
// instance like GET does. Created instances are active at once, so the resource
// does not poll.
type vpsUpdateTestAPI struct {
	*httptest.Server

	mu        sync.Mutex
	instances map[string]map[string]interface{}
	next      int
	// patches holds the body of each PATCH request, by VPS ID.
	patches map[string][]map[string]interface{}
}

func newVPSUpdateTestAPI(t *testing.T) *vpsUpdateTestAPI {
	t.Helper()
	api := &vpsUpdateTestAPI{
		instances: map[string]map[string]interface{}{},
		patches:   map[string][]map[string]interface{}{},
	}
	api.Server = httptest.NewServer(http.HandlerFunc(api.serve))
	t.Cleanup(api.Close)
	return api
}

func (api *vpsUpdateTestAPI) serve(w http.ResponseWriter, r *http.Request) {
	api.mu.Lock()
	defer api.mu.Unlock()
	w.Header().Set("Content-Type", "application/json")

	if r.Method == http.MethodPost && r.URL.Path == "/v1/vps/" {
		var body vpsCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}
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
			"region":         strings.ToUpper(strings.TrimSpace(body.Region)),
			"os_template":    body.OSTemplate,
			"status":         "active",
			"power_state":    "on",
			// Unique per instance, so a replacement that kept the old values would show.
			"ipv4_address":   fmt.Sprintf("192.0.2.%d", 9+api.next),
			"created_at":     fmt.Sprintf("2026-09-16T12:%02d:00+00:00", api.next-1),
			"provisioned_at": fmt.Sprintf("2026-09-16T12:%02d:30+00:00", api.next-1),
		}
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(api.instances[id])
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/v1/vps/")
	instance, ok := api.instances[id]
	if !strings.HasPrefix(r.URL.Path, "/v1/vps/") || !ok {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"detail": "VPS not found"}`)
		return
	}
	switch r.Method {
	case http.MethodGet:
		_ = json.NewEncoder(w).Encode(instance)
	case http.MethodPatch:
		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, err.Error(), http.StatusUnprocessableEntity)
			return
		}
		api.patches[id] = append(api.patches[id], body)
		if name, ok := body["display_name"]; ok && name != nil {
			instance["display_name"] = name
		}
		_ = json.NewEncoder(w).Encode(instance)
	case http.MethodDelete:
		delete(api.instances, id)
		fmt.Fprintf(w, `{"ok": true, "action": "delete", "vps_id": %q}`, id)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// checkDisplayNamePatched checks that the VPS got exactly one PATCH, carrying
// only the new display name, and that the API stored it.
func (api *vpsUpdateTestAPI) checkDisplayNamePatched(id, want string) resource.TestCheckFunc {
	return func(*terraform.State) error {
		api.mu.Lock()
		defer api.mu.Unlock()
		patches := api.patches[id]
		if len(patches) != 1 {
			return fmt.Errorf("API got %d PATCH requests for %s, want 1: %v", len(patches), id, patches)
		}
		if len(patches[0]) != 1 || patches[0]["display_name"] != want {
			return fmt.Errorf("PATCH body = %v, want {display_name: %q}", patches[0], want)
		}
		if got := api.instances[id]["display_name"]; got != want {
			return fmt.Errorf("API stored display_name %q, want %q", got, want)
		}
		return nil
	}
}

func testVPSDisplayNameConfig(endpoint, displayName string) string {
	return testVPSHostnameConfig(endpoint, "web01.example.com", displayName)
}

func testVPSHostnameConfig(endpoint, hostname, displayName string) string {
	return fmt.Sprintf(`
provider "wayscloud" {
  api_key  = "wayscloud_api_test1234_fakesecretkey"
  endpoint = %q
}

resource "wayscloud_vps" "test" {
  hostname     = %q
  display_name = %q
  plan_code    = "NO-Start-Linux-2cpu-4096mb-30gb"
  region       = "NO"
  os_template  = "ubuntu-24.04"
  ssh_keys     = ["ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIExample"]
}
`, endpoint, hostname, displayName)
}

// TestVPSResource_displayNameUpdate renames an existing VPS. The step must
// update in place, send PATCH /v1/vps/{id}, and leave an empty plan after apply,
// which the test framework checks after every step.
func TestVPSResource_displayNameUpdate(t *testing.T) {
	api := newVPSUpdateTestAPI(t)
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testVPSDisplayNameConfig(api.URL, "one"),
				Check:  resource.TestCheckResourceAttr("wayscloud_vps.test", "display_name", "one"),
			},
			{
				Config: testVPSDisplayNameConfig(api.URL, "two"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("wayscloud_vps.test", plancheck.ResourceActionUpdate),
						// Values that a rename cannot change stay known in the plan.
						plancheck.ExpectKnownValue("wayscloud_vps.test", tfjsonpath.New("created_at"), knownvalue.StringExact("2026-09-16T12:00:00+00:00")),
						plancheck.ExpectKnownValue("wayscloud_vps.test", tfjsonpath.New("ipv4_address"), knownvalue.StringExact("192.0.2.10")),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("wayscloud_vps.test", "id", "vps-1"),
					resource.TestCheckResourceAttr("wayscloud_vps.test", "display_name", "two"),
					resource.TestCheckResourceAttr("wayscloud_vps.test", "status", "active"),
					resource.TestCheckResourceAttr("wayscloud_vps.test", "ssh_keys.#", "1"),
					api.checkDisplayNamePatched("vps-1", "two"),
				),
			},
		},
	})
}

// TestVPSResource_replacementGetsNewComputedValues checks that a replaced VPS
// does not inherit created_at, provisioned_at or the addresses of the old one:
// UseStateForUnknown must not carry them into the plan for the new instance.
func TestVPSResource_replacementGetsNewComputedValues(t *testing.T) {
	api := newVPSUpdateTestAPI(t)
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testVPSHostnameConfig(api.URL, "web01.example.com", "web"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("wayscloud_vps.test", "id", "vps-1"),
					resource.TestCheckResourceAttr("wayscloud_vps.test", "ipv4_address", "192.0.2.10"),
					resource.TestCheckResourceAttr("wayscloud_vps.test", "created_at", "2026-09-16T12:00:00+00:00"),
				),
			},
			{
				Config: testVPSHostnameConfig(api.URL, "web02.example.com", "web"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("wayscloud_vps.test", plancheck.ResourceActionReplace),
						plancheck.ExpectUnknownValue("wayscloud_vps.test", tfjsonpath.New("created_at")),
						plancheck.ExpectUnknownValue("wayscloud_vps.test", tfjsonpath.New("ipv4_address")),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("wayscloud_vps.test", "id", "vps-2"),
					resource.TestCheckResourceAttr("wayscloud_vps.test", "ipv4_address", "192.0.2.11"),
					resource.TestCheckResourceAttr("wayscloud_vps.test", "created_at", "2026-09-16T12:01:00+00:00"),
					resource.TestCheckResourceAttr("wayscloud_vps.test", "provisioned_at", "2026-09-16T12:01:30+00:00"),
				),
			},
		},
	})
}
