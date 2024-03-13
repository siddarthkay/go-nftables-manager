package nftables

import (
	"fmt"
	"go-nftables-manager/consul"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Nftables struct {
	RulesFile string
}

const generatedRulesFile = "nftables.rules"

func NewNftables() *Nftables {
	rulesFile := filepath.Join(".", generatedRulesFile)
	return &Nftables{RulesFile: rulesFile}
}

const (
	envMetrics = "metrics"
	envBackups = "backups"
	envApp     = "app"
	envLogs    = "logs"
)

func (n *Nftables) UpdateFirewallRules(services []consul.Service) error {

	newRules := []string{
		"table ip filter {",
		"  set metrics_servers {",
		"    typeof ip saddr",
		"    elements = { ",
	}

	metricsServers := []string{}
	backupsServers := []string{}
	appServers := []string{}
	logsServers := []string{}

	for _, service := range services {
		env := service.NodeMeta["env"]
		address := service.ServiceAddress

		switch env {
		case envMetrics:
			metricsServers = append(metricsServers, address)
		case envBackups:
			backupsServers = append(backupsServers, address)
		case envApp:
			appServers = append(appServers, address)
		case envLogs:
			logsServers = append(logsServers, address)
		}
	}

	newRules = append(newRules, "      "+strings.Join(metricsServers, ", "))
	newRules = append(newRules, "    }",
		"  }",
		"  set backups_servers {",
		"    typeof ip saddr",
		"    elements = {")
	newRules = append(newRules, "      "+strings.Join(backupsServers, ", "))
	newRules = append(newRules, "    }",
		"  }",
		"  set app_servers {",
		"    typeof ip daddr",
		"    elements = {")
	newRules = append(newRules, "      "+strings.Join(appServers, ", "))
	newRules = append(newRules, "    }",
		"  }",
		"  set logs_servers {",
		"    typeof ip daddr",
		"    elements = {")
	newRules = append(newRules, "      "+strings.Join(logsServers, ", "))
	newRules = append(newRules, "    }",
		"  }",
		"  chain INPUT {")

	rules := []string{
		"    ip saddr @metrics_servers ip daddr @app_servers tcp dport 9104 accept",
		"    ip saddr @backups_servers ip daddr @app_servers tcp dport 3306 accept",
		"    ip daddr @logs_servers tcp dport 5141 accept",
		"    ip daddr @metrics_servers tcp dport 9100 accept",
		"  }",
		"}",
	}

	for _, rule := range rules {
		newRules = append(newRules, rule)
	}

	err := n.writeRulesToFile(newRules)
	if err != nil {
		return fmt.Errorf("failed to write rules to file: %v", err)
	}

	return nil
}

func (n *Nftables) writeRulesToFile(rules []string) error {
	content := []byte(strings.Join(rules, "\n"))
	err := os.WriteFile(n.RulesFile, content, 0644)
	if err != nil {
		return fmt.Errorf("error writing nftables rules file: %v", err)
	}
	return nil
}

func (n *Nftables) ApplyRules() error {
	flushCmd := exec.Command("nft", "flush", "ruleset")
	flushOutput, err := flushCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error flushing nftables ruleset: %v, output: %s", err, flushOutput)
	}

	applyCmd := exec.Command("nft", "-f", n.RulesFile)
	applyOutput, err := applyCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error applying nftables rules: %v, output: %s", err, applyOutput)
	}

	return nil
}
