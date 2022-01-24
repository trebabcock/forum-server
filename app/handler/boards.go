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

func GetBoards(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	boards := []model.Board{}
	db.Find(&boards)
	RespondJSON(w, http.StatusOK, boards)
}

func GetBoard(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["boardId"]
	board, err := getBoardByID(db, id)
	if err != nil {
		RespondError(w, http.StatusNotFound, "board not found")
		return
	}
	RespondJSON(w, http.StatusOK, board)
}

func CreateBoard(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userCtx := ctx.Value("user")

	reqId := userCtx.(*jwt.Token).Claims.(jwt.MapClaims)["id"]
	userRole, err := getUserRoleById(db, fmt.Sprintf("%v", reqId))
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "an unknown error has occurred")
		return
	}

	if userRole != "admin" {
		RespondError(w, http.StatusUnauthorized, "no access")
		return
	}

	newBoard := model.NewBoard{}
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&newBoard); err != nil {
		RespondError(w, http.StatusBadRequest, "")
		return
	}
	defer r.Body.Close()

	id, err := uuid.NewUUID()
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "an unknown error has occurred")
		return
	}

	board := model.Board{
		ID:          id.String(),
		Name:        newBoard.Name,
		Description: newBoard.Description,
		CreateDate:  time.Now().UTC(),
	}

	if err := db.Save(&board).Error; err != nil {
		RespondError(w, http.StatusInternalServerError, "an unknown error has occurred")
		return
	}
	RespondJSON(w, http.StatusCreated, board)
}

func UpdateBoard(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userCtx := ctx.Value("user")

	reqId := userCtx.(*jwt.Token).Claims.(jwt.MapClaims)["id"]
	userRole, err := getUserRoleById(db, fmt.Sprintf("%v", reqId))
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "an unknown error has occurred")
		return
	}

	if userRole != "admin" {
		RespondError(w, http.StatusUnauthorized, "no access")
		return
	}

	vars := mux.Vars(r)
	boardId := vars["boardId"]

	board, err := getBoardByID(db, boardId)
	if err != nil {
		RespondError(w, http.StatusNotFound, "board not found")
		return
	}

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&board); err != nil {
		RespondError(w, http.StatusBadRequest, "")
		return
	}
	defer r.Body.Close()

	if err := db.Save(&board).Error; err != nil {
		RespondError(w, http.StatusInternalServerError, "")
		return
	}
	RespondJSON(w, http.StatusOK, board)
}

func DeleteBoard(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userCtx := ctx.Value("user")

	reqId := userCtx.(*jwt.Token).Claims.(jwt.MapClaims)["id"]
	userRole, err := getUserRoleById(db, fmt.Sprintf("%v", reqId))
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "an unknown error has occurred")
		return
	}

	if userRole != "admin" {
		RespondError(w, http.StatusUnauthorized, "no access")
		return
	}

	vars := mux.Vars(r)
	boardId := vars["boardId"]

	board, err := getBoardByID(db, boardId)
	if err != nil {
		RespondError(w, http.StatusNotFound, "board not found")
		return
	}

	if err := db.Delete(&board).Error; err != nil {
		RespondError(w, http.StatusInternalServerError, "")
		return
	}
	RespondJSON(w, http.StatusNoContent, nil)
}

func GetLastPostTimeAndAuthor(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["boardId"]

	post, err := getLastPost(db, id)
	if err != nil {
		RespondError(w, http.StatusNotFound, "post not found")
		return
	}

	user, err := getUserById(db, post.AuthorID)
	if err != nil {
		RespondError(w, http.StatusNotFound, "user not found")
		return
	}

	count := getPostCount(db, id)

	RespondJSON(w, http.StatusOK, map[string]interface{}{"author": user.Username, "date_time": post.CreateDate, "count": count})
}

func GetBoardFromPost(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	postId := vars["postId"]

	post, err := getPostById(db, postId)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "post not found")
		return
	}

	board, err := getBoardByID(db, post.BoardID)
	if err != nil {
		RespondError(w, http.StatusBadRequest, "board not found")
		return
	}

	RespondJSON(w, http.StatusOK, board)
}

func getLastPost(db *gorm.DB, boardId string) (*model.Post, error) {
	post := model.Post{}
	if err := db.Where(&model.Post{BoardID: boardId}).Order("create_date desc").First(&post).Error; err != nil {
		return nil, err
	}
	return &post, nil
}

func getPostCount(db *gorm.DB, boardId string) int64 {
	var count int64
	if err := db.Model(&model.Post{}).Where(&model.Post{BoardID: boardId}).Count(&count).Error; err != nil {
		return 0
	}
	return count
}

func getBoardByID(db *gorm.DB, boardId string) (*model.Board, error) {
	board := model.Board{}
	if err := db.Where(&model.Board{ID: boardId}).First(&board).Error; err != nil {
		return nil, err
	}
	return &board, nil
}
