package controllers

import (
	"database/sql"
	"fmt"
	models "forum/model"
	"html/template"
	"net/http"
	"strconv"

	_ "github.com/mattn/go-sqlite3"
)

// ===== STRUCTS =====
type PostCreationPage struct {
	GlobalMessage  string
	AllTags        []*models.Tag
	AllCommunities []*models.Community
}

type PostPage struct {
	PostData   models.Post
	AuthorData models.User
	Comments   []*models.Comment
	IsAuth     bool
}

//===================

func PostHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sessionCookie, sessionErr := r.Cookie("session")
		if sessionErr != nil {
			sessionCookie = &http.Cookie{Value: ""}
		}

		p := &PostCreationPage{}

		if r.Method == http.MethodPost {
			if sessionErr != nil {
				http.Redirect(w, r, "/login", http.StatusFound)
				return
			}

			r.ParseForm()

			community := r.FormValue("community")
			title := r.FormValue("title")
			content := r.FormValue("content")
			tags := r.Form["tags"]

			authorId, sessionCookieErr := GetUserIDFromCookie(sessionCookie)
			if sessionCookieErr == nil {
				if postId, err := models.PostPost(db, community, title, authorId, content, tags); err != nil {
					switch err {
					case models.ErrTooLongTitle:
						p.GlobalMessage = err.Error()
					case models.ErrTooShortTitle:
						p.GlobalMessage = err.Error()
					case models.ErrInvalidContent:
						p.GlobalMessage = err.Error()
					case models.ErrCommuUnknow:
						p.GlobalMessage = err.Error()
					default:
						p.GlobalMessage = "Une erreur a eu lieu lors de la publication du post"
						fmt.Println(err.Error())
					}
				} else {
					http.Redirect(w, r, "/post?id="+strconv.Itoa(postId), http.StatusFound)
					return
				}
			}
		}

		query := r.URL.Query()
		idQuery := query.Get("id")
		if idQuery != "" {
			currentUserId, _ := GetUserIDFromCookie(sessionCookie)

			postData, err := models.GetPost(db, idQuery, currentUserId)
			if err != nil {
				fmt.Println(err)
				http.Redirect(w, r, "/404", http.StatusFound)
				return
			}

			authorData, err := models.GetUser(db, postData.Author)
			if err != nil {
				fmt.Println(err)
				authorData = nil
			}

			comments, err := models.GetAllCommentsOfPost(db, idQuery, currentUserId)
			if err != nil {
				http.Error(w, "Erreur lors de la récupération des commentaires", http.StatusInternalServerError)
				return
			}

			if models.HasUserLikedPost(db, postData.ID, currentUserId, false) {
				postData.Liked = true
			}

			p := &PostPage{
				PostData:   *postData,
				AuthorData: *authorData,
				Comments:   comments,
				IsAuth:     sessionErr == nil,
			}

			t, _ := template.ParseFiles("./view/postView.html")
			t.Execute(w, p)

		} else {
			if sessionErr != nil {
				http.Redirect(w, r, "/login", http.StatusFound)
				return
			}

			var err error
			p.AllTags, err = models.GetAllTags(db)
			if err != nil {
				p.AllTags = []*models.Tag{}
			}

			p.AllCommunities, err = models.GetAllCommunities(db)
			if err != nil {
				p.AllCommunities = []*models.Community{}
			}

			t, _ := template.ParseFiles("./view/post.html")
			t.Execute(w, p)
		}
	}
}

func EditPostHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		postID := query.Get("id")
		newTitle := query.Get("title")
		newContent := query.Get("content")

		sessionCookie, sessionErr := r.Cookie("session")
		if sessionErr != nil {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}

		if len(newTitle) <= 256 && len(newContent) <= 4000 && postID != "" {
			userID, err := GetUserIDFromCookie(sessionCookie)
			if err == nil {
				if newTitle != "" {
					models.EditPostTitle(db, userID, postID, newTitle)
				}

				if newContent != "" {
					models.EditPostContent(db, userID, postID, newContent)
				}
			}
		}

		http.Redirect(w, r, "/post?id="+postID, http.StatusFound)
	}
}
