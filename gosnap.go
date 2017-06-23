package gosnap

import (
	"bytes"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
)

// All the types its fit to print
type Plugin func(FileMapType)

type FrontmatterValueType map[interface{}]interface{}

type GoSnapFile struct {
	Content  []byte
	FileInfo os.FileInfo
	Data     FrontmatterValueType
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
	Ignore      []string
	IgnoreMap   StringSet
	FileMap     FileMapType
	Plugins     []Plugin
}

// utility functions for reading
func (gs *GoSnap) transformIgnoreArrayToMap() {
	if gs.Ignore != nil {
		gs.IgnoreMap = make(StringSet)

		for _, entry := range gs.Ignore {
			gs.IgnoreMap[entry] = struct{}{}
		}
	}
}
func transformToLocalPath(filePath string, source string) string {
	filePath = path.Clean(filePath)
	internalPath := strings.Replace(filePath, source, "", 1)

	if strings.HasPrefix(internalPath, "/") {
		internalPath = strings.Replace(internalPath, "/", "", 1)
	}

	return internalPath
}

var filepathWalk = filepath.Walk

func (gs *GoSnap) Read() {
	// make sure we have an acceptable set of things to ignore before we start
	gs.transformIgnoreArrayToMap()

	// start over fresh for each build
	gs.FileMap = make(FileMapType)

	readVisitor := func(filePath string, fileInfo os.FileInfo, err error) error {
		if _, ignored := gs.IgnoreMap[filePath]; !ignored && fileInfo != nil && !fileInfo.IsDir() {
			internalPath := transformToLocalPath(filePath, gs.Source)

			fmt.Println("Reading file at", internalPath)
			gs.FileMap[internalPath] = gs.ReadFile(filePath)

			gs.FileMap[internalPath].FileInfo = fileInfo
		}

		return err
	}

	filepathWalk(gs.Source, readVisitor)
}

func parseFrontmatter(data []byte) ([]byte, FrontmatterValueType) {
	if bytes.HasPrefix(data, []byte("---\n")) {
		splits := bytes.SplitN(data, []byte("\n---\n"), 2)

		if len(splits) != 2 {
			panic("Incorrect frontmatter format")
		}

		frontmatterValues := make(FrontmatterValueType)
		err := yaml.Unmarshal(splits[0], frontmatterValues)

		if err != nil {
			panic(err)
		}
		return splits[1], frontmatterValues
	} else {
		return data, nil
	}
}

var ioUtilReadFile = ioutil.ReadFile

func (gs *GoSnap) ReadFile(path string) *GoSnapFile {
	data, err := ioUtilReadFile(path)

	if err != nil {
		panic(err)
	}

	content, frontmatterValues := parseFrontmatter(data)

	return &GoSnapFile{Content: content, Data: frontmatterValues}
}

func (gs *GoSnap) Write() {
	for filePath, file := range gs.FileMap {
		gs.WriteFile(filePath, *file)
	}
}

var mkdirAll = os.MkdirAll
var ioUtilWriteFile = ioutil.WriteFile

func (gs *GoSnap) WriteFile(filePath string, file GoSnapFile) {
	// default permissions for generated files
	// user: read & write
	// group: read
	// other: read
	perm := os.FileMode(0644)

	if file.FileInfo != nil {
		perm = file.FileInfo.Mode()
	}

	finalPath := path.Join(gs.Destination, filePath)
	fmt.Println("Writing file out at", finalPath)

	mkdirAll(path.Dir(finalPath), os.ModePerm)

	ioUtilWriteFile(finalPath, file.Content, perm)
}

func (gs *GoSnap) Use(plugin Plugin) {
	gs.Plugins = append(gs.Plugins, plugin)
}

func (gs *GoSnap) Build() {
	// read all files into map
	gs.Read()
	// run files through plugins
	run(gs.FileMap, gs.Plugins)

	// clean out the output directory if necessary
	if gs.Clean {
		fileInfo, err := os.Stat(gs.Destination)
		if err != nil {
			panic(err)
		}
		os.RemoveAll(gs.Destination)
		os.MkdirAll(gs.Destination, fileInfo.Mode())
	}

	// write out new files
	gs.Write()
}

// from https://stackoverflow.com/questions/7052693/how-to-get-the-name-of-a-function-in-go
func getFunctionName(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}

func run(fileMap FileMapType, plugins []Plugin) {
	for _, plugin := range plugins {
		fmt.Println("Running plugin", getFunctionName(plugin))
		plugin(fileMap)
	}
}
