package gosnap

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
)

func TestTransformIgnoreArrayToMap(t *testing.T) {

}

type transformToLocalPathStruct struct {
	input  string
	source string
	output string
}

var transformToLocalPathTests = []transformToLocalPathStruct{
	{"/a/b/c/d", "/a/b/c", "d"},
	{"/d", "", "d"},
	{"/d", "/", "d"},
	{"a/b//d", "a", "b/d"},
}

func TestTransformToLocalPath(t *testing.T) {
	for i, test := range transformToLocalPathTests {
		output := transformToLocalPath(test.input, test.source)

		if output != test.output {
			t.Error(
				"Expected output", test.output,
				"In case", i,
				"instead got", output,
			)
		}
	}
}

type readFileStruct struct {
	path              string
	content           []byte
	expectedContent   []byte
	frontmatterValues FrontmatterValueType
}

var readFileTests = []readFileStruct{
	{"an/empty/file.html", []byte(""), []byte(""), nil},
	{"a/file.html", []byte("hi\na/file.html\nbye\n"), []byte("hi\na/file.html\nbye\n"), nil},
	{"bile.html", []byte("---\nkey: value\n---\nblahblah"), []byte("blahblah"), FrontmatterValueType{"key": "value"}},
	{"bile-arrays.html", []byte("---\nkey: [value]\n---\nblah\nblah"), []byte("blah\nblah"), FrontmatterValueType{"key": []interface{}{"value"}}},
	{"bile-arrays-other-format.html", []byte("---\nkey:\n - value\n---\nblah\nblah"), []byte("blah\nblah"), FrontmatterValueType{"key": []interface{}{"value"}}},
	{"bile-maps.html", []byte("---\nkey:\n inner: value\n---\nblah\nblah"), []byte("blah\nblah"), FrontmatterValueType{"key": FrontmatterValueType{"inner": "value"}}},
	{"bile-array-maps.html", []byte("---\nkey:\n - inner: value\n---\nblah\nblah"), []byte("blah\nblah"), FrontmatterValueType{"key": []interface{}{FrontmatterValueType{"inner": "value"}}}},
}

func TestReadFile(t *testing.T) {
	oldReadFile := ioUtilReadFile

	defer func() { ioUtilReadFile = oldReadFile }()

	index := 0
	ioUtilReadFile = func(path string) ([]byte, error) {
		return readFileTests[index].content, nil
	}

	for i, test := range readFileTests {
		index = i
		site := GoSnap{}

		file := site.ReadFile(test.path)

		if !reflect.DeepEqual(file.Content, test.expectedContent) {
			t.Error(
				"Expected file", test.path,
				"to contain:\n", string(test.content),
				"instead got:\n", string(file.Content),
			)
		}
		if !reflect.DeepEqual(file.Data, test.frontmatterValues) {
			t.Error(
				"Expected file", test.path,
				"in case", i,
				"parsed frontmatter data to be", test.frontmatterValues,
				"instead got", file.Data,
			)
		}
	}
}

type readStruct struct {
	directoryState []string
	fileMap        FileMapType
}

var readTests = []readStruct{
	{nil, nil},
	{[]string{"file.file"}, FileMapType{"file.file": &GoSnapFile{Content: []byte("hi\nfile.file\nbye\n")}}},
	{[]string{"file.file", "a/d/e/e/p/l/y/n/e/s/t/e/d/file.go"}, FileMapType{
		"file.file":                         &GoSnapFile{Content: []byte("hi\nfile.file\nbye\n")},
		"a/d/e/e/p/l/y/n/e/s/t/e/d/file.go": &GoSnapFile{Content: []byte("hi\na/d/e/e/p/l/y/n/e/s/t/e/d/file.go\nbye\n")},
	}},
	{[]string{
		"file.file",
		"a/d/e/e/p/l/y/n/e/s/t/e/d/file.go",
		"a/d/e/e/p/l/y/n/e/t/e/d/file1.go",
		"a/d/e/e/p/l/y/n/e/s/e/d/file2.go",
		"a/d/e/e/p/l/y/n/e/s/t/d/file3.go",
		"a/d/e/e/p/l/y/n/s/t/e/d/file4.go",
		"a/d/e/e/p/l/y/e/s/t/e/d/file5.go",
		"a/d/e/e/p/l/y/n/e/s/t/e/d/file6.go",
		"a/d/e/e/p/l/y/n/e/s/t/e/d/file7.go",
		"a/d/e/e/p/l/y/s/t/e/d/file8.go",
		"a/d/e/e/p/n/e/s/t/e/d/file9.go",
		"a/d/e/e/p/l/y/n/e/s/t/e/d/file10.go",
		"a/d/e/e/p/l/y/n/e/s/t/e/file11.go",
		"a/d/e/e/p/l/y/n/e/s/t/d/file12.go",
		"a/d/e/e/p/l/y/n/e/s/e/d/file13.go",
		"a/d/e/e/p/l/y/n/e/t/e/d/file14.go",
		"a/d/e/e/p/l/y/n/s/t/e/d/file15.go",
		"a/d/e/e/p/l/y/e/s/t/e/d/file16.go",
		"/a/d/e/e/p/l/n/e/s/t/e/d/file17.go",
		"/a/d/e/e/p/y/n/e/s/t/e/d/file18.go",
		"a/d/e/e/l/y/n/e/s/t/e/d/file19.go",
		"a/d/e/p/l/y/n/e/s/t/e/d/file20.go",
		"a/d/e/p/l/y/n/e/s/t/e/d/file21.go",
		"a/e/e/p/l/y/n/e/s/t/e/d/file22.go",
		"/d/e/e/p/l/y/n/e/s/t/e/d/file23.go",
		"/a/d/e/e/p/l/y/n/e/s/t/e/d/file24.go",
		"/a/d/e/e/l/y/n/e/s/t/e/d/file25.go",
		"/a/d/e/p/l/y/n/e/s/t/e/d/file26.go",
	}, FileMapType{
		"file.file":                           &GoSnapFile{Content: []byte("hi\nfile.file\nbye\n")},
		"a/d/e/e/p/l/y/n/e/s/t/e/d/file.go":   &GoSnapFile{Content: []byte("hi\na/d/e/e/p/l/y/n/e/s/t/e/d/file.go\nbye\n")},
		"a/d/e/e/p/l/y/n/e/t/e/d/file1.go":    &GoSnapFile{Content: []byte("hi\na/d/e/e/p/l/y/n/e/t/e/d/file1.go\nbye\n")},
		"a/d/e/e/p/l/y/n/e/s/e/d/file2.go":    &GoSnapFile{Content: []byte("hi\na/d/e/e/p/l/y/n/e/s/e/d/file2.go\nbye\n")},
		"a/d/e/e/p/l/y/n/e/s/t/d/file3.go":    &GoSnapFile{Content: []byte("hi\na/d/e/e/p/l/y/n/e/s/t/d/file3.go\nbye\n")},
		"a/d/e/e/p/l/y/n/s/t/e/d/file4.go":    &GoSnapFile{Content: []byte("hi\na/d/e/e/p/l/y/n/s/t/e/d/file4.go\nbye\n")},
		"a/d/e/e/p/l/y/e/s/t/e/d/file5.go":    &GoSnapFile{Content: []byte("hi\na/d/e/e/p/l/y/e/s/t/e/d/file5.go\nbye\n")},
		"a/d/e/e/p/l/y/n/e/s/t/e/d/file6.go":  &GoSnapFile{Content: []byte("hi\na/d/e/e/p/l/y/n/e/s/t/e/d/file6.go\nbye\n")},
		"a/d/e/e/p/l/y/n/e/s/t/e/d/file7.go":  &GoSnapFile{Content: []byte("hi\na/d/e/e/p/l/y/n/e/s/t/e/d/file7.go\nbye\n")},
		"a/d/e/e/p/l/y/s/t/e/d/file8.go":      &GoSnapFile{Content: []byte("hi\na/d/e/e/p/l/y/s/t/e/d/file8.go\nbye\n")},
		"a/d/e/e/p/n/e/s/t/e/d/file9.go":      &GoSnapFile{Content: []byte("hi\na/d/e/e/p/n/e/s/t/e/d/file9.go\nbye\n")},
		"a/d/e/e/p/l/y/n/e/s/t/e/d/file10.go": &GoSnapFile{Content: []byte("hi\na/d/e/e/p/l/y/n/e/s/t/e/d/file10.go\nbye\n")},
		"a/d/e/e/p/l/y/n/e/s/t/e/file11.go":   &GoSnapFile{Content: []byte("hi\na/d/e/e/p/l/y/n/e/s/t/e/file11.go\nbye\n")},
		"a/d/e/e/p/l/y/n/e/s/t/d/file12.go":   &GoSnapFile{Content: []byte("hi\na/d/e/e/p/l/y/n/e/s/t/d/file12.go\nbye\n")},
		"a/d/e/e/p/l/y/n/e/s/e/d/file13.go":   &GoSnapFile{Content: []byte("hi\na/d/e/e/p/l/y/n/e/s/e/d/file13.go\nbye\n")},
		"a/d/e/e/p/l/y/n/e/t/e/d/file14.go":   &GoSnapFile{Content: []byte("hi\na/d/e/e/p/l/y/n/e/t/e/d/file14.go\nbye\n")},
		"a/d/e/e/p/l/y/n/s/t/e/d/file15.go":   &GoSnapFile{Content: []byte("hi\na/d/e/e/p/l/y/n/s/t/e/d/file15.go\nbye\n")},
		"a/d/e/e/p/l/y/e/s/t/e/d/file16.go":   &GoSnapFile{Content: []byte("hi\na/d/e/e/p/l/y/e/s/t/e/d/file16.go\nbye\n")},
		"a/d/e/e/p/l/n/e/s/t/e/d/file17.go":   &GoSnapFile{Content: []byte("hi\n/a/d/e/e/p/l/n/e/s/t/e/d/file17.go\nbye\n")},
		"a/d/e/e/p/y/n/e/s/t/e/d/file18.go":   &GoSnapFile{Content: []byte("hi\n/a/d/e/e/p/y/n/e/s/t/e/d/file18.go\nbye\n")},
		"a/d/e/e/l/y/n/e/s/t/e/d/file19.go":   &GoSnapFile{Content: []byte("hi\na/d/e/e/l/y/n/e/s/t/e/d/file19.go\nbye\n")},
		"a/d/e/p/l/y/n/e/s/t/e/d/file20.go":   &GoSnapFile{Content: []byte("hi\na/d/e/p/l/y/n/e/s/t/e/d/file20.go\nbye\n")},
		"a/d/e/p/l/y/n/e/s/t/e/d/file21.go":   &GoSnapFile{Content: []byte("hi\na/d/e/p/l/y/n/e/s/t/e/d/file21.go\nbye\n")},
		"a/e/e/p/l/y/n/e/s/t/e/d/file22.go":   &GoSnapFile{Content: []byte("hi\na/e/e/p/l/y/n/e/s/t/e/d/file22.go\nbye\n")},
		"d/e/e/p/l/y/n/e/s/t/e/d/file23.go":   &GoSnapFile{Content: []byte("hi\n/d/e/e/p/l/y/n/e/s/t/e/d/file23.go\nbye\n")},
		"a/d/e/e/p/l/y/n/e/s/t/e/d/file24.go": &GoSnapFile{Content: []byte("hi\n/a/d/e/e/p/l/y/n/e/s/t/e/d/file24.go\nbye\n")},
		"a/d/e/e/l/y/n/e/s/t/e/d/file25.go":   &GoSnapFile{Content: []byte("hi\n/a/d/e/e/l/y/n/e/s/t/e/d/file25.go\nbye\n")},
		"a/d/e/p/l/y/n/e/s/t/e/d/file26.go":   &GoSnapFile{Content: []byte("hi\n/a/d/e/p/l/y/n/e/s/t/e/d/file26.go\nbye\n")},
	}},
}

func mapKeys(mymap FileMapType) []string {
	keys := make([]string, len(mymap))

	i := 0
	for k := range mymap {
		keys[i] = k
		i++
	}

	return keys
}

// shortcut to get a valid FileInfo value
var testFileInfo, _ = os.Lstat("./gosnap_test.go")

func TestRead(t *testing.T) {
	oldIoUtilReadFile := ioUtilReadFile
	oldFilepathWalk := filepathWalk

	defer func() { ioUtilReadFile = oldIoUtilReadFile }()
	defer func() { filepathWalk = oldFilepathWalk }()

	index := 0
	ioUtilReadFile = func(path string) ([]byte, error) {
		return []byte("hi\n" + path + "\nbye\n"), nil
	}
	filepathWalk = func(dir string, visitor filepath.WalkFunc) error {
		for _, path := range readTests[index].directoryState {
			_ = visitor(path, testFileInfo, nil)
		}

		return nil
	}

	for i, test := range readTests {
		index = i

		site := GoSnap{}

		site.Read()

		sitePaths := mapKeys(site.FileMap)
		sort.Strings(sitePaths)

		testPaths := mapKeys(test.fileMap)
		sort.Strings(testPaths)

		if !reflect.DeepEqual(sitePaths, testPaths) {
			t.Error(
				"Expected read to find", testPaths,
				"in case", i,
				"Instead got", sitePaths,
			)
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

func b(fm FileMapType) {

}

var useTests = []useStruct{
	{nil, nil, nil},
	{[]Plugin{a}, nil, []Plugin{a}},
	{nil, []Plugin{a}, []Plugin{a}},
	{[]Plugin{b}, []Plugin{a}, []Plugin{b, a}},
	{[]Plugin{a}, []Plugin{b}, []Plugin{a, b}},
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
	for i, test := range useTests {
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
				"in case", i,
			)
		}
	}
}

func TestBuild(t *testing.T) {

}

func TestRun(t *testing.T) {

}
