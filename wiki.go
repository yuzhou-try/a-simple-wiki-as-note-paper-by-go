// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
        "html/template"
        "io/ioutil"
        "net/http"
        "regexp"

        // add by yuzhou
        "path/filepath"
        "os"
        "strings"
        //"fmt"
)

type Page struct {
        Title string
        Body  []byte
}

type ChildList struct {
        Children []string
}

func (p *Page) save() error {
        filename := p.Title + ".txt"
        return ioutil.WriteFile(filename, p.Body, 0600)
}

// add by yuzhou
func loadChildren() (*ChildList, error) {

        return &ChildList{Children: []string{}}, nil
}

func loadPage(title string) (*Page, error) {
        filename := title + ".txt"
        body, err := ioutil.ReadFile(filename)
        if err != nil {
                return nil, err
        }
        return &Page{Title: title, Body: body}, nil
}

func viewHandler(w http.ResponseWriter, r *http.Request, title string) {
        p, err := loadPage(title)
        if err != nil {
                http.Redirect(w, r, "/edit/"+title, http.StatusFound)
                return
        }
        renderTemplate(w, "view", p)
}

func editHandler(w http.ResponseWriter, r *http.Request, title string) {
        p, err := loadPage(title)
        if err != nil {
                p = &Page{Title: title}
        }
        renderTemplate(w, "edit", p)
}

func saveHandler(w http.ResponseWriter, r *http.Request, title string) {
        body := r.FormValue("body")
        p := &Page{Title: title, Body: []byte(body)}
        err := p.save()
        if err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
                return
        }
        http.Redirect(w, r, "/view/"+title, http.StatusFound)
}

// add by yuzhou
func indexHandler(w http.ResponseWriter, r *http.Request) {
        l, err := loadChildren()
        if err != nil {
                http.NotFound(w, r)
                return
        }

        n := 0
        err = filepath.Walk(".", func(path string, f os.FileInfo, err error) error {

                if (f == nil) {
                        return err
                }
                if f.IsDir() {
                        return nil
                }
                if (strings.Split(f.Name(), ".")[1] == "txt") {
                        l.Children = append(l.Children, strings.Split(f.Name(), ".")[0])
                        n++
                }
                return nil
        })
        err = templates.ExecuteTemplate(w, "index.html", l)
        if err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
        }
}

var templates = template.Must(template.ParseFiles("edit.html", "view.html", "index.html"))

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {
        err := templates.ExecuteTemplate(w, tmpl+".html", p)
        if err != nil {
                http.Error(w, err.Error(), http.StatusInternalServerError)
        }
}

func makeHandler(fn func(http.ResponseWriter, *http.Request, string)) http.HandlerFunc {
        return func(w http.ResponseWriter, r *http.Request) {
                fn(w, r, title)
        }
}

func main() {
        // add by yuzhou
        http.HandleFunc("/", indexHandler)
        http.HandleFunc("/index/", indexHandler)

        http.HandleFunc("/view/", makeHandler(viewHandler))
        http.HandleFunc("/edit/", makeHandler(editHandler))
        http.HandleFunc("/save/", makeHandler(saveHandler))
        http.ListenAndServe(":8080", nil)
}
