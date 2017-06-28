package main

import (
	"fmt"
	"github.com/caeost/gosnap"
	"github.com/caeost/gosnap/plugins"
	"os"
	"path"
	"runtime"
)

func whatKey(fileMap gosnap.FileMapType) error {
	for _, file := range fileMap {
		if file.Data["key"] != nil {
			file.Content = []byte("key: " + file.Data["key"].(string))
		}
	}

	return nil
}

func main() {
	_, filename, _, ok := runtime.Caller(0)

	if !ok {
		fmt.Println("Could not figure out position of directory")
		os.Exit(1)
	}

	directory := path.Dir(filename)

	site := gosnap.GoSnap{
		Source:      path.Join(directory, "source"),
		Destination: path.Join(directory, "destination"),
	}

	site.Use(whatKey)
	site.Use(plugins.Render)
	site.Use(plugins.MinifyCSS)

	err := site.Build()

	if err != nil {
		fmt.Printf("Error running build. %v", err)
		os.Exit(1)
	}
}
