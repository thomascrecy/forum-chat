package controllers

import (
	"database/sql"
	"fmt"
	models "forum/model"
	"html/template"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	_ "image/jpeg"
	_ "image/png"

	_ "github.com/mattn/go-sqlite3"
)

// ===== STRUCTS =====
type AuthPage struct {
	Username string
	Email    string

	UsernameMessage      string
	EmailMessage         string
	PasswordMessage      string
	GlobalMessage        string
	GlobalSuccessMessage string
}

type ProfilePage struct {
	User          *models.User
	Communities   []*models.Community
	Posts         []*models.Post
	IsAuth        bool
	IsCurrentUser bool
	CurrentUser   *models.User
}

type Cookie struct {
	Name  string
	Value string

	Path       string
	Domain     string
	Expires    time.Time
	RawExpires string

	MaxAge   int
	Secure   bool
	HttpOnly bool
	SameSite http.SameSite
	Raw      string
	Unparsed []string
}

// ===================

func isImage(header *multipart.FileHeader) bool {
	allowedImageTypes := []string{"image/jpeg", "image/png", "image/gif"}
	for _, mimeType := range allowedImageTypes {
		if header.Header.Get("Content-Type") == mimeType {
			return true
		}
	}
	return false
}

func createOrOpenFile(filePath string) (*os.File, error) {
	err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm)
	if err != nil {
		return nil, err
	}
	return os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
}

func RegisterHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, sessionErr := r.Cookie("session")
		if sessionErr == nil {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}

		p := &AuthPage{}

		if r.Method == http.MethodPost {
			username := r.FormValue("username")
			email := r.FormValue("email")
			password := r.FormValue("password")

			p.Username = username
			p.Email = email

			if err := models.RegisterUser(db, username, email, password); err != nil {
				switch err {
				case models.ErrInvalidUsername:
					p.UsernameMessage = err.Error()

				case models.ErrInvalidEmail:
					p.EmailMessage = err.Error()

				case models.ErrInvalidPassword:
					p.PasswordMessage = err.Error()

				case models.ErrUserExists:
					p.GlobalMessage = err.Error()
				default:
					fmt.Println(err.Error())
					p.GlobalMessage = "Une erreur a eu lieu"
				}
			} else {
				p.GlobalSuccessMessage = "Enregistrement effectué ! Veuillez vous connecter"
				t, _ := template.ParseFiles("./view/login.html")
				t.Execute(w, p)
				return
			}
		}

		t, _ := template.ParseFiles("./view/register.html")
		t.Execute(w, p)
	}
}

func LoginHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, sessionErr := r.Cookie("session")
		if sessionErr == nil {
			http.Redirect(w, r, "/", http.StatusForbidden)
			return
		}

		p := &AuthPage{}

		if r.Method == http.MethodPost {
			email := r.FormValue("email")
			password := r.FormValue("password")
			remember := r.FormValue("remember")

			p.Email = email

			var maxAge int
			if remember == "on" {
				maxAge = 7 * 24 * 3600
			} else {
				maxAge = 0
			}

			if userId, err := models.LoginUser(db, email, password); err != nil {
				switch err {
				case models.ErrInvalidAuth:
					p.GlobalMessage = err.Error()

				case models.ErrInvalidEmail:
					p.EmailMessage = err.Error()
				default:
					p.GlobalMessage = "Une erreur a eu lieu"
				}
			} else {
				tokenString, err := createJWT(userId)
				if err != nil {
					fmt.Println("Erreur lors de la création du JWT:", err)
					return
				} else {
					cookie := http.Cookie{
						Name:     "session",
						Value:    tokenString,
						Path:     "/",
						MaxAge:   maxAge,
						HttpOnly: true,
						Secure:   true,
						SameSite: http.SameSiteNoneMode,
					}

					http.SetCookie(w, &cookie)
				}

				http.Redirect(w, r, "/", http.StatusFound)
				return
			}
		}

		t, _ := template.ParseFiles("./view/login.html")
		t.Execute(w, p)
	}
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	_, sessionErr := r.Cookie("session")
	if sessionErr == nil {
		cookie := http.Cookie{
			Name:     "session",
			Value:    "",
			Path:     "/",
			Expires:  time.Unix(0, 0),
			HttpOnly: true,
			Secure:   true,
		}

		http.SetCookie(w, &cookie)
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

func ProfileHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		userQuery := query.Get("user")

		sessionCookie, sessionErr := r.Cookie("session")

		if userQuery == "" {
			if sessionErr != nil {
				http.Redirect(w, r, "/login", http.StatusFound)
				return
			}
		}

		communities, _ := models.GetAllCommunities(db)

		var targetUserID int
		var err error
		if userQuery != "" {
			targetUserID, err = strconv.Atoi(userQuery)
			if err != nil {
				http.Redirect(w, r, "/404", http.StatusFound)
				return
			}
		}

		currentUser := &models.User{}
		if sessionErr == nil {
			currentUser.ID, err = GetUserIDFromCookie(sessionCookie)
			if err != nil {
				http.Redirect(w, r, "/login", http.StatusFound)
				return
			}
			currentUser, _ = models.GetUser(db, currentUser.ID)
			if currentUser == nil {
				http.Redirect(w, r, "/404", http.StatusFound)
				return
			}
		}

		var user *models.User
		var posts []*models.Post
		if userQuery != "" {
			posts, _ = models.GetAllPostsFromAuthor(db, targetUserID)

			user, _ = models.GetUser(db, targetUserID)
			if user == nil {
				http.Redirect(w, r, "/404", http.StatusFound)
				return
			}
		} else {
			posts, _ = models.GetAllPostsFromAuthor(db, currentUser.ID)
			user = currentUser
		}

		p := &ProfilePage{
			User:          user,
			Posts:         posts,
			Communities:   communities,
			IsAuth:        sessionErr == nil,
			IsCurrentUser: ((userQuery == "") || (sessionErr == nil && user.ID == currentUser.ID)),
			CurrentUser:   currentUser,
		}

		t, _ := template.ParseFiles("./view/profile.html")
		t.Execute(w, p)
	}
}

func EditProfileHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		newName := query.Get("name")
		newBio := query.Get("bio")

		sessionCookie, sessionErr := r.Cookie("session")
		if sessionErr != nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		if len(newName) <= 24 && len(newBio) <= 512 {
			userID, _ := GetUserIDFromCookie(sessionCookie)

			if newName != "" {
				models.EditUserName(db, userID, newName)
			}

			if newBio != "" {
				models.EditUserBio(db, userID, newBio)
			}
		}

		http.Redirect(w, r, "/profile", http.StatusFound)
	}
}

func UploadUserImageHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Redirect(w, r, "/", http.StatusForbidden)
		return
	}

	sessionCookie, sessionErr := r.Cookie("session")
	if sessionErr != nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	if !isImage(header) {
		http.Error(w, "Invalid image format", http.StatusBadRequest)
		return
	}

	currentUserID, sessionCookieErr := GetUserIDFromCookie(sessionCookie)
	if sessionCookieErr != nil {
		http.Redirect(w, r, "/login", http.StatusInternalServerError)
		return
	}

	outFile, err := createOrOpenFile(models.UsersImageUploadPath + "/" + strconv.Itoa(currentUserID) + "/" + models.DefaultProfileImageFilename)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/profile", http.StatusAccepted)
}

func UserImageHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	userTargetID := query.Get("id")

	if userTargetID == "" {
		http.Error(w, "Image introuvable", http.StatusNotFound)
		return
	}

	file, err := os.Open(models.UsersImageUploadPath + "/" + userTargetID + "/" + models.DefaultProfileImageFilename)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	defer file.Close()

	_, err = io.Copy(w, file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
