package app

import (
	"fmt"
	"net/http"
	"os"

	"forum-server/app/handler"
	"forum-server/audit"
	db "forum-server/db"

	jwtmiddleware "github.com/auth0/go-jwt-middleware"
	"github.com/form3tech-oss/jwt-go"
	"github.com/gorilla/mux"
	_ "github.com/joho/godotenv/autoload"
	"github.com/rs/cors"
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
	CORS        *cors.Cors
	Auditor     *audit.Auditor
}

func (a *App) Init(auditor *audit.Auditor) {
	a.Auditor = auditor
	a.DB = db.Init(a.Auditor)
	a.Router = mux.NewRouter()
	a.AuthRouter = mux.NewRouter()

	a.setMiddleware()
	a.setRoutes()

	a.AuthNegroni = negroni.New(negroni.HandlerFunc(a.Middleware.HandlerWithNext), negroni.Wrap(a.AuthRouter))
	a.Router.PathPrefix("/api").Handler(a.AuthNegroni)

	a.CORS = cors.New(cors.Options{
		AllowedOrigins:   []string{"https://foo.com", "https://www.foo.com", "http://localhost:8080"},
		AllowedMethods:   []string{"HEAD", "GET", "POST", "PUT", "DELETE"},
		AllowedHeaders:   []string{"*"},
		AllowCredentials: true,
		Debug:            true,
	})
}

func (a *App) setMiddleware() {
	a.Middleware = jwtmiddleware.New(jwtmiddleware.Options{
		ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
			return []byte(os.Getenv("JWT_SECRET")), nil
		},
		SigningMethod: jwt.SigningMethodHS256,
	})
}

func (a *App) setRoutes() {
	a.getNoAuth("/api/no-auth", a.noAuth)
	a.get("/api/auth", a.auth)

	// TODO: find better names for routes

	a.postNoAuth("/api/login", a.login)
	a.postNoAuth("/api/register", a.register)
	a.get("/api/users", a.getUsers)
	a.get("/api/user/{userId}", a.getUserById)
	a.put("/api/user/{userId}", a.updateUser)
	a.delete("/api/user/{userId}", a.deleteUser)

	a.get("/api/boards", a.getBoards)
	a.get("/api/board", a.getBoard)
	a.get("/api/posts/{postId}", a.getPostsFromBoard)
	a.get("/api/posts/{postId}/comments", a.getCommentsFromPost)
	a.post("/api/boards/addBoard", a.addBoard)
	a.post("/api/boards/{boardId}/addPost", a.addPost)
	a.post("/api/boards/{postId}/addComment", a.addComment)
	a.put("/api/posts/{postId}", a.updatePost)
	a.put("/api/posts/comments/{commentId}", a.updateComment)
	a.delete("/api/posts/{postId}", a.deletePost)
	a.delete("/api/posts/comments{commentId}", a.deleteComment)
}

func (a *App) getNoAuth(path string, f func(w http.ResponseWriter, r *http.Request)) {
	a.Router.HandleFunc(path, f).Methods("GET")
}

func (a *App) postNoAuth(path string, f func(w http.ResponseWriter, r *http.Request)) {
	a.Router.HandleFunc(path, f).Methods("POST")
}

func (a *App) get(path string, f func(w http.ResponseWriter, r *http.Request)) {
	a.AuthRouter.HandleFunc(path, f).Methods("GET")
}

func (a *App) post(path string, f func(w http.ResponseWriter, r *http.Request)) {
	a.AuthRouter.HandleFunc(path, f).Methods("POST")
}

func (a *App) put(path string, f func(w http.ResponseWriter, r *http.Request)) {
	a.AuthRouter.HandleFunc(path, f).Methods("PUT")
}

func (a *App) delete(path string, f func(w http.ResponseWriter, r *http.Request)) {
	a.AuthRouter.HandleFunc(path, f).Methods("DELETE")
}

func (a *App) auth(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	user := ctx.Value("user")
	for k, v := range user.(*jwt.Token).Claims.(jwt.MapClaims) {
		fmt.Fprintf(w, "%s :\t%#v\n", k, v)
	}
}

func (a *App) noAuth(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("no auth required"))
}

func (a *App) login(w http.ResponseWriter, r *http.Request) {
	handler.UserLogin(a.DB, a.Auditor, w, r)
}

func (a *App) register(w http.ResponseWriter, r *http.Request) {
	handler.UserRegister(a.DB, a.Auditor, w, r)
}

func (a *App) getUserById(w http.ResponseWriter, r *http.Request) {
	handler.GetUserById(a.DB, a.Auditor, w, r)
}

func (a *App) getUsers(w http.ResponseWriter, r *http.Request) {
	handler.GetUsers(a.DB, a.Auditor, w, r)
}

func (a *App) updateUser(w http.ResponseWriter, r *http.Request) {
	handler.UpdateUser(a.DB, a.Auditor, w, r)
}

func (a *App) deleteUser(w http.ResponseWriter, r *http.Request) {
	handler.DeleteUser(a.DB, a.Auditor, w, r)
}

func (a *App) getBoards(w http.ResponseWriter, r *http.Request) {

}

func (a *App) getBoard(w http.ResponseWriter, r *http.Request) {

}

func (a *App) getPostsFromBoard(w http.ResponseWriter, r *http.Request) {

}

func (a *App) getCommentsFromPost(w http.ResponseWriter, r *http.Request) {

}

func (a *App) addBoard(w http.ResponseWriter, r *http.Request) {

}

func (a *App) addPost(w http.ResponseWriter, r *http.Request) {

}

func (a *App) addComment(w http.ResponseWriter, r *http.Request) {

}

func (a *App) updatePost(w http.ResponseWriter, r *http.Request) {

}

func (a *App) updateComment(w http.ResponseWriter, r *http.Request) {

}

func (a *App) deletePost(w http.ResponseWriter, r *http.Request) {

}

func (a *App) deleteComment(w http.ResponseWriter, r *http.Request) {

}

func (a *App) Run(host string) {
	a.Negroni = negroni.Classic()
	a.Negroni.UseHandler(a.Router)
	//a.Negroni.Use(a.CORS)
	//a.AuthNegroni.Use(a.CORS)
	a.Auditor.Log("", "Start Server", "Success", "")
	a.Negroni.Run(host)
}
