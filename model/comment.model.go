package models

import (
	"database/sql"
	"fmt"
	"sort"
	"time"
)

// ===== BASICS =====
func InitializeCommentDB(db *sql.DB) error {
	createTableCommentSQL := `
        CREATE TABLE IF NOT EXISTS comments (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            author INTEGER,
            content TEXT NOT NULL,
            post INTEGER NOT NULL,
            createdAt INTEGER NOT NULL
        );
    `
	_, err := db.Exec(createTableCommentSQL)
	if err != nil {
		return fmt.Errorf("Erreur lors de la création de la table des commentaires: %v", err)
	}

	return nil
}

//==========

// ===== STRUCTS =====
type Comment struct {
	ID                 int
	Content            string
	Author             int
	AuthorData         *User
	Post               int
	CreatedAt          int
	CreatedAtFormatted string
	LikesNumber        int
	Liked              bool
}

//==========

//===== FONCTIONS =====

// Fonction pour poster un commentaire
func PostComment(db *sql.DB, author int, content string, post int) error {
	insertCommentSQL := `
        INSERT INTO comments (author, content, post, createdAt) VALUES (?, ?, ?, ?);
    `
	_, err := db.Exec(insertCommentSQL, author, content, post, getCurrentTimestamp())
	if err != nil {
		return fmt.Errorf("Erreur lors de la publication du commentaire: %v", err)
	}

	return nil
}

func GetCountCommentsOfPost(db *sql.DB, postId int) (int, error) {
	var counter int
	err := db.QueryRow("SELECT COUNT(*) FROM comments WHERE post = ?", postId).Scan(&counter)
	if err != nil {
		return 0, fmt.Errorf("Erreur lors de la vérification de l'existence de la table community: %v", err)
	}

	return counter, nil
}

func GetAllCommentsOfPost(db *sql.DB, postId string, currentUserId int) ([]*Comment, error) {
	var comments []*Comment

	rowsComment, err := db.Query("SELECT id, author, content, post, createdAt FROM comments WHERE post = ?", postId)
	if err != nil {
		return nil, fmt.Errorf("Erreur lors de la récupération des commentaires : %v", err)
	}
	defer rowsComment.Close()

	for rowsComment.Next() {
		var comment Comment

		err := rowsComment.Scan(&comment.ID, &comment.Author, &comment.Content, &comment.Post, &comment.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("Erreur lors de la récupération du commentaire : %v", err)
		}

		comment.AuthorData, err = GetUser(db, comment.Author)
		if err != nil {
			fmt.Println(err)
			comment.AuthorData = &User{}
		}

		comment.LikesNumber, err = GetCountLikesOfPost(db, comment.ID, true)
		if err != nil {
			comment.LikesNumber = 0
		}

		if HasUserLikedPost(db, comment.ID, currentUserId, true) {
			comment.Liked = true
		}

		timeObj := time.Unix(int64(comment.CreatedAt), 0)
		comment.CreatedAtFormatted = timeObj.Format("02/01 15:04")

		comments = append(comments, &comment)
	}

	sort.Slice(comments, func(i, j int) bool {
		return comments[i].CreatedAt > comments[j].CreatedAt
	})

	sort.Slice(comments, func(i, j int) bool {
		return comments[i].LikesNumber > comments[j].LikesNumber
	})

	return comments, nil
}

//==========
