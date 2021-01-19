package gcpclient

import (
	"os/exec"
	"strings"

	"github.com/jessesomerville/ephemeral-iam/internal/appconfig"
)

var (
	proxyAddress = "127.0.0.1"
	proxyPort    = "8084"
)

// ConfigureGcloudProxy configures the current gcloud configuration to use the auth proxy
func ConfigureGcloudProxy() error {

	if err := exec.Command("gcloud", "config", "set", "proxy/address", config.AuthProxy.ProxyAddress).Run(); err != nil {
		return err
	}
	if err := exec.Command("gcloud", "config", "set", "proxy/port", config.AuthProxy.ProxyPort).Run(); err != nil {
		return err
	}
	if err := exec.Command("gcloud", "config", "set", "proxy/type", "http").Run(); err != nil {
		return err
	}
	if err := exec.Command("gcloud", "config", "set", "core/custom_ca_certs_file", appconfig.CertFile).Run(); err != nil {
		return err
	}
	return nil
}

// UnsetGcloudProxy restores the auth proxy changes made to the gcloud config
func UnsetGcloudProxy() error {
	if err := exec.Command("gcloud", "config", "unset", "proxy/address").Run(); err != nil {
		return err
	}
	if err := exec.Command("gcloud", "config", "unset", "proxy/port").Run(); err != nil {
		return err
	}
	if err := exec.Command("gcloud", "config", "unset", "proxy/type").Run(); err != nil {
		return err
	}
	if err := exec.Command("gcloud", "config", "unset", "core/custom_ca_certs_file").Run(); err != nil {
		return err
	}
	return nil
}

// GetCurrentProject gets the active GCP project from the gcloud config
func GetCurrentProject() (string, error) {
	out, err := exec.Command("gcloud", "config", "get-value", "project").Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}