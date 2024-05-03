package config

type ConfigType struct {
	Output              string
	ClusterFilter       string
	ImageNameFilter     string
	NamespaceFilter     string
	QueryFilter         string
	VulnerabilityFilter string
	FilterType          string
}

func (cfg *ConfigType) QueryStrings() map[string]string {
	ret := map[string]string{}
	ret["CLUSTER"] = "r/" + cfg.ClusterFilter
	ret["IMAGE"] = "r/" + cfg.ImageNameFilter
	ret["NAMESPACE"] = cfg.NamespaceFilter
	ret["CVE"] = "r/" + cfg.VulnerabilityFilter
	return ret
}
