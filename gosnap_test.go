package gosnap

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
	"time"
)

// shortcut to get a valid FileInfo value
type MockFileInfo struct {
}

func (mfi MockFileInfo) Name() string {
	return "name"
}
func (mfi MockFileInfo) Size() int64 {
	return int64(0)
}
func (mfi MockFileInfo) Mode() os.FileMode {
	return 0777
}
func (mfi MockFileInfo) ModTime() time.Time {
	return time.Now()
}
func (mfi MockFileInfo) IsDir() bool {
	return false
}
func (mfi MockFileInfo) Sys() interface{} {
	return nil
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
		output := TransformToLocalPath(test.input, test.source)

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
	expectedError     error
}

var readFileTests = []readFileStruct{
	{"an/empty/file.html", []byte(""), []byte(""), nil, nil},
	{"a/file.html", []byte("hi\na/file.html\nbye\n"), []byte("hi\na/file.html\nbye\n"), nil, nil},
	{"bile.html", []byte("---\nkey: value\n---\nblahblah"), []byte("blahblah"), FrontmatterValueType{"key": "value"}, nil},
	{"bile-arrays.html", []byte("---\nkey: [value]\n---\nblah\nblah"), []byte("blah\nblah"), FrontmatterValueType{"key": []interface{}{"value"}}, nil},
	{"bile-arrays-other-format.html", []byte("---\nkey:\n - value\n---\nblah\nblah"), []byte("blah\nblah"), FrontmatterValueType{"key": []interface{}{"value"}}, nil},
	{"bile-maps.html", []byte("---\nkey:\n inner: value\n---\nblah\nblah"), []byte("blah\nblah"), FrontmatterValueType{"key": FrontmatterValueType{"inner": "value"}}, nil},
	{"bile-array-maps.html", []byte("---\nkey:\n - inner: value\n---\nblah\nblah"), []byte("blah\nblah"), FrontmatterValueType{"key": []interface{}{FrontmatterValueType{"inner": "value"}}}, nil},
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
		site := GoSnap{Logger: log.New(ioutil.Discard, "discard", log.Ldate)}

		file, err := site.ReadFile(test.path)

		if err != nil {
			if test.expectedError == nil {
				t.Error("ReadFile errored unexpectedly: %v", err)
			} else {
				if err != test.expectedError {
					t.Error(
						"Expected error", test.expectedError,
						"instead got", err,
					)
				}
			}
		} else {
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
}

type readStruct struct {
	directoryState []string
	fileMap        FileMapType
	expectedError  error
}

var readTests = []readStruct{
	{nil, nil, nil},
	{[]string{"file.file"}, FileMapType{"file.file": &GoSnapFile{Content: []byte("hi\nfile.file\nbye\n")}}, nil},
	{[]string{"file.file", "a/d/e/e/p/l/y/n/e/s/t/e/d/file.go"}, FileMapType{
		"file.file":                         &GoSnapFile{Content: []byte("hi\nfile.file\nbye\n")},
		"a/d/e/e/p/l/y/n/e/s/t/e/d/file.go": &GoSnapFile{Content: []byte("hi\na/d/e/e/p/l/y/n/e/s/t/e/d/file.go\nbye\n")},
	}, nil},
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
	}, nil},
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
			_ = visitor(path, MockFileInfo{}, nil)
		}

		return nil
	}

	for i, test := range readTests {
		index = i

		site := GoSnap{Logger: log.New(ioutil.Discard, "discard", log.Ldate),
			Source: "dir",
		}

		err := site.Read()

		if err != nil {
			if test.expectedError == nil {
				t.Error("Read errored unexpectedly: %v", err)
			} else {
				if err != test.expectedError {
					t.Error(
						"Expected error", test.expectedError,
						"instead got", err,
					)
				}
			}
		}

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

type writeFileStruct struct {
	path        string
	destination string
	file        GoSnapFile
	expected    writeResultStruct
}

type writeResultStruct struct {
	path    string
	content []byte
	perm    os.FileMode
}

var writeFileTests = []writeFileStruct{
	{
		"a/b.go",
		"/aa",
		GoSnapFile{Content: []byte("howdy"), FileInfo: MockFileInfo{}},
		writeResultStruct{path: "/aa/a/b.go", content: []byte("howdy"), perm: 0777},
	},
	{
		"b.go",
		"/",
		GoSnapFile{Content: []byte("howdy"), FileInfo: MockFileInfo{}},
		writeResultStruct{path: "/b.go", content: []byte("howdy"), perm: 0777},
	},
	{
		"c/c/c/c/b.go",
		"/c/c/c/",
		GoSnapFile{Content: []byte("howdy")},
		writeResultStruct{path: "/c/c/c/c/c/c/c/b.go", content: []byte("howdy"), perm: 0644},
	},
}

func TestWriteFile(t *testing.T) {
	oldIoUtilWriteFile := ioUtilWriteFile
	oldMkdirAll := mkdirAll

	defer func() { ioUtilWriteFile = oldIoUtilWriteFile }()
	defer func() { mkdirAll = oldMkdirAll }()

	results := make([]writeResultStruct, len(writeFileTests))

	index := 0
	ioUtilWriteFile = func(path string, content []byte, perm os.FileMode) error {
		results[index] = writeResultStruct{path: path, content: content, perm: perm}
		return nil
	}
	mkdirAll = func(path string, perm os.FileMode) error {
		return nil
	}

	for i, test := range writeFileTests {
		index = i

		site := GoSnap{Logger: log.New(ioutil.Discard, "discard", log.Ldate),
			Destination: test.destination,
		}

		site.WriteFile(test.path, test.file)

		if test.expected.path != results[i].path {
			t.Error(
				"Expected write to path", test.expected.path,
				"in case", i,
				"Instead got", results[i].path,
			)
		}
		if !reflect.DeepEqual(test.expected.content, results[i].content) {
			t.Error(
				"Expected to write", test.expected.content,
				"in case", i,
				"Instead wrote", results[i].content,
			)
		}
		if test.expected.perm != results[i].perm {
			t.Error(
				"Expected write with permissions", test.expected.perm,
				"in case", i,
				"Instead got", results[i].perm,
			)
		}
	}
}

type writeStruct struct {
	fileMap       FileMapType
	expected      []string
	expectedError error
}

var writeTests = []writeStruct{
	{
		FileMapType{
			"file.file":                         &GoSnapFile{Content: []byte("hi\nfile.file\nbye\n")},
			"a/d/e/e/p/l/y/n/e/s/t/e/d/file.go": &GoSnapFile{Content: []byte("hi\na/d/e/e/p/l/y/n/e/s/t/e/d/file.go\nbye\n")},
		},
		[]string{
			"/out/file.file",
			"/out/a/d/e/e/p/l/y/n/e/s/t/e/d/file.go",
		},
		nil,
	},
	{
		FileMapType{
			"file.file": &GoSnapFile{Content: []byte("hi\nfile.file\nbye\n")},
		},
		[]string{
			"/out/file.file",
		},
		nil,
	},
}

func TestWrite(t *testing.T) {
	oldIoUtilWriteFile := ioUtilWriteFile
	oldMkdirAll := mkdirAll

	defer func() { ioUtilWriteFile = oldIoUtilWriteFile }()
	defer func() { mkdirAll = oldMkdirAll }()

	index := 0
	results := make([][]string, len(writeFileTests))

	ioUtilWriteFile = func(path string, content []byte, perm os.FileMode) error {
		if results[index] != nil {
			results[index] = append(results[index], path)
		} else {
			results[index] = []string{path}
		}
		return nil
	}
	mkdirAll = func(path string, perm os.FileMode) error {
		return nil
	}

	for i, test := range writeTests {
		index = i

		site := GoSnap{Logger: log.New(ioutil.Discard, "discard", log.Ldate),
			FileMap:     test.fileMap,
			Destination: "/out",
		}

		err := site.Write()

		if err != nil {
			if test.expectedError == nil {
				t.Error("Write errored unexpectedly: %v", err)
			} else {
				if err != test.expectedError {
					t.Error(
						"Expected error", test.expectedError,
						"instead got", err,
					)
				}
			}
		}

		sort.Strings(test.expected)

		sort.Strings(results[i])

		if !reflect.DeepEqual(test.expected, results[i]) {
			t.Error(
				"Expected to write out", test.expected,
				"In case", i,
				"Instead got", results[i],
			)
		}
	}
}

type useStruct struct {
	initial  []Plugin
	toAdd    []Plugin
	endState []Plugin
}

func a(fm FileMapType) error {
	return nil
}

func b(fm FileMapType) error {
	return nil
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
		site := GoSnap{Logger: log.New(ioutil.Discard, "discard", log.Ldate)}

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

type runStruct struct {
	fileMap  FileMapType
	plugins  []Plugin
	expected FileMapType
}

var runTests = []runStruct{
	{nil, nil, nil},
	{
		FileMapType{
			"file.file":                         &GoSnapFile{Content: []byte("hi\nfile.file\nbye\n")},
			"a/d/e/e/p/l/y/n/e/s/t/e/d/file.go": &GoSnapFile{Content: []byte("hi\na/d/e/e/p/l/y/n/e/s/t/e/d/file.go\nbye\n")},
		},
		nil,
		FileMapType{
			"file.file":                         &GoSnapFile{Content: []byte("hi\nfile.file\nbye\n")},
			"a/d/e/e/p/l/y/n/e/s/t/e/d/file.go": &GoSnapFile{Content: []byte("hi\na/d/e/e/p/l/y/n/e/s/t/e/d/file.go\nbye\n")},
		},
	},
	{
		FileMapType{
			"file.file":                         &GoSnapFile{Content: []byte("hi\nfile.file\nbye\n")},
			"a/d/e/e/p/l/y/n/e/s/t/e/d/file.go": &GoSnapFile{Content: []byte("hi\na/d/e/e/p/l/y/n/e/s/t/e/d/file.go\nbye\n")},
		},
		[]Plugin{a},
		FileMapType{
			"file.file":                         &GoSnapFile{Content: []byte("hi\nfile.file\nbye\n")},
			"a/d/e/e/p/l/y/n/e/s/t/e/d/file.go": &GoSnapFile{Content: []byte("hi\na/d/e/e/p/l/y/n/e/s/t/e/d/file.go\nbye\n")},
		},
	},
}

func TestRun(t *testing.T) {
	for i, test := range runTests {
		Run(test.fileMap, test.plugins)

		if !reflect.DeepEqual(test.fileMap, test.expected) {
			t.Error(
				"Expected", test.expected,
				"Instead got", test.fileMap,
				"using plugins", test.plugins,
				"in case", i,
			)
		}
	}
}

type buildStruct struct {
	directoryState []string
	plugins        []Plugin
	expected       FileMapType
	expectedError  error
}

var buildTests = []buildStruct{
	{nil, nil, FileMapType{}, nil},
	{[]string{}, nil, FileMapType{}, nil},
	{[]string{"file.file", "a/d/e/e/p/l/y/n/e/s/t/e/d/file.go"}, nil, FileMapType{
		"file.file":                         &GoSnapFile{Content: []byte("hi\nfile.file\nbye\n")},
		"a/d/e/e/p/l/y/n/e/s/t/e/d/file.go": &GoSnapFile{Content: []byte("hi\na/d/e/e/p/l/y/n/e/s/t/e/d/file.go\nbye\n")},
	}, nil},
}

func testEqualFileMap(a FileMapType, b FileMapType) bool {
	if len(a) != len(b) {
		return false
	} else {
		for name, fp := range a {
			if b[name] == nil || !reflect.DeepEqual(fp.Content, b[name].Content) {
				return false
			}
		}

		return true
	}
}

func TestBuild(t *testing.T) {
	oldIoUtilReadFile := ioUtilReadFile
	defer func() { ioUtilReadFile = oldIoUtilReadFile }()
	ioUtilReadFile = func(path string) ([]byte, error) {
		return []byte("hi\n" + path + "\nbye\n"), nil
	}

	oldFilepathWalk := filepathWalk
	defer func() { filepathWalk = oldFilepathWalk }()
	index := 0
	filepathWalk = func(dir string, visitor filepath.WalkFunc) error {
		for _, path := range buildTests[index].directoryState {
			_ = visitor(path, MockFileInfo{}, nil)
		}

		return nil
	}

	oldIoUtilWriteFile := ioUtilWriteFile
	defer func() { ioUtilWriteFile = oldIoUtilWriteFile }()
	ioUtilWriteFile = func(path string, content []byte, perm os.FileMode) error {
		return nil
	}

	oldMkdirAll := mkdirAll
	defer func() { mkdirAll = oldMkdirAll }()
	mkdirAll = func(path string, perm os.FileMode) error {
		return nil
	}

	for i, test := range buildTests {
		index = i

		site := GoSnap{Logger: log.New(ioutil.Discard, "discard", log.Ldate),
			Source:      "/dir",
			Destination: "/out",
			Plugins:     test.plugins,
		}

		err := site.Build()

		if err != nil {
			if test.expectedError == nil {
				t.Error("Write errored unexpectedly: %v", err)
			} else {
				if err != test.expectedError {
					t.Error(
						"Expected error", test.expectedError,
						"instead got", err,
					)
				}
			}
		}

		if !testEqualFileMap(site.FileMap, test.expected) {
			t.Error(
				"Expected", test.expected,
				"Instead got", site.FileMap,
				"using plugins", test.plugins,
				"in case", i,
			)
		}
	}
}

func TestIgnore(t *testing.T) {

}
