package models

import (
	"database/sql"
	"errors"
	"fmt"
)

// ===== BASICS =====
func InitializeCommunityDB(db *sql.DB) error {
	checkTableSQL := `SELECT count(*) FROM sqlite_master WHERE type='table' AND name='community';`
	var tableExists int
	err := db.QueryRow(checkTableSQL).Scan(&tableExists)
	if err != nil {
		return fmt.Errorf("Erreur lors de la vérification de l'existence de la table community: %v", err)
	}

	if tableExists == 0 {
		createTableSQL := `
            CREATE TABLE community (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                name VARCHAR(255),
                cid VARCHAR(64) UNIQUE
            );
        `
		_, err := db.Exec(createTableSQL)
		if err != nil {
			return fmt.Errorf("Erreur lors de la création de la table community: %v", err)
		}

		COMMUNITIES := []string{
			"Siamois",
			"Persan",
			"Chartreux",
			"Écaille de tortue",
			"Bengal",
		}

		COMMUNITIES_ID := []string{
			"siamois",
			"persan",
			"chartreux",
			"ecailledetortue",
			"bengal",
		}

		insertCommuSQL := `
            INSERT INTO community (name, cid) VALUES (?, ?);
        `
		for i, commu := range COMMUNITIES {
			_, err := db.Exec(insertCommuSQL, commu, COMMUNITIES_ID[i])
			if err != nil {
				return fmt.Errorf("Erreur lors de l'ajout de la commu: %v", err)
			}
		}
	}

	return nil
}

//==========

// ===== STRUCTS =====
type Community struct {
	ID          int
	Name        string
	Cid         string
	PostsNumber int
}

//==========

// ===== Messages d'erreurs =====
var (
	ErrCommuUnknow = errors.New("Communauté inconnue")
)

//==========

// ===== FONCTIONS =====
func GetCommunity(db *sql.DB, id string) (*Community, error) {
	var community *Community

	rowsCommu, err := db.Query("SELECT id, name, cid FROM community WHERE id = ?", id)
	if err != nil {
		return community, fmt.Errorf("Erreur lors de la vérification de l'existence de la commu: %v", err)
	}
	defer rowsCommu.Close()

	for rowsCommu.Next() {
		var id int
		var name string
		var cid string
		err := rowsCommu.Scan(&id, &name, &cid)
		if err != nil {
			return community, fmt.Errorf("Erreur lors de la récupération de la commu: %v", err)
		}

		community = &Community{
			ID:   id,
			Name: name,
			Cid:  cid,
		}

		return community, nil
	}

	return community, ErrCommuUnknow
}

func GetCommunityFromCid(db *sql.DB, cid string) (*Community, error) {
	var community *Community

	rowsCommu, err := db.Query("SELECT id, name, cid FROM community WHERE cid = ?", cid)
	if err != nil {
		return community, fmt.Errorf("Erreur lors de la vérification de l'existence de la commu: %v", err)
	}
	defer rowsCommu.Close()

	for rowsCommu.Next() {
		var id int
		var name string
		var cid string
		err := rowsCommu.Scan(&id, &name, &cid)
		if err != nil {
			return community, fmt.Errorf("Erreur lors de la récupération de la commu: %v", err)
		}

		community = &Community{
			ID:   id,
			Name: name,
			Cid:  cid,
		}

		return community, nil
	}

	return community, ErrCommuUnknow
}

func GetCommunityFromName(db *sql.DB, name string) (*Community, error) {
	var community *Community

	rowsCommu, err := db.Query("SELECT id, name, cid FROM community WHERE name = ?", name)
	if err != nil {
		return community, fmt.Errorf("Erreur lors de la vérification de l'existence de la commu: %v", err)
	}
	defer rowsCommu.Close()

	for rowsCommu.Next() {
		var id int
		var name string
		var cid string
		err := rowsCommu.Scan(&id, &name, &cid)
		if err != nil {
			return community, fmt.Errorf("Erreur lors de la récupération de la commu: %v", err)
		}

		community = &Community{
			ID:   id,
			Name: name,
			Cid:  cid,
		}

		return community, nil
	}

	return community, ErrCommuUnknow
}

func GetAllCommunities(db *sql.DB) ([]*Community, error) {
	communities := []*Community{}

	rowsCommu, err := db.Query("SELECT id, name, cid FROM community")
	if err != nil {
		return communities, fmt.Errorf("Erreur lors de la vérification de l'existence de la communauté: %v", err)
	}
	defer rowsCommu.Close()

	for rowsCommu.Next() {
		var id int
		var name string
		var cid string
		err := rowsCommu.Scan(&id, &name, &cid)
		if err != nil {
			return communities, fmt.Errorf("Erreur lors de la récupération de la communauté: %v", err)
		}

		postNumber, _ := GetAllPostsFromCommunity(db, id, "", []string{})

		community := &Community{
			ID:          id,
			Name:        name,
			Cid:         cid,
			PostsNumber: len(postNumber),
		}

		communities = append(communities, community)
	}

	return communities, nil
}

//==========
