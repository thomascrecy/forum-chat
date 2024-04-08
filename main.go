package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3"

	controllers "forum/controller"
	models "forum/model"
)

// ===== STRUCTS =====
// Page représente la structure des données passées aux templates HTML
type Page struct {
	Title       string
	Communities []*models.Community
	Posts       []*models.Post
	Tags        []*models.Tag
	IsAuth      bool
	CurrentUser *models.User
}

//===================

//===== Pages HTML =====

// handler pour la page d'accueil
func handler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			NotFound(w, r)
			return
		}

		filter := r.URL.Query().Get("filter")
		tagfilter := r.URL.Query()["tag"]

		posts, err := models.GetAllPosts(db, filter, tagfilter, "")
		if err != nil {
			http.Error(w, "Erreur lors de la récupération des articles", http.StatusInternalServerError)
			return
		}

		currentUser := &models.User{}

		sessionCookie, sessionErr := r.Cookie("session")
		if sessionErr != nil {
			sessionCookie = &http.Cookie{Value: ""}
		} else {
			currentUserID, err := controllers.GetUserIDFromCookie(sessionCookie)
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

		tags, err := models.GetAllTags(db)
		if err != nil {
			http.Error(w, "Erreur lors de la récupération des tags", http.StatusInternalServerError)
			return
		}

		communities, err := models.GetAllCommunities(db)
		if err != nil {
			http.Error(w, "Erreur lors de la récupération des communautés", http.StatusInternalServerError)
			return
		}

		p := &Page{
			Title:       "Accueil",
			Communities: communities,
			Posts:       posts,
			Tags:        tags,
			IsAuth:      (sessionErr == nil),
			CurrentUser: currentUser,
		}
		t, err := template.ParseFiles("./view/index.html")
		if err != nil {
			http.Error(w, "Erreur lors de la lecture du template HTML", http.StatusInternalServerError)
			return
		}
		t.Execute(w, p)
	}
}

// Handler Page 404
func NotFound(w http.ResponseWriter, r *http.Request) {
	p := &Page{
		Title: "Page non trouvée",
	}
	t, err := template.ParseFiles("./view/404.html")
	if err != nil {
		http.Error(w, "Erreur lors de la lecture de la page 404", http.StatusInternalServerError)
		fmt.Println(err)
		return
	}
	t.Execute(w, p)
}

//======================

//===== API =====

func main() {
	db, err := sql.Open("sqlite3", "./forum.db")
	if err != nil {
		fmt.Println("Error opening database:", err)
		return
	}
	defer db.Close()

	//=== Création de la table commentaires ===
	if err := models.InitializeTagsDB(db); err != nil {
		log.Fatal("Erreur lors de l'initialisation de la base de données tags:", err)
	}
	if err := models.InitializeCommentDB(db); err != nil {
		log.Fatal("Erreur lors de l'initialisation de la base de données comments:", err)
	}
	if err := models.InitializeUsersDB(db); err != nil {
		log.Fatal("Erreur lors de l'initialisation de la base de données users:", err)
	}
	if err := models.InitializePostsDB(db); err != nil {
		log.Fatal("Erreur lors de l'initialisation de la base de données posts:", err)
	}
	if err := models.InitializeCommunityDB(db); err != nil {
		log.Fatal("Erreur lors de l'initialisation de la base de données community:", err)
	}
	if err := models.InitializeLikesDB(db); err != nil {
		log.Fatal("Erreur lors de l'initialisation de la base de données likes:", err)
	}
	//======

	//=== Serveur de fichier statiques ===
	fs := http.FileServer(http.Dir("./static/"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	//======

	//=== Routes ===
	http.HandleFunc("/profile", controllers.ProfileHandler(db))
	http.HandleFunc("/profile/edit", controllers.EditProfileHandler(db))

	http.HandleFunc("/login", controllers.LoginHandler(db))
	http.HandleFunc("/register", controllers.RegisterHandler(db))
	http.HandleFunc("/logout", controllers.LogoutHandler)

	http.HandleFunc("/post", controllers.PostHandler(db))
	http.HandleFunc("/post/edit", controllers.EditPostHandler(db))
	http.HandleFunc("/comment", controllers.CommentHandler(db))

	http.HandleFunc("/404", NotFound)

	http.HandleFunc("/race/", controllers.CommunityHandler(db))

	http.HandleFunc("/like", controllers.LikeHandler(db))

	http.HandleFunc("/upload", controllers.UploadUserImageHandler)
	http.HandleFunc("/image/user", controllers.UserImageHandler)

	http.HandleFunc("/", handler(db))
	//======

	fmt.Println("Serveur web lancé sur http://127.0.0.1:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
