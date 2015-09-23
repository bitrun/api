package main

import (
	"fmt"
	"path/filepath"
	"strings"
)

type Language struct {
	Image   string
	Command string
}

// TODO: This should be extracted
var Extensions = map[string]Language{
	".rb":     Language{"ruby:2.2", "ruby %s"},
	".py":     Language{"python:2.7", "python %s"},
	".js":     Language{"node:0.12", "node %s"},
	".go":     Language{"golang:1.5", "go run %s"},
	".php":    Language{"php:5.6", "php %s"},
	".coffee": Language{"coffescript:0.12", "coffee %s"},
}

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
		return nil, fmt.Errorf("Extension is not supported:", filename)
	}

	lang := Extensions[ext]
	return &lang, nil
}
