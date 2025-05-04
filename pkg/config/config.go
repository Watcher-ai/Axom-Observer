package config

import (
	"io/ioutil"

	"gopkg.in/yaml.v3"
)

type Rules struct {
	Protocols        map[string]ProtocolConfig `yaml:"protocols"`
	OutcomeDetection OutcomeDetection          `yaml:"outcome_detection"`
	BehaviorProfiles []BehaviorProfile         `yaml:"behavior_profiles"`
}

type ProtocolConfig struct {
	Ports []int `yaml:"ports"`
	Paths []struct {
		Pattern string `yaml:"pattern"`
		Extract []struct {
			JSONPath string `yaml:"json_path,omitempty"`
			Metric   string `yaml:"metric,omitempty"`
			Header   string `yaml:"header,omitempty"`
		} `yaml:"extract"`
	} `yaml:"paths"`
	Services []string `yaml:"services"`
}

type OutcomeDetection struct {
	SuccessConditions []struct {
		Protocol     string `yaml:"protocol"`
		StatusCodes  []int  `yaml:"status_codes"`
		ContentMatch string `yaml:"content_match"`
	} `yaml:"success_conditions"`
}

type BehaviorProfile struct {
	Name      string `yaml:"name"`
	Condition string `yaml:"condition"`
	Severity  string `yaml:"severity"`
}

func LoadRules(path string) (*Rules, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var rules Rules
	if err := yaml.Unmarshal(data, &rules); err != nil {
		return nil, err
	}
	return &rules, nil
}
