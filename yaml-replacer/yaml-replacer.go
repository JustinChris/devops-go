package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type yamlFlag struct {
	Key   string
	Value string
	Dir   string
}

func yamlReplacerArgParser() yamlFlag {
	var flags yamlFlag
	flag.StringVar(&flags.Key, "key", "", "Key parameter to replace")
	flag.StringVar(&flags.Value, "value", "", "Value to replace parameter")
	flag.StringVar(&flags.Dir, "dir", "", "Parent directory path")
	flag.Parse()
	return flags
}

func getAllFileInParentDirectory(dirPath string) ([]string, error) {
	var yamlFiles []string
	print(dirPath)
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && (strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml")) {
			yamlFiles = append(yamlFiles, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return yamlFiles, nil
}

func updateYamlData(data *interface{}, key string, value interface{}) {
	switch d := (*data).(type) {
	case map[string]interface{}:
		for k, v := range d {
			if k == key {
				d[k] = value
			} else {
				updateYamlData(&v, key, value)
			}
		}
	case []interface{}:
		for _, v := range d {
			updateYamlData(&v, key, value)
		}
	}
}

func updateYamlFile(yamlFile string, key string, value string) error {
	data, err := os.ReadFile(yamlFile)
	if err != nil {
		return err
	}

	var yamlData interface{}
	err = yaml.Unmarshal(data, &yamlData)
	if err != nil {
		return err
	}

	updateYamlData(&yamlData, key, value)

	yamlBytes, err := yaml.Marshal(yamlData)
	if err != nil {
		return err
	}

	err = os.WriteFile(yamlFile, yamlBytes, 0666)
	if err != nil {
		return err
	}

	fmt.Printf("Updated '%s' in %s\n", key, yamlFile)
	return nil
}

func main() {
	var args yamlFlag = yamlReplacerArgParser()
	print(args.Dir)
	yamlFiles, err := getAllFileInParentDirectory(args.Dir)
	if err != nil {
		fmt.Printf("Error retrieving YAML files from directory: %v\n", err)
		return
	}

	if strings.Contains(args.Key, ".") {
		childKeys := strings.Split(args.Key, ".")
		for _, yamlFile := range yamlFiles {
			err := updateYamlFileWithPath(yamlFile, childKeys, args.Value)
			if err != nil {
				fmt.Printf("Error updating YAML file %s: %v\n", yamlFile, err)
			}
		}
	} else {
		for _, yamlFile := range yamlFiles {
			err := updateYamlFile(yamlFile, args.Key, args.Value)
			if err != nil {
				fmt.Printf("Error updating YAML file %s: %v\n", yamlFile, err)
			}
		}
	}
}

func updateYamlFileWithPath(filepath string, keys []string, value interface{}) error {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return err
	}

	var yamlData interface{}
	err = yaml.Unmarshal(data, &yamlData)
	if err != nil {
		return err
	}

	current := yamlData
	for _, key := range keys[:len(keys)-1] {

		if m, ok := current.(map[string]interface{}); ok {
			if next, exists := m[key]; exists {
				current = next
			}
		} else {
			return fmt.Errorf("invalid path in YAML structure")
		}
	}

	lastKey := keys[len(keys)-1]
	if m, ok := current.(map[string]interface{}); ok {
		if _, exists := m[lastKey]; exists {
			m[lastKey] = value
		}
	} else {
		return fmt.Errorf("invalid path in YAML structure")
	}

	yamlBytes, err := yaml.Marshal(yamlData)
	if err != nil {
		return err
	}

	err = os.WriteFile(filepath, yamlBytes, 0644)
	if err != nil {
		return err
	}

	fmt.Printf("Updated '%s' in %s\n", keys, filepath)
	return nil
}
