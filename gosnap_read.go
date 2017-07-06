package gosnap

import (
	"bytes"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

type FrontmatterValueType map[interface{}]interface{}

type HeaderGetter func() http.Header

// utility functions for reading
func TransformToLocalPath(filePath string, source string) string {
	filePath = path.Clean(filePath)
	internalPath := strings.Replace(filePath, source, "", 1)

	if strings.HasPrefix(internalPath, "/") {
		internalPath = strings.Replace(internalPath, "/", "", 1)
	}

	return internalPath
}

func parseFrontmatter(data []byte) ([]byte, FrontmatterValueType, error) {
	if bytes.HasPrefix(data, []byte("---\n")) {
		splits := bytes.SplitN(data, []byte("\n---\n"), 2)

		if len(splits) != 2 {
			return nil, nil, errors.New("Incorrect format for file. If file includes frontmatter it must start with it and surround it with lines containing only ---")
		}

		frontmatterValues := make(FrontmatterValueType)

		if err := yaml.Unmarshal(splits[0], frontmatterValues); err != nil {
			return nil, nil, errors.Wrap(err, "Could not parse front matter YAML")
		}

		return splits[1], frontmatterValues, nil
	} else {
		return data, nil, nil
	}
}

func parseHeaders(filePath string, frontmatterValues FrontmatterValueType) HeaderGetter {
	return func() http.Header {
		return http.Header{
			// these are response headers that seem relevant to static files
			"Content-Encoding": []string{},
			"Content-Language": []string{},
			"Content-Length":   []string{},
			"Content-Location": []string{},
			"Content-MD5":      []string{},
			"Content-Type":     []string{mime.TypeByExtension(filepath.Ext(filePath))},
			"Last-Modified":    []string{},
			"Set-Cookie":       []string{},
		}
	}
}

var ioUtilReadFile = ioutil.ReadFile

func (gs *GoSnap) ReadFile(path string) (*GoSnapFile, error) {
	data, err := ioUtilReadFile(path)

	if err != nil {
		return &GoSnapFile{}, errors.Wrap(err, "Could not read file from filesystem")
	}

	content, frontmatterValues, yamlErr := parseFrontmatter(data)
	if yamlErr != nil {
		return &GoSnapFile{}, errors.Wrapf(yamlErr, "Error parsing YAML in %v", path)
	}

	headers := parseHeaders(path, frontmatterValues)

	return &GoSnapFile{Content: content, Data: frontmatterValues, Headers: headers}, nil
}

var filepathWalk = filepath.Walk

func (gs *GoSnap) Ignore(ignore string) {
	if gs.IgnoreMap == nil {
		gs.IgnoreMap = make(StringSet)
	}

	gs.IgnoreMap[ignore] = struct{}{}
}

func (gs *GoSnap) IgnoreAll(ignore ...string) {
	for _, ignored := range ignore {
		gs.Ignore(ignored)
	}
}

func (gs *GoSnap) Read() error {
	if gs.Source == "" {
		return errors.New("No Source set in GoSnap object")
	}

	// start over fresh for each build
	gs.FileMap = make(FileMapType)

	readVisitor := func(filePath string, fileInfo os.FileInfo, err error) error {
		if err != nil {
			return errors.Wrapf(err, "Filesystem walk error at %v", filePath)
		}

		if _, ignored := gs.IgnoreMap[filePath]; !ignored && fileInfo != nil && !fileInfo.IsDir() {
			internalPath := TransformToLocalPath(filePath, gs.Source)

			file, err := gs.ReadFile(filePath)

			if err != nil {
				return errors.Wrapf(err, "Could not read file %v", filePath)
			}

			file.FileInfo = fileInfo
			// Format example: "Mon, 02 Jan 2006 15:04:05 MST"
			file.Headers().Set("Last-Modified", fileInfo.ModTime().Format(time.RFC1123))
			gs.FileMap[internalPath] = file
		}

		return nil
	}

	return filepathWalk(gs.Source, readVisitor)
}
