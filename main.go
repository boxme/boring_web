package main

import (
	"fmt"
	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	"lenslocked.com/controllers"
	"lenslocked.com/middleware"
	"lenslocked.com/models"
	"lenslocked.com/rand"
	"net/http"
)

func main() {
	dbCfg := DefaultPostgresConfig()
	services, err := models.NewServices(dbCfg.Dialect(), dbCfg.ConnectionInfo())
	if err != nil {
		panic(err)
	}
	defer services.Close()
	services.AutoMigrate()

	r := mux.NewRouter()

	staticC := controllers.NewStatic()
	usersC := controllers.NewUsers(services.User)
	galleriesC := controllers.NewGalleries(services.Gallery, services.Image, r)

	userMw := middleware.User{
		UserService: services.User,
	}

	requireUserMw := middleware.RequireUser{}

	r.Handle("/", staticC.Home).Methods("GET")
	r.Handle("/contact", staticC.Contact).Methods("GET")
	r.Handle("/faq", staticC.Faq).Methods("GET")
	r.HandleFunc("/signup", usersC.New).Methods("GET")
	r.HandleFunc("/signup", usersC.Create).Methods("POST")

	// Use Handle instead of HandleFunc because LoginView is not a handler
	r.Handle("/login", usersC.LoginView).Methods("GET")
	r.HandleFunc("/login", usersC.Login).Methods("POST")

	r.HandleFunc("/logout", usersC.Logout).Methods("GET")

	r.HandleFunc("/cookietest", usersC.CookieTest).Methods("GET")

	newGallery := requireUserMw.Apply(galleriesC.New)
	r.HandleFunc("/galleries/new", newGallery).Methods("GET")

	createGallery := requireUserMw.ApplyFn(galleriesC.Create)
	r.HandleFunc("/galleries", createGallery).Methods("POST")

	r.HandleFunc("/galleries", requireUserMw.ApplyFn(galleriesC.Index)).Methods("GET").Name(controllers.IndexGalleries)

	r.HandleFunc("/galleries/{id:[0-9]+}", galleriesC.Show).Methods("GET").Name(controllers.ShowGallery)
	r.HandleFunc("/galleries/{id:[0-9]+}/edit", requireUserMw.ApplyFn(galleriesC.Edit)).Methods("GET").Name(controllers.EditGallery)

	r.HandleFunc("/galleries/{id:[0-9]+}/update", requireUserMw.ApplyFn(galleriesC.Update)).Methods("POST")
	r.HandleFunc("/galleries/{id:[0-9]+}/delete", requireUserMw.ApplyFn(galleriesC.Delete)).Methods("POST")

	r.HandleFunc("/galleries/{id:[0-9]+}/images", requireUserMw.ApplyFn(galleriesC.ImageUpload)).Methods("POST")
	r.HandleFunc("/galleries/{id:[0-9]+}/images/{filename}/delete", requireUserMw.ApplyFn(galleriesC.ImageDelete)).Methods("POST")

	// Image routes
	imageHandler := http.FileServer(http.Dir("./images/"))
	r.PathPrefix("/images/").Handler(http.StripPrefix("/images/", imageHandler))

	// Assets
	assetHandler := http.FileServer(http.Dir("./assets/"))
	assetHandler = http.StripPrefix("/assets/", assetHandler)
	r.PathPrefix("/assets/").Handler(assetHandler)

	var handlerFor404 http.Handler = http.HandlerFunc(unknown404)
	r.NotFoundHandler = handlerFor404

	b, err := rand.Bytes(32)
	if err != nil {
		panic(err)
	}

	cfg := DefaultConfig()
	// Creating CSRF protection middleware
	csrfMw := csrf.Protect(b, csrf.Secure(cfg.IsProd()))

	// User middleware will run before the router routes a user to the page
	http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), csrfMw(userMw.Apply(r)))
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func unknown404(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, "404 not found here my boy")
}
