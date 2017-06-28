package plugins

import (
	"github.com/caeost/gosnap"
	"github.com/pkg/errors"
	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/css"
	"github.com/tdewolff/minify/html"
	"github.com/tdewolff/minify/js"
	"github.com/tdewolff/minify/json"
	"github.com/tdewolff/minify/svg"
	"github.com/tdewolff/minify/xml"
	"regexp"
	"strings"
)

func setup() *minify.M {
	m := minify.New()
	m.AddFunc("text/css", css.Minify)
	m.AddFunc("text/html", html.Minify)
	m.AddFunc("text/javascript", js.Minify)
	m.AddFunc("image/svg+xml", svg.Minify)
	m.AddFuncRegexp(regexp.MustCompile("[/+]json$"), json.Minify)
	m.AddFuncRegexp(regexp.MustCompile("[/+]xml$"), xml.Minify)

	return m
}

var minifier = setup()

func MinifyCSS(fileMap gosnap.FileMapType) error {
	for filePath, file := range fileMap {
		if strings.HasSuffix(filePath, ".css") {
			if val, exists := file.Data["minify"]; !exists || val == true {
				minified, err := minifier.Bytes("text/css", file.Content)

				if err != nil {
					return errors.Wrapf(err, "Could not minify file %v", filePath)
				}

				file.Content = minified
			}
		}
	}

	return nil
}
