package gosnap

import (
	"io/ioutil"
	"os"
	"path"

	"github.com/pkg/errors"
)

// default permissions for generated files
// user: read & write
// group: read
// other: read
const (
	DEFAULT_PERM = os.FileMode(0644)
)

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
