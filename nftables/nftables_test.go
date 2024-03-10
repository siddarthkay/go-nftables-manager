package nftables

import (
	"encoding/json"
	"go-nftables-manager/consul"
	"os"
	"testing"
)

func fetchServicesFromTestData() ([]consul.Service, error) {
	data, err := os.ReadFile("../testdata/services.json")
	if err != nil {
		return nil, err
	}

	var services []consul.Service
	err = json.Unmarshal(data, &services)
	if err != nil {
		return nil, err
	}

	return services, nil
}

func TestUpdateFirewallRules(t *testing.T) {
	nft := NewNftables()

	services, err := fetchServicesFromTestData()
	if err != nil {
		t.Fatalf("Failed to fetch services from URL: %v", err)
	}

	// Test configuration present inside nftables/testdata/services.json
	err = nft.UpdateFirewallRules(services)
	if err != nil {
		t.Errorf("UpdateFirewallRules returned an error: %v", err)
	}

	// Handle empty case
	err = nft.UpdateFirewallRules([]consul.Service{})
	if err != nil {
		t.Errorf("UpdateFirewallRules with an empty slice returned an error: %v", err)
	}

}
