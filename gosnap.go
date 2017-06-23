package gosnap

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
)

type Plugin func(FileMapType)

type GoSnapFile struct {
	Contents []byte
	FileInfo os.FileInfo
}

type FileMapType map[string]*GoSnapFile

type GoSnapFunctionality interface {
	Read()
	ReadFile(string, os.FileInfo)
	Write()
	WriteFile(string, GoSnapFile)
	Use(Plugin)
	Build()
}

type StringSet map[string]struct{}

type GoSnap struct {
	Source      string
	Destination string
	Clean       bool
	Ignore      []string
	IgnoreMap   StringSet
	FileMap     FileMapType
	Plugins     []Plugin
}

func (gs *GoSnap) transformIgnoreArrayToMap() {
	if gs.Ignore != nil {
		gs.IgnoreMap = make(StringSet)

		for _, entry := range gs.Ignore {
			gs.IgnoreMap[entry] = struct{}{}
		}
	}
}

func (gs *GoSnap) transformToLocalPath(filePath string) string {
	internalPath := strings.Replace(filePath, gs.Source, "", 1)

	if internalPath[1] == '/' {
		internalPath = internalPath[1:]
	}

	return internalPath
}

func (gs *GoSnap) Read() {
	// make sure we have an acceptable set of things to ignore before we start
	gs.transformIgnoreArrayToMap()

	FileMap := gs.FileMap

	readVisitor := func(filePath string, fileInfo os.FileInfo, err error) error {
		if _, ignored := gs.IgnoreMap[filePath]; !ignored && !fileInfo.IsDir() {

			internalPath := gs.transformToLocalPath(filePath)
			fmt.Println("Reading file at", internalPath)
			FileMap[internalPath] = gs.ReadFile(filePath, fileInfo)
		}

		return err
	}

	filepath.Walk(gs.Source, readVisitor)
}

func (gs *GoSnap) ReadFile(path string, fileInfo os.FileInfo) *GoSnapFile {

	data, err := ioutil.ReadFile(path)

	if err != nil {
		panic(err)
	}

	return &GoSnapFile{Contents: data, FileInfo: fileInfo}
}

func (gs *GoSnap) Write() {
	for filePath, file := range gs.FileMap {
		gs.WriteFile(filePath, *file)
	}
}

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
	ioutil.WriteFile(finalPath, file.Contents, perm)
}

func (gs *GoSnap) Use(plugin Plugin) {
	gs.Plugins = append(gs.Plugins, plugin)
}

func (gs *GoSnap) Build() {
	// start over fresh for each build
	gs.FileMap = make(FileMapType)
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
