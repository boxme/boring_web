package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"lenslocked.com/controllers"
	"lenslocked.com/models"
	"net/http"
)

const (
	host   = "localhost"
	port   = 5432
	user   = "desmond"
	dbname = "lenslocked_dev"
)

func main() {

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s dbname=%s sslmode=disable", host, port, user, dbname)

	services, err := models.NewServices(psqlInfo)
	if err != nil {
		panic(err)
	}
	defer services.Close()
	services.AutoMigrate()

	staticC := controllers.NewStatic()
	usersC := controllers.NewUsers(services.User)

	r := mux.NewRouter()
	r.Handle("/", staticC.Home).Methods("GET")
	r.Handle("/contact", staticC.Contact).Methods("GET")
	r.Handle("/faq", staticC.Faq).Methods("GET")
	r.HandleFunc("/signup", usersC.New).Methods("GET")
	r.HandleFunc("/signup", usersC.Create).Methods("POST")

	// Use Handle instead of HandleFunc because LoginView is not a handler
	r.Handle("/login", usersC.LoginView).Methods("GET")
	r.HandleFunc("/login", usersC.Login).Methods("POST")

	r.HandleFunc("/cookietest", usersC.CookieTest).Methods("GET")

	var handlerFor404 http.Handler = http.HandlerFunc(unknown404)
	r.NotFoundHandler = handlerFor404
	http.ListenAndServe(":3000", r)
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
