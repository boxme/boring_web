package views

import (
	"bytes"
	"errors"
	"github.com/gorilla/csrf"
	"html/template"
	"io"
	"lenslocked.com/context"
	"net/http"
	"net/url"
	"path/filepath"
)

var (
	LayoutDir   string = "views/layouts/"
	TemplateDir string = "views/"
	TemplateExt string = ".gohtml"
)

type View struct {
	Template *template.Template
	Layout   string
}

func NewView(layout string, files ...string) *View {
	addTemplatePath(files)
	addTemplateExt(files)
	files = append(files, layoutFiles()...)

	// We are now changing how we create our templates, calling
	// New("") to give us a template that we can add a function to
	// before finally passing in files to parse as part of the template.
	t, err := template.New("").Funcs(template.FuncMap{
		// If this is called without being replace with a proper implementation
		// returning an error as the second argument will cause our template
		// package to return an error when executed
		"csrfField": func() (template.HTML, error) {
			return "", errors.New("csrfField is not implemented")
		},
		"pathEscape": func(s string) string {
			return url.PathEscape(s)
		},
		// Once we have our template with a function we are going to pass in files
		// to parse
	}).ParseFiles(files...)
	if err != nil {
		panic(err)
	}

	return &View{
		Template: t,
		Layout:   layout,
	}
}

func addTemplatePath(files []string) {
	for i, f := range files {
		files[i] = TemplateDir + f
	}
}

func addTemplateExt(files []string) {
	for i, f := range files {
		files[i] = f + TemplateExt
	}
}

// implements gorilla/mux http.Handler interface
func (v *View) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	v.Render(w, r, nil)
}

func (v *View) Render(w http.ResponseWriter, r *http.Request, data interface{}) {
	w.Header().Set("Content-Type", "text/html")
	var vd Data
	switch d := data.(type) {
	case Data:
		// We need to do this so we can access the data in a var
		// with the type Data
		vd = d
	default:
		// data is not of type Data. We create one and set the data
		// to the Yield field like before
		vd = Data{
			Yield: data,
		}
	}
	// Lookup and set th user to the User field
	vd.User = context.User(r.Context())
	var buf bytes.Buffer

	// We need to create the csrfField using the current http request
	csrfField := csrf.TemplateField(r)
	tpl := v.Template.Funcs(template.FuncMap{
		"csrfField": func() template.HTML {
			// We can then create this closure that returns the csrfField for
			// any templates that need access to it.
			return csrfField
		},
	})

	err := tpl.ExecuteTemplate(w, v.Layout, vd)
	if err != nil {
		http.Error(w, "Something went wrong. If the problem persists, please email us", http.StatusInternalServerError)
		return
	}
	io.Copy(w, &buf)
}

func layoutFiles() []string {
	files, err := filepath.Glob(LayoutDir + "*" + TemplateExt)
	if err != nil {
		panic(err)
	}

	return files
}
