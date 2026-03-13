package auth

import (
	"fmt"
	"os/exec"
	"strings"
)

const kustoResource = "https://api.kusto.windows.net"

// GetToken retrieves an Azure AD access token for Kusto using the Azure CLI.
func GetToken() (string, error) {
	cmd := exec.Command(
		"az", "account", "get-access-token",
		"--resource", kustoResource,
		"--query", "accessToken",
		"-o", "tsv",
	)
	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("az CLI failed: %s", strings.TrimSpace(string(exitErr.Stderr)))
		}
		if execErr, ok := err.(*exec.Error); ok && execErr.Err == exec.ErrNotFound {
			return "", fmt.Errorf("Azure CLI (az) not found: install it from https://aka.ms/installazurecli")
		}
		return "", fmt.Errorf("getting access token: %w", err)
	}

	token := strings.TrimSpace(string(out))
	if token == "" {
		return "", fmt.Errorf("empty token returned: ensure you are logged in with 'az login'")
	}
	return token, nil
}
