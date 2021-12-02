package handler

import (
	"encoding/json"
	"fmt"
	"forum-server/app/auth"
	"forum-server/app/model"
	"forum-server/audit"
	"log"
	"net/http"
	"time"

	"github.com/form3tech-oss/jwt-go"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// TODO: implement pagination
func GetUsers(db *gorm.DB, auditor *audit.Auditor, w http.ResponseWriter, r *http.Request) {
	var IDs []string
	db.Model(&model.User{}).Pluck("id", &IDs)

	var users []model.PublicUser
	for _, id := range IDs {
		user, _ := publicUser(db, id)
		users = append(users, *user)
	}

	RespondJSON(w, http.StatusOK, users)
}

func GetPublicUser(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["userId"]

	retUser, err := publicUser(db, id)
	if err != nil {
		RespondError(w, http.StatusNotFound, "user not found")
		return
	}

	RespondJSON(w, http.StatusOK, retUser)
}

func GetPublicUserByUsername(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	username := vars["username"]

	user, err := getUserByUsername(db, username)
	if err != nil {
		RespondError(w, http.StatusNotFound, "user not found")
	}

	retUser, err := publicUser(db, user.ID)
	if err != nil {
		RespondError(w, http.StatusNotFound, "user not found")
		return
	}

	RespondJSON(w, http.StatusOK, retUser)
}

func GetUserById(db *gorm.DB, auditor *audit.Auditor, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userCtx := ctx.Value("user")

	vars := mux.Vars(r)
	id := vars["userId"]

	reqId := userCtx.(*jwt.Token).Claims.(jwt.MapClaims)["id"]

	userRole, err := getUserRoleById(db, fmt.Sprintf("%v", reqId))
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "an unknown error has occurred")
		return
	}

	if reqId != id {
		if userRole != "admin" {
			retUser, err := publicUser(db, id)
			if err != nil {
				RespondError(w, http.StatusNotFound, "user not found")
				return
			}
			RespondJSON(w, http.StatusOK, retUser)
			return
		}
	}

	retUser, err := getUserById(db, id)
	if err != nil {
		RespondError(w, http.StatusNotFound, "user not found")
		return
	}

	RespondJSON(w, http.StatusOK, retUser)
	auditor.Log("", "Get User", "Success", id)
}

func UpdateUser(db *gorm.DB, auditor *audit.Auditor, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userCtx := ctx.Value("user")

	vars := mux.Vars(r)
	id := vars["userId"]

	reqId := userCtx.(*jwt.Token).Claims.(jwt.MapClaims)["id"]

	if reqId != id {
		RespondError(w, http.StatusUnauthorized, "no access")
		return
	}

	user, err := getUserById(db, id)
	if err != nil {
		RespondError(w, http.StatusNotFound, "user not found")
		return
	}

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&user); err != nil {
		RespondError(w, http.StatusInternalServerError, "an unknown error has occurred")
		return
	}
	defer r.Body.Close()

	if err := db.Save(&user).Error; err != nil {
		RespondError(w, http.StatusInternalServerError, "an unknown error has occurred")
		return
	}
	RespondJSON(w, http.StatusOK, user)
}

func DeleteUser(db *gorm.DB, auditor *audit.Auditor, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userCtx := ctx.Value("user")

	vars := mux.Vars(r)
	id := vars["userId"]

	reqId := userCtx.(*jwt.Token).Claims.(jwt.MapClaims)["id"]

	if reqId != id {
		RespondError(w, http.StatusUnauthorized, "no access")
		return
	}

	user, err := getUserById(db, id)
	if err != nil {
		RespondError(w, http.StatusNotFound, "user not found")
		return
	}
	if err := db.Delete(&user).Error; err != nil {
		RespondError(w, http.StatusInternalServerError, "")
		return
	}
	RespondJSON(w, http.StatusNoContent, nil)
}

func UserLogin(db *gorm.DB, auditor *audit.Auditor, w http.ResponseWriter, r *http.Request) {
	creds := model.LoginCredentials{}

	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&creds); err != nil {
		log.Println("ERROR DECODE:", err)
		RespondError(w, http.StatusBadRequest, "an unknown error has occurred")
		return
	}

	log.Println(creds.Email, creds.Password)

	user, err := getUserByEmail(db, creds.Email)
	if err != nil {
		log.Println("ERROR GET USER:", err)
		RespondError(w, http.StatusUnauthorized, "invalid username or password")
		return
	}

	log.Println(user.Username, user.Email, user.Password)

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(creds.Password)); err != nil {
		RespondError(w, http.StatusUnauthorized, "invalid username or password")
		return
	}

	token, err := auth.GenerateToken(user)
	if err != nil {
		log.Println("ERROR GENERATE:", err)
		RespondError(w, http.StatusInternalServerError, "an unknown error has occurred")
		return
	}

	resp := model.LoginResponse{
		ID:       user.ID,
		Username: user.Username,
		Token:    token,
	}

	RespondJSON(w, http.StatusOK, resp)
}

// TODO: log errors

func UserRegister(db *gorm.DB, auditor *audit.Auditor, w http.ResponseWriter, r *http.Request) {
	creds := model.RegisterCredentials{}
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&creds); err != nil {
		RespondError(w, http.StatusBadRequest, "an unknown error has occurred")
		return
	}
	defer r.Body.Close()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(creds.Password), 8)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "an unknown error has occurred")
		return
	}

	id, err := uuid.NewUUID()
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "an unknown error has occurred")
		return
	}

	user := model.User{
		ID:         id.String(),
		Username:   creds.Username,
		Email:      creds.Email,
		Password:   string(hashedPassword),
		Bio:        "",
		Reputation: 0,
		AvatarURL:  "", // will add default later
		Role:       "user",
		Active:     true,
		CreateDate: time.Now().UTC().Format(time.RFC3339),
	}

	if err := db.Save(&user).Error; err != nil {
		RespondError(w, http.StatusInternalServerError, "")
		return
	}

	// TODO: Decide what to return. Also, automatically login after register, or redirect to login page?

	RespondJSON(w, http.StatusOK, "")
}

func BanUser(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userCtx := ctx.Value("user")

	reqId := userCtx.(*jwt.Token).Claims.(jwt.MapClaims)["id"]
	userRole, err := getUserRoleById(db, fmt.Sprintf("%v", reqId))
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "an unknown error has occurred")
		return
	}

	if userRole != "admin" && userRole != "moderator" {
		RespondError(w, http.StatusUnauthorized, "no access")
		return
	}

	vars := mux.Vars(r)
	userId := vars["userId"]

	user, err := getUserById(db, userId)
	if err != nil {
		RespondError(w, http.StatusNotFound, "user not found")
		return
	}

	user.Active = false

	if err := db.Save(&user).Error; err != nil {
		RespondError(w, http.StatusInternalServerError, "")
		return
	}
	RespondJSON(w, http.StatusOK, user)
}

func CheckRole(db *gorm.DB, w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userCtx := ctx.Value("user")

	reqId := userCtx.(*jwt.Token).Claims.(jwt.MapClaims)["id"]

	userRole, err := getUserRoleById(db, fmt.Sprintf("%v", reqId))
	if err != nil {
		RespondError(w, http.StatusInternalServerError, "an unknown error has occurred")
		return
	}

	RespondJSON(w, http.StatusOK, map[string]string{"role": userRole})
}

func publicUser(db *gorm.DB, userId string) (*model.PublicUser, error) {
	private := model.User{}
	if err := db.Where(&model.User{ID: userId}).Find(&private).Error; err != nil {
		return nil, err
	}

	public := model.PublicUser{
		ID:         private.ID,
		Username:   private.Username,
		Bio:        private.Bio,
		Reputation: private.Reputation,
		AvatarURL:  private.AvatarURL,
		Role:       private.Role,
		CreateDate: private.CreateDate,
	}

	return &public, nil
}

func getUserById(db *gorm.DB, userId string) (*model.User, error) {
	user := model.User{}
	if err := db.Where(&model.User{ID: userId}).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func getUserByUsername(db *gorm.DB, username string) (*model.User, error) {
	user := model.User{}
	if err := db.Where(&model.User{Username: username}).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func getUserByEmail(db *gorm.DB, email string) (*model.User, error) {
	user := model.User{}
	if err := db.Where(&model.User{Email: email}).First(&user).Error; err != nil {
		//if err := db.Find(&user).Where(&model.User{Email: email}).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func getUserRoleById(db *gorm.DB, userId string) (string, error) {
	user, err := getUserById(db, userId)
	if err != nil {
		return "", err
	}
	return user.Role, nil
}
