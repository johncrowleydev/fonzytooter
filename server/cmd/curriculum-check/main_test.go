package main

import (
	"bytes"
	"testing"
	"testing/fstest"

	"github.com/johncrowleydev/fonzytooter/server/internal/curriculum"
)

func TestWriteUnusedSourceWarningsIsDeterministic(t *testing.T) {
	catalog, err := curriculum.Load(fstest.MapFS{
		"sources.yaml":                              &fstest.MapFile{Data: []byte("sources:\n  z-unused:\n    title: Z\n    url: https://example.com/z\n  a-unused:\n    title: A\n    url: https://example.com/a\n")},
		"courses/course/course.yaml":                &fstest.MapFile{Data: []byte("id: course\ntitle: Course\ndescription: Course.\norder: 0\n")},
		"courses/course/modules/module/module.yaml": &fstest.MapFile{Data: []byte("id: module\ntitle: Module\norder: 0\nobjectives: []\nvideos: []\nlessons: []\n")},
	})
	if err != nil {
		t.Fatalf("load curriculum: %v", err)
	}

	var output bytes.Buffer
	writeUnusedSourceWarnings(&output, catalog)
	if got, want := output.String(), "warning: unused source id \"a-unused\"\nwarning: unused source id \"z-unused\"\n"; got != want {
		t.Fatalf("warnings = %q, want %q", got, want)
	}
}
