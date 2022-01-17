package app

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"forum-server/app/handler"
	"forum-server/audit"
	db "forum-server/db"

	jwtmiddleware "github.com/auth0/go-jwt-middleware"
	"github.com/form3tech-oss/jwt-go"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	_ "github.com/joho/godotenv/autoload"
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
	a.get("/api/checkRole", a.checkRole)

	// TODO: find better names for routes

	a.postNoAuth("/api/login", a.login)
	a.postNoAuth("/api/register", a.register)
	a.get("/api/users", a.getUsers)
	a.get("/api/user/{userId}", a.getUserById)
	a.getNoAuth("/api/user/public/{userId}", a.getPublicUser)
	a.getNoAuth("/api/user/publicByUsername/{username}", a.getPublicUserByUsername)
	a.put("/api/user/{userId}", a.updateUser)
	a.delete("/api/user/{userId}", a.deleteUser)

	a.getNoAuth("/api/boards", a.getBoards)
	a.getNoAuth("/api/board/{boardId}", a.getBoard)
	a.getNoAuth("/api/board/{boardId}/posts", a.getPostsFromBoard)
	a.getNoAuth("/api/posts/{postId}/comments", a.getCommentsFromPost)
	a.getNoAuth("/api/posts/{postId}", a.getPost)
	a.getNoAuth("/api/board/{boardId}/lastPost", a.getLastPost)
	a.post("/api/boards/addBoard", a.addBoard)
	a.put("/api/boards/{boardId}", a.updateBoard)
	a.delete("/api/boards/{boardId}", a.deleteBoard)
	a.post("/api/boards/{boardId}/newPost", a.addPost)
	a.post("/api/post/addComment", a.addComment)
	a.put("/api/posts/{postId}", a.updatePost)
	a.put("/api/posts/comments/{commentId}", a.updateComment)
	a.delete("/api/posts/{postId}", a.deletePost)
	a.delete("/api/posts/comments{commentId}", a.deleteComment)

	a.getNoAuth("/api/user/fromPost/{postId}", a.getPostAuthor)
	a.getNoAuth("/api/post/{postId}/getLastComment", a.getLastCommentFromPost)
	a.post("/api/user/{userId}/avatar", a.uploadAvatar)
}

func (a *App) getNoAuth(path string, f func(w http.ResponseWriter, r *http.Request)) {
	a.Router.HandleFunc(path, f).Methods("GET", "OPTIONS")
}

func (a *App) postNoAuth(path string, f func(w http.ResponseWriter, r *http.Request)) {
	a.Router.HandleFunc(path, f).Methods("POST", "OPTIONS")
}

func (a *App) get(path string, f func(w http.ResponseWriter, r *http.Request)) {
	a.AuthRouter.HandleFunc(path, f).Methods("GET", "OPTIONS")
}

func (a *App) post(path string, f func(w http.ResponseWriter, r *http.Request)) {
	a.AuthRouter.HandleFunc(path, f).Methods("POST", "OPTIONS")
}

func (a *App) put(path string, f func(w http.ResponseWriter, r *http.Request)) {
	a.AuthRouter.HandleFunc(path, f).Methods("PUT", "OPTIONS")
}

func (a *App) delete(path string, f func(w http.ResponseWriter, r *http.Request)) {
	a.AuthRouter.HandleFunc(path, f).Methods("DELETE", "OPTIONS")
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

func (a *App) checkRole(w http.ResponseWriter, r *http.Request) {
	handler.CheckRole(a.DB, w, r)
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

func (a *App) getPublicUser(w http.ResponseWriter, r *http.Request) {
	handler.GetPublicUser(a.DB, w, r)
}

func (a *App) getPostAuthor(w http.ResponseWriter, r *http.Request) {
	handler.GetPostAuthor(a.DB, w, r)
}

func (a *App) getPublicUserByUsername(w http.ResponseWriter, r *http.Request) {
	handler.GetPublicUserByUsername(a.DB, w, r)
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
	handler.GetBoards(a.DB, w, r)
}

func (a *App) getBoard(w http.ResponseWriter, r *http.Request) {
	handler.GetBoard(a.DB, w, r)
}

func (a *App) updateBoard(w http.ResponseWriter, r *http.Request) {
	handler.UpdateBoard(a.DB, w, r)
}

func (a *App) deleteBoard(w http.ResponseWriter, r *http.Request) {
	handler.DeleteBoard(a.DB, w, r)
}

func (a *App) getPostsFromBoard(w http.ResponseWriter, r *http.Request) {
	handler.GetPostsFromBoard(a.DB, w, r)
}

func (a *App) getLastPost(w http.ResponseWriter, r *http.Request) {
	handler.GetLastPostTimeAndAuthor(a.DB, w, r)
}

func (a *App) getPost(w http.ResponseWriter, r *http.Request) {
	handler.GetPost(a.DB, w, r)
}

func (a *App) getCommentsFromPost(w http.ResponseWriter, r *http.Request) {
	handler.GetCommentsFromPost(a.DB, w, r)
}

func (a *App) getLastCommentFromPost(w http.ResponseWriter, r *http.Request) {
	handler.GetLastCommentFromPost(a.DB, w, r)
}

func (a *App) addBoard(w http.ResponseWriter, r *http.Request) {
	handler.CreateBoard(a.DB, w, r)
}

func (a *App) addPost(w http.ResponseWriter, r *http.Request) {
	handler.AddPost(a.DB, w, r)
}

func (a *App) addComment(w http.ResponseWriter, r *http.Request) {
	handler.AddComment(a.DB, w, r)
}

func (a *App) uploadAvatar(w http.ResponseWriter, r *http.Request) {
	handler.UploadAvatar(a.DB, w, r)
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

	headers := handlers.AllowedHeaders([]string{"Accept", "Content-Type", "Content-Length", "Accept-Encoding", "X-CSRF-Token", "Authorization", "Origin", "Cache-Control", "X-Requested-With"})
	methods := handlers.AllowedMethods([]string{"GET", "PUT", "POST", "DELETE", "OPTIONS"})
	origins := handlers.AllowedOrigins([]string{"*"})

	server := &http.Server{
		Addr:         host,
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      handlers.CORS(headers, methods, origins)(a.Negroni),
	}

	log.Fatal(server.ListenAndServe())
	//a.Negroni.Run(host)
}
