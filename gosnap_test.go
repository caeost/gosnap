package gosnap

import (
	"reflect"
	"testing"
)

func TestTransformIgnoreArrayToMap(t *testing.T) {

}

func TestTransformToLocalPath(t *testing.T) {

}

type readStruct struct {
	initial  []Plugin
	toAdd    []Plugin
	endState []Plugin
}

var readTests = []readStruct{
	{nil, nil, nil},
}

func TestRead(t *testing.T) {
	oldReadFile := ioUtilReadFile

	defer func() { ioUtilReadFile = oldReadFile }()

	ioUtilReadFile = func(path string) ([]byte, error) {
		return []byte("hi\n" + path + "\nbye\n"), nil
	}
}

func TestParseFrontmatter(t *testing.T) {

}

type readFileStruct struct {
	path              string
	content           []byte
	frontmatterValues map[interface{}]interface{}
}

var readFileTests = []readFileStruct{
	{"a/file.html", []byte("hi\na/file.html\nbye\n"), nil},
}

func TestReadFile(t *testing.T) {
	oldReadFile := ioUtilReadFile

	defer func() { ioUtilReadFile = oldReadFile }()

	ioUtilReadFile = func(path string) ([]byte, error) {
		return []byte("hi\n" + path + "\nbye\n"), nil
	}

	for _, test := range readFileTests {
		site := GoSnap{}

		file := site.ReadFile(test.path)

		if !reflect.DeepEqual(file.Content, test.content) {
			t.Error("Expected file", test.path,
				"to contain:\n", string(test.content),
				"instead got:\n", string(file.Content))
		}
		if !reflect.DeepEqual(file.Data, test.frontmatterValues) {
			t.Error("Expected file", test.path,
				"parsed frontmatter data to be", test.frontmatterValues,
				"instead got", file.Data)
		}
	}
}

func TestWrite(t *testing.T) {

}

func TestWriteFile(t *testing.T) {

}

type useStruct struct {
	initial  []Plugin
	toAdd    []Plugin
	endState []Plugin
}

func a(fm FileMapType) {

}

var useTests = []useStruct{
	{nil, nil, nil},
	{[]Plugin{a}, nil, []Plugin{a}},
	{nil, []Plugin{a}, []Plugin{a}},
}

func containsSamePlugins(a []Plugin, b []Plugin) bool {
	if len(a) != len(b) {
		return false
	} else {
		for i, plugin := range a {
			if reflect.ValueOf(plugin).Pointer() != reflect.ValueOf(b[i]).Pointer() {
				return false
			}
		}

		return true
	}
}

func TestUse(t *testing.T) {
	for _, test := range useTests {
		site := GoSnap{}

		if test.initial != nil {
			site.Plugins = test.initial
		}

		if test.toAdd != nil {
			for _, plugin := range test.toAdd {
				site.Use(plugin)
			}
		}

		if !containsSamePlugins(site.Plugins, test.endState) {
			t.Error(
				"For initial state", test.initial,
				"Adding", test.toAdd,
				"expected", test.endState,
				"got", site.Plugins,
			)
		}
	}
}

func TestBuild(t *testing.T) {

}

func TestRun(t *testing.T) {

}
