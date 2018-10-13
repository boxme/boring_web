package controllers

import (
	"fmt"
	"github.com/gorilla/mux"
	"lenslocked.com/context"
	"lenslocked.com/models"
	"lenslocked.com/views"
	"net/http"
	"strconv"
)

type Galleries struct {
	New      *views.View
	ShowView *views.View
	gs       models.GalleryService
}

type GalleryForm struct {
	Title string `schema: "title"`
}

func NewGalleries(gs models.GalleryService) *Galleries {
	return &Galleries{
		New:      views.NewView("bootstrap", "galleries/new"),
		ShowView: views.NewView("bootstrap", "galleries/show"),
		gs:       gs,
	}
}

func (g *Galleries) Create(w http.ResponseWriter, r *http.Request) {
	var vd views.Data
	var form GalleryForm
	if err := parseForm(r, &form); err != nil {
		vd.SetAlert(err)
		g.New.Render(w, vd)
		return
	}

	user := context.User(r.Context())
	gallery := models.Gallery{
		Title:  form.Title,
		UserID: user.ID,
	}

	if err := g.gs.Create(&gallery); err != nil {
		vd.SetAlert(err)
		g.New.Render(w, vd)
		return
	}
	fmt.Fprintln(w, gallery)
}

func (g *Galleries) Show(w http.ResponseWriter, r *http.Request) {
	// Get variables from path, like the "id" variable
	vars := mux.Vars(r)
	// Get the "id" variables from our vars
	idStr := vars["id"]
	// Convert idStr, which is a string, to an integer
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid gallery ID", http.StatusNotFound)
		return
	}

	// Prevent the compiler from complaining that id is not used
	_ = id

	gallery := models.Gallery{
		Title: "A temporary fake gallery with ID: " + idStr,
	}

	var vd views.Data
	vd.Yield = gallery
	g.ShowView.Render(w, vd)
}
