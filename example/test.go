package main

import (
	"github.com/caeost/gosnap"
	"path"
	"runtime"
	"strings"
)

func noHi(fileMap gosnap.FileMapType) {
	for _, file := range fileMap {
		result := strings.Replace(string(file.Content[:]), "hi", "noooo", -1)

		file.Content = []byte(result)
	}
}

func whatKey(fileMap gosnap.FileMapType) {
	for _, file := range fileMap {
		if file.Data["key"] != nil {
			file.Content = []byte("key: " + file.Data["key"].(string))
		}
	}
}

func main() {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("Could not figure out position of directory")
	}

	directory := path.Dir(filename)

	site := gosnap.GoSnap{
		Source:      path.Join(directory, "source"),
		Destination: path.Join(directory, "destination"),
		Plugins:     []gosnap.Plugin{noHi, whatKey},
	}

	site.Build()
}
