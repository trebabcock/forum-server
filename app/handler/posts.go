package handler

import (
	"encoding/json"
	"fmt"
	"forum-server/app/model"
	"log"
	"net/http"
	"time"

	"github.com/form3tech-oss/jwt-go"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

func GetPost(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	postId := vars["postId"]
	post, err := getPostById(db, postId)
	if err != nil {
		RespondError(w, http.StatusNotFound, "post not found")
		return
	}
	RespondJSON(w, http.StatusOK, post)
}

func GetPostsFromUser(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userId := vars["userId"]
	posts, err := getPostsFromUser(db, userId)
	if err != nil {
		RespondError(w, http.StatusNotFound, "posts not found")
		return
	}
	RespondJSON(w, http.StatusOK, posts)
}

func GetPostsFromBoard(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	boardId := vars["boardId"]
	posts, err := getPostsFromBoard(db, boardId)
	if err != nil {
		RespondError(w, http.StatusNotFound, "posts not found")
		return
	}
	RespondJSON(w, http.StatusOK, posts)
}

func UpdatePost(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userCtx := ctx.Value("user")

	reqId := userCtx.(*jwt.Token).Claims.(jwt.MapClaims)["id"]

	vars := mux.Vars(r)
	postId := vars["postId"]

	post, err := getPostById(db, postId)
	if err != nil {
		RespondError(w, http.StatusNotFound, "post not found")
		return
	}

	if reqId != post.AuthorID {
		RespondError(w, http.StatusUnauthorized, "no access")
		return
	}

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&post); err != nil {
		RespondError(w, http.StatusBadRequest, "")
		return
	}
	defer r.Body.Close()

	if err := db.Save(&post).Error; err != nil {
		RespondError(w, http.StatusInternalServerError, "")
		return
	}
	RespondJSON(w, http.StatusOK, post)
}

func AddPost(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userCtx := ctx.Value("user")

	reqId := userCtx.(*jwt.Token).Claims.(jwt.MapClaims)["id"]

	newPost := model.NewPost{}
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&newPost); err != nil {
		RespondError(w, http.StatusBadRequest, "an unknown error has occurred")
		return
	}
	defer r.Body.Close()

	vars := mux.Vars(r)
	boardId := vars["boardId"]

	postId, err := uuid.NewUUID()
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "an unknown error has occurred")
		return
	}

	post := model.Post{
		ID:         postId.String(),
		AuthorID:   fmt.Sprintf("%v", reqId),
		BoardID:    boardId,
		Title:      newPost.Title,
		Content:    newPost.Content,
		CreateDate: time.Now().UTC().Format(time.RFC3339),
	}

	if err = db.Save(&post).Error; err != nil {
		RespondError(w, http.StatusInternalServerError, "an unknown error has occurred")
		return
	}

	RespondJSON(w, http.StatusOK, map[string]string{"id": post.ID})
}

func DeletePost(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userCtx := ctx.Value("user")

	reqId := userCtx.(*jwt.Token).Claims.(jwt.MapClaims)["id"]
	userRole, err := getUserRoleById(db, fmt.Sprintf("%v", reqId))
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "an unknown error has occurred")
		return
	}

	vars := mux.Vars(r)
	postId := vars["postId"]

	post, err := getPostById(db, postId)
	if err != nil {
		RespondError(w, http.StatusNotFound, "post not found")
		return
	}

	if reqId != post.AuthorID && userRole != "admin" && userRole != "moderator" {
		RespondError(w, http.StatusUnauthorized, "no access")
		return
	}

	if err := db.Delete(&post).Error; err != nil {
		RespondError(w, http.StatusInternalServerError, "")
		return
	}
	RespondJSON(w, http.StatusNoContent, nil)
}

func GetPostAuthor(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	postId := vars["postId"]
	post, err := getPostById(db, postId)
	if err != nil {
		RespondError(w, http.StatusNotFound, "post not found")
		log.Fatal(err)
		return
	}

	author, err := publicUser(db, post.AuthorID)
	if err != nil {
		log.Fatal(err)
		RespondError(w, http.StatusNotFound, "user not found")
	}

	RespondJSON(w, http.StatusOK, author)
}

func getPostById(db *gorm.DB, postId string) (*model.Post, error) {
	post := model.Post{}
	if err := db.Where(&model.Post{ID: postId}).First(&post).Error; err != nil {
		return nil, err
	}
	return &post, nil
}

func getPostsFromUser(db *gorm.DB, userId string) (*[]model.Post, error) {
	var posts []model.Post
	if err := db.Where(&model.Post{AuthorID: userId}).Find(&posts).Error; err != nil {
		return nil, err
	}
	return &posts, nil
}

func getPostsFromBoard(db *gorm.DB, boardId string) (*[]model.Post, error) {
	var posts []model.Post
	if err := db.Where(&model.Post{BoardID: boardId}).Find(&posts).Error; err != nil {
		return nil, err
	}
	return &posts, nil
}
