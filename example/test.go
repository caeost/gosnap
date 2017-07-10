package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"runtime"

	"github.com/caeost/gosnap"
	"github.com/caeost/gosnap/plugins"
)

func whatKey(fileMap gosnap.FileMapType) error {
	for _, file := range fileMap {
		if file.Data["key"] != nil {
			file.Content = []byte("key: " + file.Data["key"].(string))
		}
	}

	return nil
}

func getDirectory() string {
	_, filename, _, ok := runtime.Caller(0)

	if !ok {
		fmt.Println("Could not figure out position of directory")
		os.Exit(1)
	}

	return path.Dir(filename)
}

func main() {
	directory := getDirectory()

	site := gosnap.GoSnap{
		Source:      path.Join(directory, "source"),
		Destination: path.Join(directory, "destination"),
		Clean:       true,
		Logger:      log.New(os.Stderr, "Snap: ", log.Lshortfile|log.Ldate|log.Ltime),
	}

	site.Use(whatKey)
	site.Use(plugins.Render)
	site.Use(plugins.MinifyCSS)
	site.Use(plugins.MinifyJS)

	err := site.Build()

	if err != nil {
		fmt.Printf("Error running build. %v", err)
		os.Exit(1)
	}
}
