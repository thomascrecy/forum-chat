package controllers

import (
	"database/sql"
	"fmt"
	models "forum/model"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

// ===== STRUCTS =====
type CommunityPage struct {
	Title       string
	Posts       []*models.Post
	Communities []*models.Community
	Tags        []*models.Tag
	IsAuth      bool
	CurrentUser *models.User
}

//===================

func CommunityHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		filter := r.URL.Query().Get("filter")
		tagFilters := r.URL.Query()["tag"]

		p := &CommunityPage{}

		URLParts := strings.Split(r.URL.Path, "/")
		communityCidUrl := strings.ToLower(URLParts[len(URLParts)-1])

		if communityCidUrl != " " {
			communityData, err := models.GetCommunityFromCid(db, communityCidUrl)
			if err != nil {
				fmt.Println(err)
				http.Redirect(w, r, "/404", http.StatusFound)
				return
			}

			posts, err := models.GetAllPosts(db, filter, tagFilters, strconv.Itoa(communityData.ID))
			if err != nil {
				posts = []*models.Post{}
			}

			currentUser := &models.User{}

			sessionCookie, sessionErr := r.Cookie("session")
			if sessionErr != nil {
				sessionCookie = &http.Cookie{Value: ""}
			} else {
				currentUserID, err := GetUserIDFromCookie(sessionCookie)
				if err != nil {
					currentUser = &models.User{}
				} else {
					currentUser, err = models.GetUser(db, currentUserID)
					if err != nil {
						currentUser = &models.User{}
					}

					for _, p := range posts {
						if models.HasUserLikedPost(db, p.ID, currentUserID, false) {
							p.Liked = true
						}
					}
				}
			}

			p.Tags, err = models.GetAllTags(db)
			if err != nil {
				http.Error(w, "Erreur lors de la récupération des tags", http.StatusInternalServerError)
				return
			}

			p.Communities, err = models.GetAllCommunities(db)
			if err != nil {
				http.Error(w, "Erreur lors de la récupération des communautés", http.StatusInternalServerError)
				return
			}

			p.Title = communityData.Name
			p.Posts = posts
			p.IsAuth = (sessionErr == nil)
			p.CurrentUser = currentUser

			t, _ := template.ParseFiles("./view/index.html")
			t.Execute(w, p)

		} else {
			t, _ := template.ParseFiles("./view/index.html")
			t.Execute(w, p)
		}
	}
}
