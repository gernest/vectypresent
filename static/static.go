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
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/urfave/cli"
	valid "gopkg.in/asaskevich/govalidator.v9"
)

const descriptionFIle = "README.md"
const infoFile = "info"

// Language is natural language. Example for English, Short is en.
type Language valid.ISO693Entry

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

// OutputWriter is an interface for writing output files.
type OutputWriter interface {

	// WriteFile saves the file with filename, set data as file content and perm as
	// file permission.
	//
	// This is similar to ioutil.WriteFile except it is up to the implementation to
	// decide whenre the file lives.
	WriteFile(filename string, data []byte, perm os.FileMode) error
}

// Site is an object with all the docs site state.
type Site struct {

	// Context is a map of languages to the map of chapters.
	Context map[Language]map[string]*Chapter
}

// NewSite loads a new Site object form docsDirectory which contains all
// documentation files with the catacademy layout. exerciseDirectory contains
// all exercisesfiles in the catacademy exercise layout.
func NewSite(docsDirectory, exercisesDirectory Dir) (*Site, error) {
	ctx, err := LoadChapters(docsDirectory)
	if err != nil {
		return nil, err
	}
	return &Site{Context: ctx}, nil
}

// ListChapters returns chapters for the the given language. Note that the
// language is not programming language it is the language of source material
// for example en for english..
func (s *Site) ListChapters(lang Language) []*Chapter {
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

// Languages returns a sorted list of loaded languages.
func (s *Site) Languages() []Language {
	var langs []Language
	for k := range s.Context {
		langs = append(langs, k)
	}
	sort.SliceStable(langs, func(i, j int) bool {
		return langs[i].English < langs[j].English
	})
	return langs
}

// Template is an interface for rendering static html files.
type Template interface {

	// ExecuteTemplate renders the template matching name, using data as context.
	ExecuteTemplate(out io.Writer, name string, data interface{}) error
}

// Theme is an interface of important keys that determines how static files are
// rendered.
type Theme interface {

	// PageTpl is the name of the template used to render pages.
	PageTpl() string
	PageIndexTpl() string

	// TopicTpl is the name of the template used to render topic page.
	TopicTpl() string
	TopicIndexTpl() string

	// ChapterTpl is the name of the template used to render chapter page.
	ChapterTpl() string
	ChapterIndexTpl() string

	LanguageTpl() string
	LanguageIndexTpl() string

	// IndexTpl is the name of the template to render the site entrypoint.
	IndexTpl() string
}

// Generate generates static html pages using the tpl template. The generated
// files are handed to out for persistance.
func (s *Site) Generate(out OutputWriter, tpl Template, theme Theme) error {
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

func (i pathInfo) getPath() string {
	if i.path != "" {
		return i.path
	}
	return filepath.Join(i.lang, i.chapter, i.topic, i.prog, i.page)
}

func (i *pathInfo) language() Language {
	return getLanguage(i.lang)
}

func getLanguage(lang string) Language {
	switch len(lang) {
	case 2:
		for _, entry := range valid.ISO693List {
			if lang == entry.Alpha2Code {
				return Language(entry)
			}
		}
	case 3:
		for _, entry := range valid.ISO693List {
			if lang == entry.Alpha3bCode {
				return Language(entry)
			}
		}
	}
	return Language{}
}

// File is an interface for reading file contents.
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
func LoadChapters(dir Dir) (map[Language]map[string]*Chapter, error) {
	chapters := make(map[Language]map[string]*Chapter)
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
			if c, ok = chapters[i.language()][i.chapter]; !ok {
				info, err := readInfo(i.chapter, chapterPath, dir)
				if err != nil {
					return nil, err
				}
				c = &Chapter{
					Info:   info,
					Topics: make(map[string]*Topic),
				}
				if chapters[i.language()] == nil {
					chapters[i.language()] = make(map[string]*Chapter)
				}
				chapters[i.language()][i.chapter] = c
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

// FileDir is an implementation of Dir interface that uses the filesystem.
type FileDir struct {
	files []string
	base  string
}

// NewFileDir walks recursively through base and returns FileDir will all the
// files found.
func NewFileDir(base string) (*FileDir, error) {
	f := &FileDir{base: base}
	ferr := filepath.Walk(base, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		f.files = append(f.files, path)
		return nil
	})
	if ferr != nil {
		return nil, ferr
	}
	return f, nil
}

// List returns a list of all files present in the base directory. This is a
// recursive list of top level files and subdirectories.
func (f *FileDir) List() []string {
	return f.files
}

// Open returns a File matching the name.
func (f *FileDir) Open(name string) (File, error) {
	b, err := ioutil.ReadFile(name)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(b), nil
}

const (
	languageIndexTpl = "languages.html"
	languageTpl      = "language.html"
	chapterIndexTpl  = "chapters.html"
	chapterTpl       = "chapter.html"
	topicsIndexTpl   = "topics.html"
	topicTpl         = "topic.html"
	pagesIndexTpl    = "pages.html"
	pageTpl          = "page.html"
	homeTpl          = "index.html"
)

var _ Template = (*BaseTemplate)(nil)

// BaseTemplate implements Template interface.
type BaseTemplate struct {
	*template.Template
}

// NewBaseTemplate loads the templates from dir for the given theme.
func NewBaseTemplate(dir Dir, themeName string) (*BaseTemplate, error) {
	base := filepath.Join("themes", themeName)
	ts := filterList(dir.List(), func(s string) bool {
		return strings.HasPrefix(s, base) && filepath.Ext(s) == ".tpl"
	})
	if ts == nil {
		return nil, fmt.Errorf("theme %s not found", themeName)
	}
	t := template.New(themeName)
	for _, v := range ts {
		tpl := t.New(strings.TrimPrefix(base, v))
		d, err := dir.Open(v)
		if err != nil {
			return nil, err
		}
		b, err := ioutil.ReadAll(d)
		if err != nil {
			return nil, err
		}
		_, err = tpl.Parse(string(b))
		if err != nil {
			return nil, err
		}
	}
	return &BaseTemplate{Template: t}, nil
}

func filterList(s []string, f func(string) bool) []string {
	var out []string
	for _, v := range s {
		if f(v) {
			out = append(out, v)
		}
	}
	return out
}

type simpleTheme struct {
	languageIndexTpl string
	languageTpl      string
	chapterIndexTpl  string
	chapterTpl       string
	topicsIndexTpl   string
	topicTpl         string
	pagesIndexTpl    string
	pageTpl          string
	homeTpl          string
}

func newDefaultTheme() *simpleTheme {
	return &simpleTheme{
		languageIndexTpl: languageIndexTpl,
		languageTpl:      languageTpl,
		chapterIndexTpl:  chapterIndexTpl,
		chapterTpl:       chapterTpl,
		topicsIndexTpl:   topicsIndexTpl,
		topicTpl:         topicTpl,
		pagesIndexTpl:    pagesIndexTpl,
		pageTpl:          pageTpl,
		homeTpl:          homeTpl,
	}
}

func (s *simpleTheme) PageTpl() string      { return s.pageTpl }
func (s *simpleTheme) PageIndexTpl() string { return s.pagesIndexTpl }

// TopicTpl is the name of the template used to render topic page.
func (s *simpleTheme) TopicTpl() string      { return s.topicTpl }
func (s *simpleTheme) TopicIndexTpl() string { return s.topicsIndexTpl }

// ChapterTpl is the name of the template used to render chapter page.
func (s *simpleTheme) ChapterTpl() string      { return s.chapterTpl }
func (s *simpleTheme) ChapterIndexTpl() string { return s.chapterIndexTpl }

func (s *simpleTheme) LanguageTpl() string      { return s.languageTpl }
func (s *simpleTheme) LanguageIndexTpl() string { return s.languageIndexTpl }

// IndexTpl is the name of the template to render the site entrypoint.
func (s *simpleTheme) IndexTpl() string { return s.homeTpl }

// Command command for running docs generation.
func Command() cli.Command {
	return cli.Command{
		Name:        "docs",
		Description: "Generates static documentation website",
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:   "src",
				Usage:  "source directory",
				Value:  "docs",
				EnvVar: "CATAC_DOCS_SRC",
			},
			cli.StringFlag{
				Name:   "themeDir",
				Usage:  "directory with themes",
				Value:  "themes",
				EnvVar: "CATAC_THEMES_DIR",
			},
			cli.StringFlag{
				Name:   "theme",
				Usage:  "the name of the theme",
				Value:  "default",
				EnvVar: "CATAC_THEME_NAME",
			},
		},
	}
}

type options struct {
	src, theme, themeDir string
}

func run(ctx *cli.Context) error {
	opts := &options{
		src:      ctx.String("src"),
		theme:    ctx.String("theme"),
		themeDir: ctx.String("themeDir"),
	}
	srcDir, err := NewFileDir(opts.src)
	if err != nil {
		return err
	}
	themeDir, err := NewFileDir(opts.themeDir)
	if err != nil {
		return err
	}
	tpl, err := NewBaseTemplate(themeDir, opts.theme)
	if err != nil {
		return err
	}
	site, err := NewSite(srcDir, nil)
	if err != nil {
		return err
	}
	return site.Generate(nil, tpl, newDefaultTheme())
}
