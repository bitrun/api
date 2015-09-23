package main

import (
	"fmt"
	"net/http"
	"strings"
)

type Request struct {
	Filename string
	Content  string
	Command  string
	Image    string
}

func normalizeString(val string) string {
	return strings.ToLower(strings.TrimSpace(val))
}

func ParseRequest(r *http.Request) (*Request, error) {
	req := Request{
		Filename: normalizeString(r.FormValue("filename")),
		Command:  normalizeString(r.FormValue("command")),
		Content:  r.FormValue("content"),
	}

	if req.Filename == "" {
		return nil, fmt.Errorf("Filename is required")
	}

	if req.Content == "" {
		return nil, fmt.Errorf("Content is required")
	}

	lang, err := GetLanguageConfig(req.Filename)
	if err != nil {
		return nil, err
	}

	req.Image = lang.Image

	if req.Command == "" {
		req.Command = fmt.Sprintf(lang.Command, req.Filename)
	}

	return &req, nil
}
