package tritonhttp

import (
	"log"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

// VHConfigs is a struct to hold the virtual host configuration
type VHConfigs struct {
	VirtualHosts []struct {
		HostName string `yaml:"hostName"`
		DocRoot  string `yaml:"docRoot"`
	} `yaml:"virtual_hosts"`
}

// ParseVHConfigFile parses the virtual host configuration file (YAML) and returns a map
// of virtual hosts to their docroot paths.
func ParseVHConfigFile(vhConfigFilePath string, docrootDirsPath string) map[string]string {
	// Read the YAML file
	f, err := os.ReadFile(vhConfigFilePath)
	if err != nil {
		log.Fatalf("Failed to read configuration file %s: %v", vhConfigFilePath, err)
	}

	// Unmarshal the YAML file
	vhostConfigs := VHConfigs{}
	if yaml.Unmarshal(f, &vhostConfigs) != nil {
		log.Fatalf("Failed to unmarshal YAML: %v", err)
	}

	// Iterate through the virtual hosts and construct the map
	vhMap := make(map[string]string)
	for _, vhost := range vhostConfigs.VirtualHosts {
		docrootPath := filepath.Join(docrootDirsPath, vhost.DocRoot)

		// Check if the path exists
		_, err := os.Stat(docrootPath)
		if err != nil {
			log.Fatalf("Docroot %s does not exist: %v", docrootPath, err)
		}

		// Add the virtual host to the map
		vhMap[vhost.HostName] = docrootPath
	}

	return vhMap
}
