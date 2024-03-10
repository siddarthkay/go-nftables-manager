package nftables

import (
	"fmt"
	"go-nftables-manager/consul"
	"os/exec"
	"strings"
)

type Nftables struct{}

func NewNftables() *Nftables {
	return &Nftables{}
}

const (
	envMetrics = "metrics"
	envBackups = "backups"
	envApp     = "app"
	envLogs    = "logs"

	portMySQLExporter = 9104
	portMySQL         = 3306
	portLogstash      = 5141
	portNodeExporter  = 9100
)

func (n *Nftables) UpdateFirewallRules(services []consul.Service) error {
	existingRules, err := n.getExistingRules()
	if err != nil {
		return fmt.Errorf("failed to fetch existing rules: %v", err)
	}

	metricsAddresses := make(map[string]bool)
	backupsAddresses := make(map[string]bool)

	for _, service := range services {
		switch service.NodeMeta["env"] {
		case envMetrics:
			metricsAddresses[service.ServiceAddress] = true
		case envBackups:
			backupsAddresses[service.ServiceAddress] = true
		}
	}

	for _, service := range services {
		env := service.NodeMeta["env"]
		address := service.ServiceAddress

		switch env {
		case envApp:
			if err := n.applyAppRules(address, metricsAddresses, backupsAddresses, existingRules); err != nil {
				return err
			}
		case envLogs:
			if err := n.applyLogstashRule(address, existingRules); err != nil {
				return err
			}
		case envMetrics:
			if err := n.applyNodeExporterRule(address, metricsAddresses, existingRules); err != nil {
				return err
			}
		}
	}

	return nil
}

func (n *Nftables) applyAppRules(address string, metricsAddresses, backupsAddresses map[string]bool, existingRules map[string]bool) error {
	for metricsAddress := range metricsAddresses {
		rule := fmt.Sprintf("add rule inet filter INPUT ip saddr %s ip daddr %s tcp dport %d accept", metricsAddress, address, portMySQLExporter)
		if err := n.applyRule(rule, existingRules); err != nil {
			return fmt.Errorf("failed to apply MySQL exporter access rule: %v", err)
		}
	}
	for backupsAddress := range backupsAddresses {
		rule := fmt.Sprintf("add rule inet filter INPUT ip saddr %s ip daddr %s tcp dport %d accept", backupsAddress, address, portMySQL)
		if err := n.applyRule(rule, existingRules); err != nil {
			return fmt.Errorf("failed to apply MySQL access rule: %v", err)
		}
	}
	return nil
}

func (n *Nftables) applyLogstashRule(address string, existingRules map[string]bool) error {
	rule := fmt.Sprintf("add rule inet filter INPUT ip daddr %s tcp dport %d accept", address, portLogstash)
	if err := n.applyRule(rule, existingRules); err != nil {
		return fmt.Errorf("failed to apply logstash access rule: %v", err)
	}
	return nil
}

func (n *Nftables) applyNodeExporterRule(address string, metricsAddresses map[string]bool, existingRules map[string]bool) error {
	if _, found := metricsAddresses[address]; found {
		rule := fmt.Sprintf("add rule inet filter INPUT ip daddr %s tcp dport %d accept", address, portNodeExporter)
		if err := n.applyRule(rule, existingRules); err != nil {
			return fmt.Errorf("failed to apply node exporter access rule: %v", err)
		}
	}
	return nil
}

func (n *Nftables) applyRule(rule string, existingRules map[string]bool) error {
	if !existingRules[rule] {
		output, err := n.executeCommand(strings.Fields(rule)...)
		if err != nil {
			return fmt.Errorf("error applying rule '%s': %v, output: %s", rule, err, output)
		}
	}
	return nil
}

func (n *Nftables) getExistingRules() (map[string]bool, error) {
	existingRules := make(map[string]bool)
	output, err := n.executeCommand("list", "ruleset")
	if err != nil {
		return nil, fmt.Errorf("error listing nft ruleset: %v, output: %s", err, output)
	}
	rulesOutput := string(output)
	lines := strings.Split(rulesOutput, "\n")
	for _, line := range lines {
		if strings.Contains(line, "tcp dport") && strings.Contains(line, "accept") {
			existingRules[line] = true
		}
	}
	return existingRules, nil
}

func (n *Nftables) executeCommand(args ...string) ([]byte, error) {
	cmd := exec.Command("nft", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return output, fmt.Errorf("error executing nft command: %v, output: %s", err, output)
	}
	return output, nil
}
