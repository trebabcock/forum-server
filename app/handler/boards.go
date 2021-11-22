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
		CreateDate:  time.Now(),
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

func getBoardByID(db *gorm.DB, boardId string) (*model.Board, error) {
	board := model.Board{}
	if err := db.Where(&model.Board{ID: boardId}).First(&board).Error; err != nil {
		return nil, err
	}
	return &board, nil
}
