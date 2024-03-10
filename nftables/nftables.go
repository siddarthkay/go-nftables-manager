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
)

func (n *Nftables) UpdateFirewallRules(services []consul.Service) error {
	existingRules, err := n.getExistingRules()
	if err != nil {
		return fmt.Errorf("failed to fetch existing rules: %v", err)
	}

	if err := n.createSets(); err != nil {
		return err
	}

	if err := n.flushSets(); err != nil {
		return err
	}

	for _, service := range services {
		env := service.NodeMeta["env"]
		address := service.ServiceAddress

		switch env {
		case envMetrics:
			if err := n.addToSet("metrics_servers", address); err != nil {
				return err
			}
		case envBackups:
			if err := n.addToSet("backups_servers", address); err != nil {
				return err
			}
		case envApp:
			if err := n.addToSet("app_servers", address); err != nil {
				return err
			}
		case envLogs:
			if err := n.addToSet("logs_servers", address); err != nil {
				return err
			}
		}
	}

	if err := n.applyRules(existingRules); err != nil {
		return err
	}

	return nil
}

func (n *Nftables) createSets() error {
	sets := []string{"metrics_servers", "backups_servers", "app_servers", "logs_servers"}
	for _, set := range sets {
		output, err := n.executeCommand("add", "set", "filter", set, "{", "type", "ipv4_addr;", "}")
		if err != nil {
			return fmt.Errorf("failed to create set '%s': %v, output: %s", set, err, output)
		}
	}
	return nil
}

func (n *Nftables) flushSets() error {
	sets := []string{"metrics_servers", "backups_servers", "app_servers", "logs_servers"}
	for _, set := range sets {
		output, err := n.executeCommand("flush", "set", "filter", set)
		if err != nil {
			return fmt.Errorf("failed to flush set '%s': %v, output: %s", set, err, output)
		}
	}
	return nil
}

func (n *Nftables) addToSet(set, element string) error {
	output, err := n.executeCommand("add", "element", "filter", set, "{", element, "}")
	if err != nil {
		return fmt.Errorf("failed to add element '%s' to set '%s': %v, output: %s", element, set, err, output)
	}
	return nil
}

func (n *Nftables) applyRules(existingRules map[string]bool) error {
	rules := []string{
		"add rule filter INPUT ip saddr @metrics_servers ip daddr @app_servers tcp dport 9104 accept",
		"add rule filter INPUT ip saddr @backups_servers ip daddr @app_servers tcp dport 3306 accept",
		"add rule filter INPUT ip daddr @logs_servers tcp dport 5141 accept",
		"add rule filter INPUT ip daddr @metrics_servers tcp dport 9100 accept",
	}

	for _, rule := range rules {
		if !existingRules[rule] {
			output, err := n.executeCommand(strings.Fields(rule)...)
			if err != nil {
				return fmt.Errorf("failed to apply rule '%s': %v, output: %s", rule, err, output)
			}
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
