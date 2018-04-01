package static

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
)

type mockDir struct {
	files map[string]mockFile
}

func (d *mockDir) List() []string {
	var o []string
	for k := range d.files {
		o = append(o, k)
	}
	return o
}

func (d *mockDir) Open(name string) (File, error) {
	if f, ok := d.files[name]; ok {
		return &f, nil
	}
	return nil, fmt.Errorf("%s not found", name)
}

type mockFile struct {
	*strings.Reader
}

func newMocFile(s string) mockFile {
	return mockFile{strings.NewReader(s)}
}

func TestParseInfo(t *testing.T) {
	sample := []struct {
		name string
		pass bool
	}{
		{
			"en/Introduction_To_Programming/", true,
		},
		{
			"en/Data_Structures/", true,
		},
		{
			"", false,
		},
		{
			"en/", true,
		},
		{
			"fail/", false,
		},
	}
	for _, v := range sample {
		_, err := parseInfo(v.name)
		if v.pass && err != nil {
			println(v.name)
			t.Error(err)
		}
		if !v.pass && err == nil {
			t.Error("expected an error")
		}
	}
}

func TestLoad(t *testing.T) {
	sample := []pathInfo{
		{lang: "en", chapter: "one", topic: "first", prog: "ho", page: "one.md"},
		{lang: "en", chapter: "one", topic: "first", prog: "js", page: "one.md"},
		{lang: "en", chapter: "one", topic: "first", prog: "py", page: "one.md"},
		{lang: "en", chapter: "one", path: "en/one/info.json"},
		{lang: "en", chapter: "one", path: "en/one/README.md"},
		{lang: "en", chapter: "one", topic: "first", path: "en/one/first/README.md"},
	}
	dir := &mockDir{files: make(map[string]mockFile)}
	for i, v := range sample {
		if v.path != "" {
			_, fname := filepath.Split(v.path)
			switch fname {
			case descriptionFIle:
				dir.files[v.path] = newMocFile(descriptionFIle)
			case infoFile:
				name := v.topic
				if name == "" {
					name = v.chapter
				}
				b, err := json.Marshal(Info{
					Name:  name,
					Index: i,
				})

				if err != nil {
					t.Fatal(err)
				}
				dir.files[v.path] = newMocFile(string(b))
			}
		} else {
			dir.files[v.getPath()] = newMocFile(v.getPath())
		}
	}
	_, err := LoadChapters(dir)
	if err != nil {
		t.Fatal(err)
	}
}
