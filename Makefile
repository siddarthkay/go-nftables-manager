.PHONY: run-config test

run-config:
	nix-shell shell.nix --run "go run main.go"

test:
	nix-shell shell.nix --run "cd nftables && go test"
