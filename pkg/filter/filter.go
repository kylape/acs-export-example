package filter

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/kylape/acs-export-example/pkg/config"
	storage "github.com/stackrox/rox/generated/storage"
)

func ClientVulnFilter(deployments []*storage.Deployment, images []*storage.Image, cfg config.ConfigType) (filteredDeployments []*storage.Deployment, filteredImages []*storage.Image) {
	for _, image := range images {
		vulnFound := false
		if image.Scan != nil {
			for _, component := range image.Scan.Components {
				vulnsToKeep := []*storage.EmbeddedVulnerability{}
				for _, vuln := range component.Vulns {
					if strings.Contains(vuln.Cve, cfg.VulnerabilityFilter) {
						vulnFound = true
						vulnsToKeep = append(vulnsToKeep, vuln)
					}
				}
				component.Vulns = vulnsToKeep
			}
		}

		if vulnFound {
			filteredImages = append(filteredImages, image)
		}
	}

	filteredDeployments = deployments
	return
}

func ClientFilter(deployments []*storage.Deployment, images []*storage.Image, cfg config.ConfigType) (filteredDeployments []*storage.Deployment, filteredImages []*storage.Image) {

	for _, deployment := range deployments {
		if !strings.Contains(deployment.Namespace, cfg.NamespaceFilter) {
			continue
		}

		if !strings.Contains(deployment.ClusterName, cfg.ClusterFilter) {
			continue
		}

		imageFound := false
		for _, container := range deployment.Containers {
			if strings.Contains(container.Image.Name.FullName, cfg.ImageNameFilter) {
				imageFound = true
				continue
			}
		}

		if !imageFound {
			continue
		}

		filteredDeployments = append(filteredDeployments, deployment)
	}

	for _, image := range images {
		if strings.Contains(image.Name.FullName, cfg.ImageNameFilter) {
			filteredImages = append(filteredImages, image)
		}

	}
	return
}

var queryMap = map[string]string{}

func BuildServerQuery(cfg config.ConfigType) string {
	var buffer bytes.Buffer

	for k, v := range cfg.QueryStrings() {
		if strings.TrimPrefix(v, "r/") != "" {
			buffer.WriteString(fmt.Sprintf("%s:%s", k, v))
			buffer.WriteString("+")
		}
	}

	ret := buffer.String()

	if len(ret) > 0 {
		println(ret[:len(ret)-1])
		return ret[:len(ret)-1]
	}
	return ""
}
