package gosnap

import (
	"github.com/pkg/errors"
	"io"
	"os"
	"reflect"
	"runtime"
)

// All the types its fit to print
type Plugin func(FileMapType) error

type GoSnapFile struct {
	Content  []byte
	FileInfo os.FileInfo
	Data     FrontmatterValueType
	Headers  HeaderGetter
}

// Implement io.Writer interface so that plugins can write to the file as if it is a real file
// note: since this only appends if you want to overwrite you need to clear the file first
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
	Read()                        // defined in gosnap_read.go
	ReadFile(string, os.FileInfo) // defined in gosnap_read.go
	Write()                       // defined in gosnap_write.go
	WriteFile(string, GoSnapFile) // defined in gosnap_write.go
	Use(Plugin)
	UseAll(...Plugin)
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
