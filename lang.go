package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
)

type Language struct {
	Image   string
	Command string
}

var Extensions map[string]Language

func ValidLanguage(ext string) bool {
	for k, _ := range Extensions {
		if k == ext {
			return true
		}
	}

	return false
}

func GetLanguageConfig(filename string) (*Language, error) {
	ext := filepath.Ext(strings.ToLower(filename))

	if !ValidLanguage(ext) {
		return nil, fmt.Errorf("Extension is not supported: %s", ext)
	}

	lang := Extensions[ext]
	return &lang, nil
}

func LoadLanguages(file string) error {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, &Extensions)
	return err
}
