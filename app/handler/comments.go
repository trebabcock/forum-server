package handler

import (
	"encoding/json"
	"fmt"
	"forum-server/app/model"
	"net/http"
	"time"

	"github.com/form3tech-oss/jwt-go"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"gorm.io/gorm"
)

func GetComment(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	commentId := vars["commentId"]
	comment, err := getCommentById(db, commentId)
	if err != nil {
		RespondError(w, http.StatusNotFound, "comment not found")
		return
	}
	RespondJSON(w, http.StatusOK, comment)
}

func GetCommentsFromUser(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userId := vars["userId"]
	comments, err := getCommentsFromUser(db, userId)
	if err != nil {
		RespondError(w, http.StatusNotFound, "comments not found")
		return
	}
	RespondJSON(w, http.StatusOK, comments)
}

func GetCommentsFromPost(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	postId := vars["postId"]
	comments, err := getCommentsFromPost(db, postId)
	if err != nil {
		RespondError(w, http.StatusNotFound, "comments not found")
		return
	}
	RespondJSON(w, http.StatusOK, comments)
}

func AddComment(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userCtx := ctx.Value("user")

	reqId := userCtx.(*jwt.Token).Claims.(jwt.MapClaims)["id"]

	newComment := model.NewComment{}
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&newComment); err != nil {
		RespondError(w, http.StatusBadRequest, "an unknown error has occurred")
		return
	}
	defer r.Body.Close()

	commentId, err := uuid.NewUUID()
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "an unknown error has occurred")
		return
	}

	comment := model.Comment{
		ID:         commentId.String(),
		AuthorID:   fmt.Sprintf("%v", reqId),
		PostID:     newComment.PostID,
		ParentID:   "",
		Content:    newComment.Content,
		CreateDate: time.Now().UTC().Format(time.RFC3339),
	}

	if err = db.Save(&comment).Error; err != nil {
		RespondError(w, http.StatusInternalServerError, "an unknown error has occurred")
		return
	}

	RespondJSON(w, http.StatusOK, comment)
}

func UpdateComment(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userCtx := ctx.Value("user")

	reqId := userCtx.(*jwt.Token).Claims.(jwt.MapClaims)["id"]

	vars := mux.Vars(r)
	commentId := vars["commentId"]

	comment, err := getCommentById(db, commentId)
	if err != nil {
		RespondError(w, http.StatusNotFound, "comment not found")
		return
	}

	if reqId != comment.AuthorID {
		RespondError(w, http.StatusUnauthorized, "no access")
		return
	}

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&comment); err != nil {
		RespondError(w, http.StatusBadRequest, "")
		return
	}
	defer r.Body.Close()

	if err := db.Save(&comment).Error; err != nil {
		RespondError(w, http.StatusInternalServerError, "")
		return
	}
	RespondJSON(w, http.StatusOK, comment)
}

func DeleteComment(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userCtx := ctx.Value("user")

	reqId := userCtx.(*jwt.Token).Claims.(jwt.MapClaims)["id"]
	userRole, err := getUserRoleById(db, fmt.Sprintf("%v", reqId))
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "an unknown error has occurred")
		return
	}

	vars := mux.Vars(r)
	commentId := vars["commentId"]

	comment, err := getCommentById(db, commentId)
	if err != nil {
		RespondError(w, http.StatusNotFound, "comment not found")
		return
	}

	if reqId != comment.AuthorID && userRole != "admin" && userRole != "moderator" {
		RespondError(w, http.StatusUnauthorized, "no access")
		return
	}

	if err := db.Delete(&comment).Error; err != nil {
		RespondError(w, http.StatusInternalServerError, "")
		return
	}
	RespondJSON(w, http.StatusNoContent, nil)
}

func GetLastCommentFromPost(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["postId"]

	comment, err := getLastComment(db, id)
	if err != nil {
		RespondError(w, http.StatusNotFound, "comment not found")
		return
	}

	user, err := getUserById(db, comment.AuthorID)
	if err != nil {
		RespondError(w, http.StatusNotFound, "user not found")
		return
	}

	count := getCommentCount(db, id)

	RespondJSON(w, http.StatusOK, map[string]interface{}{"author": user.Username, "date_time": comment.CreateDate, "count": count})
}

func getCommentCount(db *gorm.DB, postId string) int64 {
	var count int64
	if err := db.Model(&model.Comment{}).Where(&model.Comment{PostID: postId}).Count(&count).Error; err != nil {
		return 0
	}
	return count
}

func getLastComment(db *gorm.DB, postId string) (*model.Comment, error) {
	comment := model.Comment{}
	if err := db.Where(&model.Comment{PostID: postId}).Order("create_date desc").First(&comment).Error; err != nil {
		return nil, err
	}
	return &comment, nil
}

func getCommentById(db *gorm.DB, commentId string) (*model.Comment, error) {
	comment := model.Comment{}
	if err := db.Where(&model.Comment{ID: commentId}).First(&comment).Error; err != nil {
		return nil, err
	}
	return &comment, nil
}

func getCommentsFromUser(db *gorm.DB, userId string) (*[]model.Comment, error) {
	var comments []model.Comment
	if err := db.Where(&model.Comment{AuthorID: userId}).Find(&comments).Error; err != nil {
		return nil, err
	}
	return &comments, nil
}

func getCommentsFromPost(db *gorm.DB, postId string) (*[]model.Comment, error) {
	var comments []model.Comment
	if err := db.Where(&model.Comment{PostID: postId}).Find(&comments).Error; err != nil {
		return nil, err
	}
	return &comments, nil
}
