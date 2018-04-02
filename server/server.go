package server

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gernest/CatAcademy/present"
	"github.com/gernest/CatAcademy/present/models"
	"github.com/urfave/cli"
)

func Command() cli.Command {
	return cli.Command{
		Name: "serve",
		Action: func(ctx *cli.Context) error {
			return Server()
		},
	}
}

func Server() error {
	mux := http.NewServeMux()
	mux.Handle("/static/", http.StripPrefix(
		"/static/", http.FileServer(http.Dir("static")),
	))
	t, err := template.ParseGlob("templates/slides/*")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(t.DefinedTemplates())
	mux.Handle("/slide/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		Slide(t, w, r)
	}))
	name := "talks/2013/distsys.slide"
	data, err := ioutil.ReadFile(name)
	if err != nil {
		log.Fatal(err)
	}
	doc, err := present.Parse(bytes.NewReader(data), name, 0)
	if err != nil {
		log.Fatal(err)
	}
	var enc bytes.Buffer

	err = models.Encode(&enc, doc)
	if err != nil {
		log.Fatal(err)
	}
	mux.Handle("/data/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.Copy(w, &enc)
	}))
	return http.ListenAndServe(":8080", mux)
}

func Slide(t *template.Template, w http.ResponseWriter, r *http.Request) {
	err := t.ExecuteTemplate(w, "index.html", nil)
	if err != nil {
		log.Println(err)
	}
}
