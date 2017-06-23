package main

import (
	"fmt"
	"github.com/caeost/gosnap"
	"path"
	"path/filepath"
	"strings"
)

func noHi(fileMap gosnap.FileMapType) {
	for _, file := range fileMap {
		result := strings.Replace(string(file.Content[:]), "hi", "noooo", -1)

		file.Content = []byte(result)
	}
}

func whatKey(fileMap gosnap.FileMapType) {
	for fp, file := range fileMap {
		if file.Data["key"] != nil {
			file.Content = []byte("key: " + file.Data["key"].(string))
		}
	}
}

func main() {
	directory, _ := filepath.Abs("./")

	site := gosnap.GoSnap{
		Source:      path.Join(directory, "source"),
		Destination: path.Join(directory, "destination"),
		Clean:       false,
		Plugins:     []gosnap.Plugin{noHi, whatKey},
	}

	site.Build()
}
