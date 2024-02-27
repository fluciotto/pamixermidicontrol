package configuration

import (
	_ "embed"
	"os"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v5"
	"gopkg.in/yaml.v3"
)

// Embed schema.json in binary
//
//go:embed schema.json
var schema []byte

func Load() (Config, error) {
	var config Config
	var rawConfig map[string]interface{}
	var content []byte
	// Read configuration file
	paths := [...]string{
		"./config.yaml",
		"${HOME}/.config/pamixermidicontrol/config.yaml",
	}
	for _, path := range paths {
		if content != nil {
			break
		}
		content, _ = os.ReadFile(path)
	}
	// Unmarshal and check
	err := yaml.Unmarshal(content, &rawConfig)
	if err != nil {
		return config, err
	}
	err = check(rawConfig)
	if err != nil {
		return config, err
	}
	// Real unmarshal
	err = yaml.Unmarshal(content, &config)
	for i, rule := range config.Rules {
		for j, action := range rule.Actions {
			var iface interface{}
			if action.Type == SetDefaultOutput {
				iface = &Target{}
			} else {
				iface = &TypedTarget{}
			}
			action.RawTarget.Decode(iface)
			if err != nil {
				return config, err
			}
			config.Rules[i].Actions[j].Target = iface
		}
	}
	return config, err
}

func check(configMap map[string]interface{}) error {
	compiler := jsonschema.NewCompiler()
	schemaReader := strings.NewReader(string(schema))
	if err := compiler.AddResource("schema.json", schemaReader); err != nil {
		return err
	}
	schema, err := compiler.Compile("schema.json")
	if err != nil {
		return err
	}
	if err := schema.Validate(configMap); err != nil {
		return err
	}
	return nil
}
