package app

import (
	"log"
	"net/http"

	jwtmiddleware "github.com/auth0/go-jwt-middleware"
	"github.com/form3tech-oss/jwt-go"
	"github.com/gorilla/mux"
	"github.com/urfave/negroni"
	"gorm.io/gorm"
)

type App struct {
	Router      *mux.Router
	AuthRouter  *mux.Router
	AuthNegroni *negroni.Negroni
	Negroni     *negroni.Negroni
	Middleware  *jwtmiddleware.JWTMiddleware
	DB          *gorm.DB
}

func (a *App) Init() {
	a.Router = mux.NewRouter()
	a.AuthRouter = mux.NewRouter()

	a.Middleware = jwtmiddleware.New(jwtmiddleware.Options{
		ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
			return []byte("secret"), nil
		},
		SigningMethod: jwt.SigningMethodHS256,
	})

	a.AuthNegroni = negroni.New(negroni.HandlerFunc(a.Middleware.HandlerWithNext), negroni.Wrap(a.AuthRouter))
	a.Router.PathPrefix("/api").Handler(a.AuthNegroni)

	a.setRoutes()
}

func (a *App) setRoutes() {
	a.getNoAuth("/no-auth", a.noAuth)
	a.get("/api/auth", a.auth)
}

func (a *App) getNoAuth(path string, f func(w http.ResponseWriter, r *http.Request)) {
	a.Router.HandleFunc(path, f).Methods("GET")
}

func (a *App) get(path string, f func(w http.ResponseWriter, r *http.Request)) {
	a.AuthRouter.HandleFunc(path, f).Methods("GET")
}

func (a *App) auth(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL.RawPath)
	w.Write([]byte("auth required"))
}

func (a *App) noAuth(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL.RawPath)
	w.Write([]byte("no auth required"))
}

func (a *App) Run(host string) {
	a.Negroni = negroni.Classic()
	a.Negroni.UseHandler(a.Router)
	log.Printf("Server running at %s\n", host)
	a.Negroni.Run(host)
}
