package tritonhttp

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

type VHConfigs struct {
	VirtualHosts []struct {
		HostName string `yaml:"hostName"`
		DocRoot  string `yaml:"docRoot"`
	} `yaml:"virtual_hosts"`
}

func ParseVHConfigFile(vhConfigFilePath string, docroot_dirs_path string) map[string]string {
	vh_map := make(map[string]string)
	f, err := ioutil.ReadFile(vhConfigFilePath)

	if err != nil {
		log.Fatalf("could not read config file %s : %v", vhConfigFilePath, err)
	}

	vhostConfigs := VHConfigs{}
	err = yaml.Unmarshal(f, &vhostConfigs)

	for _, vhost := range vhostConfigs.VirtualHosts {
		docroot_path := filepath.Join(docroot_dirs_path, vhost.DocRoot)

		// Check if the path exists
		_, err := os.Stat(docroot_path)
		if err != nil {
			log.Fatalf("path to docroot %s doesn't exist : %v", docroot_path, err)
		}
		vh_map[vhost.HostName] = docroot_path
	}

	return vh_map
}
