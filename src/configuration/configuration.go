package configuration

import (
	_ "embed"
	"fmt"
	"os"
	"strings"

	"github.com/santhosh-tekuri/jsonschema/v5"
	"gopkg.in/yaml.v3"
)

// Embed schema.json in binary
//
//go:embed schema.json
var schema []byte

func Load() (Config, string, error) {
	var configPath string
	var config Config
	var rawConfig map[string]interface{}
	var content []byte
	// Read configuration file
	homeDir, _ := os.UserHomeDir()
	paths := [...]string{
		"./config.yaml",
		fmt.Sprintf("%s/.config/pamixermidicontrol/config.yaml", homeDir),
	}
	for _, path := range paths {
		if content != nil {
			break
		}
		fileContent, err := os.ReadFile(path)
		if err == nil {
			configPath = path
			content = fileContent
		}
	}
	if content == nil {
		return config, "", fmt.Errorf("could not find configuration file at %s", paths[len(paths)-1])
	}
	// Unmarshal and check
	err := yaml.Unmarshal(content, &rawConfig)
	if err != nil {
		return config, configPath, err
	}
	err = check(rawConfig)
	if err != nil {
		return config, configPath, err
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
				return config, configPath, err
			}
			config.Rules[i].Actions[j].Target = iface
		}
	}
	return config, configPath, err
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
