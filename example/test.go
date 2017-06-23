package main

import (
	"fmt"
	"github.com/caeost/gosnap"
	"path"
	"path/filepath"
	"strings"
)

func noHi(fileMap gosnap.FileMapType) {
	for filePath, file := range fileMap {
		result := strings.Replace(string(file.Contents[:]), "hi", "noooo", -1)
		fmt.Println("Replacing hi in", filePath, "resulting in", result)
		file.Contents = []byte(result)
	}
}

func main() {
	directory, _ := filepath.Abs("./")

	site := gosnap.GoSnap{
		Source:      path.Join(directory, "source"),
		Destination: path.Join(directory, "destination"),
		Clean:       false,
		Plugins:     []gosnap.Plugin{noHi},
	}

	site.Build()
}
