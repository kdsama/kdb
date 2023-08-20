package prom

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

const (
	prometheusFilePath    = "prometheus/prom.yml"
	prometheusDefaultPort = ":8080"
)

type StaticConfig struct {
	Targets []string `yaml:"targets"`
	Labels  map[string]string
}

type ScrapeConfig struct {
	JobName       string         `yaml:"job_name"`
	StaticConfigs []StaticConfig `yaml:"static_configs"`
}

type GlobalConfig struct {
	ScrapeInterval string `yaml:"scrape_interval"`
}

type PrometheusConfig struct {
	Global        GlobalConfig   `yaml:"global"`
	ScrapeConfigs []ScrapeConfig `yaml:"scrape_configs"`
}

func NewPrometheusConfig() *PrometheusConfig {
	config := &PrometheusConfig{
		Global: GlobalConfig{
			ScrapeInterval: "2s",
		}}
	return config
}

func (ps *PrometheusConfig) AddScrapeConfig(name string, port string) {

	ps.ScrapeConfigs = append(ps.ScrapeConfigs, ScrapeConfig{
		JobName: name,
		StaticConfigs: []StaticConfig{
			{
				Targets: []string{name + port},
				Labels:  map[string]string{"instance": name},
			},
		},
	})

}

func (pc *PrometheusConfig) ResetData() {
	os.Remove(prometheusFilePath)
}

func (ps *PrometheusConfig) Generate() {
	yamlBytes, err := yaml.Marshal(ps)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Write the YAML data to a file
	err = os.WriteFile(prometheusFilePath, yamlBytes, 0644)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
}
