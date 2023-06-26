package config

import (
	"flag"
	"os"
)

type Config struct {
	Listen      string
	HostKeyPath string
	LogLevel    string

	Namespace      string
	ContainerImage string
	PVCSize        string
}

func Load() *Config {
	c := new(Config)
	flag.StringVar(&c.Listen, "listen", envOrDefault("BOOMBOX_LISTEN", ":2828"), "The address the server binds to.")
	flag.StringVar(&c.HostKeyPath, "host-key-path", envOrDefault("BOOMBOX_HOST_KEY_PATH", ".ssh/boombox_ed25519"), "The host key path.")
	flag.StringVar(&c.Namespace, "namespace", envOrDefault("BOOMBOX_NAMESPACE", "default"), "The namespace to create PVCs and Pods (default: default).")
	flag.StringVar(&c.ContainerImage, "container-image", envOrDefault("BOOMBOX_CONTAINER_IMAGE", "ubuntu"), "The Docker image to use in the container (default: ubuntu).")
	flag.StringVar(&c.PVCSize, "pvc-size", envOrDefault("BOOMBOX_PVC_SIZE", "10Gi"), "The size for the user PVC with units (default: 10Gi).")
	flag.StringVar(&c.LogLevel, "log-level", envOrDefault("BOOMBOX_LOG_LEVEL", "info"), "The log level. (default: INFO).")

	return c
}

func envOrDefault(variable, fallback string) string {
	if v, ok := os.LookupEnv(variable); ok {
		return v
	}
	return fallback
}
