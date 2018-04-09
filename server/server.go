package server

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/gernest/vectypresent/present"
	"github.com/gernest/vectypresent/present/models"
	"github.com/urfave/cli"
)

func Command() cli.Command {
	return cli.Command{
		Name: "serve",
		Action: func(ctx *cli.Context) error {
			return Server(ctx.Args().First())
		},
	}
}

func Server(path string) error {
	mux := http.NewServeMux()
	cache := &sync.Map{}
	t, err := template.ParseGlob("templates/present/*")
	if err != nil {
		log.Fatal(err)
	}
	dirDoc, err := Load(path)
	if err != nil {
		return err
	}
	basePath := fmt.Sprintf("/%s/", filepath.Base(path))
	cache.Store("/", dirDoc)
	dirDoc.Cache(cache)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		t.ExecuteTemplate(w, "index.html", dirDoc)
	})
	mux.HandleFunc("/context", func(w http.ResponseWriter, r *http.Request) {
		WriteJson(w, dirDoc)
	})
	mux.Handle("/static/", http.StripPrefix(
		"/static/", http.FileServer(http.Dir("static")),
	))
	fileServer := http.FileServer(http.Dir(path))
	mux.Handle(basePath, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u := r.URL.Path
		if strings.HasSuffix(u, "/") {
			// It is a directory listing.
			dir := strings.TrimSuffix(u, "/")
			if doc, ok := cache.Load(dir); ok {
				t.ExecuteTemplate(w, "index.html", doc)
				return
			}
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		ext := filepath.Ext(u)
		switch ext {
		case "", ".article", ".slide":
			if doc, ok := cache.Load(u); ok {
				t.ExecuteTemplate(w, "index.html", doc)
				return
			}
		}
		http.StripPrefix(basePath, fileServer).ServeHTTP(w, r)
	}))
	mux.Handle("/files/", http.StripPrefix("/files", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u := r.URL.Path
		if v, ok := cache.Load(u); ok {
			d := v.(*models.File)
			if d.IsDir {
				http.Error(w, "bad request", http.StatusBadRequest)
				return
			}
			ext := filepath.Ext(d.Name)
			switch ext {
			case ".slide", ".article":
				f, err := os.Open(d.Path())
				if err != nil {
				}
				defer f.Close()
				dc, err := present.Parse(f, d.Path(), 0)
				if err != nil {
				}
				err = models.Encode(w, dc)
				if err != nil {
					log.Println(err)
				}
				return
			default:
				http.ServeFile(w, r, d.Path())
				return
			}
		}
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	})))
	return http.ListenAndServe(":8080", mux)
}
func cleanPath(path string) string {
	path = filepath.Clean(path)
	if filepath.IsAbs(path) {
		wd, err := os.Getwd()
		if err != nil {
			log.Fatal(err)
		}
		rel, err := filepath.Rel(wd, path)
		if err != nil {
			log.Fatal(err)
		}
		return rel
	}
	return path
}

func WriteJson(o io.Writer, v interface{}) error {
	return json.NewEncoder(o).Encode(v)
}

func Slide(t *template.Template, w http.ResponseWriter, r *http.Request) {
	err := t.ExecuteTemplate(w, "index.html", nil)
	if err != nil {
		log.Println(err)
	}
}

func LoadChildren(d *models.File) (*models.File, error) {
	o, err := ioutil.ReadDir(d.Path())
	if err != nil {
		return nil, err
	}
	for _, info := range o {
		if !strings.HasPrefix(info.Name(), ".") {
			if !info.IsDir() && !matchExt(filepath.Ext(info.Name())) {
				continue
			}
			c, err := loadIInfo(d, info)
			if err != nil {
				return nil, err
			}
			if c != nil {
				d.Children = append(d.Children, c)
			}
		}
	}
	return d, nil
}

func Load(path string) (*models.File, error) {
	stat, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	return loadIInfo(nil, stat)
}

func loadIInfo(parent *models.File, info os.FileInfo) (*models.File, error) {
	name := info.Name()
	if parent != nil {
		name = filepath.Join(parent.Name, info.Name())
	}
	child := &models.File{IsDir: info.IsDir(), Name: name}
	if child.IsDir {
		return LoadChildren(child)
	}
	return child, nil
}

func matchExt(ext string) bool {
	switch ext {
	case ".article", ".slide", ".go":
		return true
	default:
		return false
	}
}
