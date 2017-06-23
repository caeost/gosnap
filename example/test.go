package main

import (
	"fmt"
	"github.com/caeost/gosnap"
	"path"
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
	//_, executablePath, _, ok := runtime.Caller(1)

	// if !ok {
	// 	panic("error!!")
	// }

	// TODO unhardcode this
	directory := "/Users/carlostberg/go/src/github.com/caeost/gosnap/example" //path.Dir(executablePath)

	site := gosnap.GoSnap{
		Source:      path.Join(directory, "source"),
		Destination: path.Join(directory, "destination"),
		Clean:       false,
		Plugins:     []gosnap.Plugin{noHi},
	}

	site.Build()
}
