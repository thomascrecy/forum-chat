package controllers

import (
	"database/sql"
	"fmt"
	models "forum/model"
	"net/http"
	"strconv"
)

// ===== STRUCTS =====
type CommentPage struct {
	Title string
}

//===================

func CommentHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionCookie, sessionErr := r.Cookie("session")
		if sessionErr != nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		postIdStr := r.URL.Query().Get("id")
		postId, err := strconv.Atoi(postIdStr)
		if err != nil {
			http.Error(w, "ID de commentaire invalide", http.StatusBadRequest)
			return
		}

		//Si c'est pour publier un commentaire
		if r.Method == http.MethodPost {
			content := r.FormValue("content")
			author, err := GetUserIDFromCookie(sessionCookie)
			if err == nil {
				post := postId

				if err := models.PostComment(db, author, content, post); err != nil {
					switch err {
					default:
						fmt.Println("Une erreur a eu lieu lors de la publication du commentaire")
					}
				}
			}

			http.Redirect(w, r, "/post?id="+strconv.Itoa(postId), http.StatusSeeOther)
			return
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}
