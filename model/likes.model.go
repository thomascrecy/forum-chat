package models

import (
	"database/sql"
	"fmt"
)

// ===== BASICS =====
func InitializeLikesDB(db *sql.DB) error {
	checkTableSQL := `SELECT count(*) FROM sqlite_master WHERE type='table' AND name='likes';`
	var tableExists int
	err := db.QueryRow(checkTableSQL).Scan(&tableExists)
	if err != nil {
		return fmt.Errorf("Erreur lors de la vérification de l'existence de la table likes: %v", err)
	}

	if tableExists == 0 {
		createTableSQL := `
            CREATE TABLE likes (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                link INTEGER NOT NULL,
                isComment BOOLEAN DEFAULT False,
                author INTEGER NOT NULL
            );
        `
		_, err := db.Exec(createTableSQL)
		if err != nil {
			return fmt.Errorf("Erreur lors de la création de la table likes: %v", err)
		}
	}

	return nil
}

//==========

// ===== STRUCTS =====
type Like struct {
	ID        int
	Link      int
	IsComment bool
	Author    int
}

//==========

//===== Messages d'erreurs =====
//==========

// ===== FONCTIONS =====
// Fonction pour envoyer un like
func SendLike(db *sql.DB, linkId int, authorId int, isComment bool) (bool, error) {
	if HasUserLikedPost(db, linkId, authorId, isComment) {
		// Supprimer le like
		deleteLikeSQL := `
            DELETE FROM likes WHERE link = ? AND author = ? AND isComment = ?;
        `
		_, err := db.Exec(deleteLikeSQL, linkId, authorId, isComment)
		if err != nil {
			return false, fmt.Errorf("Erreur lors de la suppression du like: %v", err)
		}

		return false, nil
	} else {
		// Mettre le like
		insertLikeSQL := `
            INSERT INTO likes (link, author, isComment) VALUES (?, ?, ?);
        `
		_, err := db.Exec(insertLikeSQL, linkId, authorId, isComment)
		if err != nil {
			return false, fmt.Errorf("Erreur lors de l'ajout du like: %v", err)
		}

		return true, nil
	}
}

func HasUserLikedPost(db *sql.DB, linkId int, authorId int, isComment bool) bool {
	if authorId == -1 {
		return false
	}

	var counter int
	err := db.QueryRow("SELECT COUNT(*) FROM likes WHERE link = ? AND author = ? AND isComment = ?", linkId, authorId, isComment).Scan(&counter)
	if err != nil {
		return false
	}

	return counter != 0
}

func GetCountLikesOfPost(db *sql.DB, linkId int, isComment bool) (int, error) {
	var counter int
	err := db.QueryRow("SELECT COUNT(*) FROM likes WHERE link = ? AND isComment = ?", linkId, isComment).Scan(&counter)
	if err != nil {
		return 0, fmt.Errorf("Erreur lors de la vérification de l'existence de la table likes: %v", err)
	}

	return counter, nil
}

//==========
