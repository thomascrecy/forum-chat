package controllers

import (
	"database/sql"
	"fmt"
	models "forum/model"
	"net/http"
	"strconv"
)

//===== STRUCTS =====
//===================

func LikeHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionCookie, sessionErr := r.Cookie("session")
		if sessionErr != nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		if r.Method == http.MethodPost {
			postIdStr := r.URL.Query().Get("postId")
			postId, err := strconv.Atoi(postIdStr)
			if err != nil {
				http.Error(w, "ID de post invalide", http.StatusBadRequest)
				return
			}
			authorId, err := GetUserIDFromCookie(sessionCookie)
			if err == nil {
				linkId := postId

				isComment := r.URL.Query().Get("commentId") != ""
				if isComment {
					commentIdStr := r.URL.Query().Get("commentId")
					commentId, err := strconv.Atoi(commentIdStr)
					if err != nil {
						http.Error(w, "ID de commentaire invalide", http.StatusBadRequest)
						return
					}

					linkId = commentId
				}

				_, err = models.SendLike(db, linkId, authorId, isComment)
				if err != nil {
					switch err {
					default:
						fmt.Println("Une erreur a eu lieu lors de l'ajout du like")
					}
				}
			}

			http.Redirect(w, r, "/post?id="+strconv.Itoa(postId), http.StatusSeeOther)
			return
		} else {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
	}
}
