package filter

import (
	"github.com/kylape/acs-export-example/pkg/config"
	storage "github.com/stackrox/rox/generated/storage"
	"strings"
)

func Filter(deployments []*storage.Deployment, images []*storage.Image, cfg config.ConfigType) (filteredDeployments []*storage.Deployment, filteredImages []*storage.Image, err error) {

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
		if !strings.Contains(image.Name.FullName, cfg.ImageNameFilter) {
			continue
		}

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

		if !vulnFound {
			continue
		}

		filteredImages = append(filteredImages, image)
	}
	return
}
