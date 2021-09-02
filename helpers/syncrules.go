package helpers

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	log "github.com/sirupsen/logrus"

	"gopkg.in/yaml.v2"
)

type SyncRule struct {
	Direction   string
	Namespaces  []string `yaml:"sourceNamspaces"`
	Annotations map[string]string
	OnConflict  string
	Resources   []string
}

func GetSyncRules(path string) map[string]SyncRule {
	var syncrules map[string]SyncRule
	var content []byte
	// filepath.Walk
	files, err := OSReadDir(path)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}
	for _, file := range files {
		contentTemp, err := ioutil.ReadFile(filepath.Join(path, file))
		if err != nil {
			log.Fatal(err)
		}
		content = append(content, []byte("\n")...)
		content = append(content, contentTemp...)
	}

	yaml.Unmarshal(content, &syncrules)
	return syncrules
}
