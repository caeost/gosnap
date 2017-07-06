package gosnap

import (
	"bytes"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"time"
)

// default permissions for generated files
// user: read & write
// group: read
// other: read
const (
	DEFAULT_PERM = os.FileMode(0644)
)

// All the types its fit to print
type Plugin func(FileMapType) error

type FrontmatterValueType map[interface{}]interface{}

type GoSnapFile struct {
	Content  []byte
	FileInfo os.FileInfo
	Data     FrontmatterValueType
	Headers  http.Header
}

// Implement io.Writer interface so that plugins can write to the file as if it is a real file
func (gsf *GoSnapFile) Write(p []byte) (n int, err error) {
	gsf.Content = append(gsf.Content, p...)

	return len(p), nil
}

// Implement io.Read interface so that plugins can read from the file as if it is a real file
func (gsf *GoSnapFile) Read(p []byte) (n int, err error) {
	copy(p, gsf.Content)

	return len(p), io.EOF
}

type FileMapType map[string]*GoSnapFile

type StringSet map[string]struct{}

// Exposed API
type GoSnapFunctionality interface {
	Read()
	ReadFile(string, os.FileInfo)
	Write()
	WriteFile(string, GoSnapFile)
	Use(Plugin)
	Build()
}

// Structure of the main object
type GoSnap struct {
	Source      string
	Destination string
	Clean       bool
	IgnoreMap   StringSet
	FileMap     FileMapType
	Plugins     []Plugin
}

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

func parseHeaders(filePath string, frontmatterValues FrontmatterValueType) http.Header {
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
			return errors.Wrapf(err, "Filesystem walk error found at %v", filePath)
		}

		if _, ignored := gs.IgnoreMap[filePath]; !ignored && fileInfo != nil && !fileInfo.IsDir() {
			internalPath := TransformToLocalPath(filePath, gs.Source)

			file, err := gs.ReadFile(filePath)

			if err != nil {
				return errors.Wrapf(err, "Could not read file %v", filePath)
			}

			file.FileInfo = fileInfo
			// Format example: "Mon, 02 Jan 2006 15:04:05 MST"
			file.Headers.Set("Last-Modified", fileInfo.ModTime().Format(time.RFC1123))
			gs.FileMap[internalPath] = file
		}

		return nil
	}

	return filepathWalk(gs.Source, readVisitor)
}

var mkdirAll = os.MkdirAll
var ioUtilWriteFile = ioutil.WriteFile

func (gs *GoSnap) WriteFile(filePath string, file GoSnapFile) error {
	if gs.Destination == "" {
		return errors.New("No Destination set in GoSnap object")
	}

	perm := DEFAULT_PERM
	if file.FileInfo != nil {
		perm = file.FileInfo.Mode()
	}

	finalPath := path.Join(gs.Destination, filePath)

	mkdirErr := mkdirAll(path.Dir(finalPath), os.ModePerm)

	if mkdirErr != nil {
		return errors.Wrapf(mkdirErr, "Could not create required directories for %v", finalPath)
	}

	return ioUtilWriteFile(finalPath, file.Content, perm)
}

func cleanOutput(destination string) error {
	fileInfo, err := os.Stat(destination)
	if err != nil {
		return err
	}

	err = os.RemoveAll(destination)
	if err != nil {
		return err
	}

	err = os.MkdirAll(destination, fileInfo.Mode())
	if err != nil {
		return err
	}

	return nil
}

func (gs *GoSnap) Write() error {
	if gs.Destination == "" {
		return errors.New("No Destination set in GoSnap object")
	}

	// clean out the output directory if necessary
	if gs.Clean {
		if err := cleanOutput(gs.Destination); err != nil {
			return errors.Wrapf(err, "Could not clean output directory %v before write", gs.Destination)
		}
	}

	for filePath, file := range gs.FileMap {
		err := gs.WriteFile(filePath, *file)

		if err != nil {
			return errors.Wrapf(err, "Exiting because of failure to write file %v", filePath)
		}
	}

	return nil
}

func (gs *GoSnap) Use(plugin Plugin) {
	if gs.Plugins == nil {
		gs.Plugins = []Plugin{}
	}

	gs.Plugins = append(gs.Plugins, plugin)
}

func (gs *GoSnap) UseAll(plugins ...Plugin) {
	for _, plugin := range plugins {
		gs.Use(plugin)
	}
}

// from https://stackoverflow.com/questions/7052693/how-to-get-the-name-of-a-function-in-go
func getFunctionName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}

func Run(fileMap FileMapType, plugins []Plugin) error {
	for _, plugin := range plugins {
		pluginName := getFunctionName(plugin)

		if err := plugin(fileMap); err != nil {
			return errors.Wrapf(err, "Exiting because of problem in plugin %v", pluginName)
		}
	}

	return nil
}

func (gs *GoSnap) Build() (err error) {
	// read all files into map
	err = gs.Read()

	if err != nil {
		return errors.Wrap(err, "Build failed at read step")
	}

	// run files through plugins
	err = Run(gs.FileMap, gs.Plugins)

	if err != nil {
		return errors.Wrap(err, "Build failed during plugin run")
	}

	// write out new files
	err = gs.Write()

	if err != nil {
		return errors.Wrap(err, "Build failed writing files")
	}

	return nil
}
