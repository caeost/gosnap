package plugins

import (
	"github.com/caeost/gosnap"
	"github.com/pkg/errors"
	"text/template"
)

func Render(fileMap gosnap.FileMapType) error {
	for filePath, file := range fileMap {
		if file.Data["template"] == true {
			tem, err := template.New(filePath).Parse(string(file.Content))

			if err != nil {
				return errors.Wrapf(err, "Could not parse template in %v", filePath)
			}

			// clear out Content since it is the template and no longer necessary
			file.Content = []byte{}
			err = tem.ExecuteTemplate(file, tem.Name(), file.Data)

			if err != nil {
				return errors.Wrapf(err, "Could not render template in %v", filePath)
			}
		}
	}

	return nil
}
