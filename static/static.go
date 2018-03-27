// Package static provides utilities for rendering static content of the static
// academy website.
//
// The content is sourced from a directory which expects a strict layout. Layout
// is important since we favor convention over configuration. Content is
// organized by Language (English/Swahili) / Chapters ( These are based on the
// andela home study) / Topic / Programming language (Go/Python) .
//
// The files are loaded from disk, with support for SIGHUP which signals
// reloading of the content when it is changed.
package static

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"
	"sort"
	"strings"
	"time"

	valid "gopkg.in/asaskevich/govalidator.v9"
)

const descriptionFIle = "README.md"
const infoFile = "info"

// Language is natural language. Example for English, Short is en.
type Language struct {
	Short string `json:"short"`
	Long  string `json:"long"`
}

// ProgrammingLanguage details about the programming langgages.
type ProgrammingLanguage struct {
	Name        string `json:"name"`
	Short       string `json:"short"`
	Description string `json:"description"`
}

// Page represent the file expalining about about a chapter topic. Each page has
// a main author and collaborators.
type Page struct {
	Metadata            Metadata            `json:"meta"`
	Language            Language            `json:"language"`
	ProgrammingLanguage ProgrammingLanguage `json:"programming_language"`
	Path                string              `json:"path"`
	ModTime             time.Time           `json:"mod_time"`
	Prev                *Page               `json:"-"`
	Next                *Page               `json:"-"`
}

// Metadata additional information about the page.
type Metadata struct {
	// MainAuthor is the person who initially created the topic.
	MainAuthor Author

	// Collaborators anyone else from the main author who contributed to the topic.
	Collaborators []Author

	CreatedAt time.Time
}

// Author details about the post author.
type Author struct {
	Name    string `json:"name"`
	Twitter string `json:"twitter"`
	Github  string `json:"github"`
}

// Topic details about specific area of a chapter.
type Topic struct {
	Info        *Info
	Name        string `json:"name"`
	Description string `json:"desc"`
	Path        string `json:"path"`
}

// Chapter details about specific lesson/subject a student need to know in
// programming.
type Chapter struct {
	Info   *Info
	Topics map[string]*Topic
}

type Info struct {
	Name        string `json:"name"`
	Description string `json:"desc"`
	Path        string `json:"path"`
	Index       int    `json:"index"`
}

type Site struct {
	Context map[string]map[string]*Chapter
}

// ListChapters returns chapters for the the given language. Note that the
// language is not programming language it is the language of source material
// for example en for english..
func (s *Site) ListChapters(lang string) []*Chapter {
	if cp, ok := s.Context[lang]; ok {
		var chapters []*Chapter
		for _, v := range cp {
			chapters = append(chapters, v)
		}
		sort.SliceStable(chapters, func(i, j int) bool {
			return chapters[i].Info.Index < chapters[j].Info.Index
		})
		return chapters
	}
	return nil
}

type pathInfo struct {
	lang    string
	chapter string
	prog    string
	topic   string
	page    string
	path    string
}

var errBadLayout = errors.New("catac: bad layout expected files with path format /{language}//{chapter}/{topic}/{programming_language}/{page}.md")

func parseInfo(fileName string) (*pathInfo, error) {
	if fileName == "" {
		return nil, errors.New("empty filename")
	}
	dir, file := filepath.Split(fileName)
	parts := strings.Split(dir, string(filepath.Separator))
	info := &pathInfo{
		path: file,
	}
	switch len(parts) {
	case 1:
		info.lang = parts[0]
	case 2:
		info.lang = parts[0]
		info.chapter = parts[1]
	case 3:
		info.lang = parts[0]
		info.chapter = parts[1]
		info.topic = parts[2]
	case 4:
		info.lang = parts[0]
		info.chapter = parts[1]
		info.topic = parts[2]
		info.prog = parts[3]
	case 5:
		info.lang = parts[0]
		info.chapter = parts[1]
		info.topic = parts[2]
		info.prog = parts[3]
		info.page = parts[4]
	}
	return info, info.valid()
}

func (p pathInfo) getPath() string {
	if p.path != "" {
		return p.path
	}
	return filepath.Join(p.lang, p.chapter, p.topic, p.prog, p.page)
}

type File interface {
	io.Reader
}

// Dir is an interface for file resources which have the website sources.
type Dir interface {

	// Open returns File object with the given name.
	Open(name string) (File, error)

	// List returns unsorted list of all the files under this directory. This
	// excludes sub directories and the files are recursively listed from all the
	// subdirectories.
	List() []string
}

var errBadLanguageCode = "%s is not a valid language code"

// validates the information stored in the pathInfo.
func (i *pathInfo) valid() error {
	// validate language
	switch len(i.lang) {
	case 2:
		if !valid.IsISO693Alpha2(i.lang) {
			return fmt.Errorf(errBadLanguageCode, i.lang)
		}
	case 3:
		if !valid.IsISO693Alpha3b(i.lang) {
			return fmt.Errorf(errBadLanguageCode, i.lang)
		}
	default:
		return fmt.Errorf(errBadLanguageCode, i.lang)
	}
	return nil
}

func (i *pathInfo) topicPath() string {
	return filepath.Join(i.lang, i.chapter, i.topic)
}

func (i *pathInfo) chapterPath() string {
	return filepath.Join(i.lang, i.chapter)
}

func (i *pathInfo) sameChapter(n *pathInfo) bool {
	return i.lang == n.lang &&
		i.chapter == n.chapter
}

func (i *pathInfo) sameTopic(n *pathInfo) bool {
	return i.lang == n.lang &&
		i.chapter == n.chapter &&
		i.topic == n.topic
}

// LoadChapters traverses the dir and loads chapters based on the path layout.
// This supports chapters which have no topics and topics without pages.
//
// If there is a chapter, chapter info must be provided if it missing this will
// return an error. The same happens to topics, if there is a topic then topic
// info must be provided.
//
//  Chapter info file must be in bath {lang}/{chapter}/info.json.
//  Topic info file must be in bath {lang}/{chapter}/{topic}/info.json.
//
// For details about the content of info file see `Info``struct.
func LoadChapters(dir Dir) (map[string]map[string]*Chapter, error) {
	chapters := make(map[string]map[string]*Chapter)
	processed := make(map[string]bool)
	infoMap := make(map[string]*pathInfo)
	var i, n *pathInfo
	var ok bool
	var err error
	list := dir.List()
	for _, v := range list {
		if filepath.Base(v) == descriptionFIle {
			continue
		}
		if i, ok = infoMap[v]; !ok {
			i, err = parseInfo(v)
			if err != nil {
				return nil, err
			}
			infoMap[v] = i
		}
		if i.chapter != "" {
			var c *Chapter
			chapterPath := i.chapterPath()
			var ok bool
			if c, ok = chapters[i.lang][i.chapter]; !ok {
				info, err := readInfo(i.chapter, chapterPath, dir)
				if err != nil {
					return nil, err
				}
				c = &Chapter{
					Info:   info,
					Topics: make(map[string]*Topic),
				}
				if chapters[i.lang] == nil {
					chapters[i.lang] = make(map[string]*Chapter)
				}
				chapters[i.lang][i.chapter] = c
			}
			for _, item := range list {
				if processed[item] {
					continue
				}
				if n, ok = infoMap[item]; !ok {
					n, err = parseInfo(item)
					if err != nil {
						return nil, err
					}
					infoMap[item] = n
				}
				if n.topic != "" {
					if i.sameChapter(n) {
						var t *Topic
						if t, ok = c.Topics[i.topic]; !ok {
							t = &Topic{}
						}
						if t.Info == nil {
							topicInfo, err := readInfo(i.topic, i.topicPath(), dir)
							if err != nil {
								return nil, err
							}
							t.Info = topicInfo
						}
						processed[item] = true
					}
				}
			}
		}
	}
	return chapters, nil
}

func readInfo(name, path string, dir Dir) (*Info, error) {
	data, err := dir.Open(filepath.Join(path, descriptionFIle))
	if err != nil {
		return nil, fmt.Errorf("failed to read description  file:%s error:%v", path, err)
	}
	b, err := ioutil.ReadAll(data)
	if err != nil {
		return nil, err
	}
	c := &Info{
		Name:        name,
		Description: string(b),
		Path:        path,
	}
	return c, nil
}
