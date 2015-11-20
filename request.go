package main

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

type Request struct {
	Filename    string
	Content     string
	Command     string
	Image       string
	Format      string
	MemoryLimit int64
}

var FilenameRegexp = regexp.MustCompile(`\A([a-z\d\-\_]+)\.[a-z]{1,6}\z`)

func normalizeString(val string) string {
	return strings.ToLower(strings.TrimSpace(val))
}

func parseInt(val string) int64 {
	if val == "" {
		return 0
	}

	result, err := strconv.Atoi(val)
	if err != nil {
		return 0
	}

	if result < 0 {
		result = 0
	}

	return int64(result)
}

func ParseRequest(r *http.Request) (*Request, error) {
	req := Request{
		Filename:    normalizeString(r.FormValue("filename")),
		Command:     normalizeString(r.FormValue("command")),
		Content:     r.FormValue("content"),
		Image:       r.FormValue("image"),
		MemoryLimit: parseInt(r.FormValue("memory_limit")),
	}

	if req.Filename == "" {
		return nil, fmt.Errorf("Filename is required")
	}

	if !FilenameRegexp.Match([]byte(req.Filename)) {
		return nil, fmt.Errorf("Invalid filename")
	}

	if req.Content == "" {
		return nil, fmt.Errorf("Content is required")
	}

	lang, err := GetLanguageConfig(req.Filename)
	if err != nil {
		return nil, err
	}

	req.Format = lang.Format

	if req.Image == "" {
		req.Image = lang.Image
	}

	if req.Command == "" {
		req.Command = fmt.Sprintf(lang.Command, req.Filename)
	}

	return &req, nil
}
