package models

import (
    "database/sql"
    "fmt"
    "errors"
)

//===== BASICS =====
func InitializeTagsDB(db *sql.DB) error {
    checkTableSQL := `SELECT count(*) FROM sqlite_master WHERE type='table' AND name='tag';`
    var tableExists int
    err := db.QueryRow(checkTableSQL).Scan(&tableExists)
    if err != nil {
        return fmt.Errorf("Erreur lors de la vérification de l'existence de la table tag: %v", err)
    }

    if tableExists == 0 {
        createTableSQL := `
            CREATE TABLE tag (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                name VARCHAR(64),
                color VARCHAR(6)
            );
        `
        _, err := db.Exec(createTableSQL)
        if err != nil {
            return fmt.Errorf("Erreur lors de la création de la table tag: %v", err)
        }

        TAGS := []string{
            "Mignon",
            "Moche",
            "Noir",
            "Chauve",
            "Louche",
            "Adoption",
            "Viande",
        }

        TAGS_COLOR := []string{
            "be508a",
            "f0361b",
            "5b73fe",
            "fe0e69",
            "c24eeb",
            "fdd563",
            "5ca431",
        }

        insertTagSQL := `
            INSERT INTO tag (name, color) VALUES (?, ?);
        `
        for i, tag := range TAGS {
            _, err := db.Exec(insertTagSQL, tag, TAGS_COLOR[i])
            if err != nil {
                return fmt.Errorf("Erreur lors de l'ajout du tag: %v", err)
            }
        }
    }

    return nil
}

//==========

//===== STRUCTS =====
type Tag struct {
    ID        int
    Name string
    Color string
}
//==========

//===== Messages d'erreurs =====
var (
    ErrTagUnknow = errors.New("Tag inconnu")
)
//==========

//===== FONCTIONS =====
func GetTag(db *sql.DB, id string) (*Tag, error) {
    var tag *Tag

    rowsTag, err := db.Query("SELECT id, name, color FROM tag WHERE id = ?", id)
    if err != nil {
        return tag, fmt.Errorf("Erreur lors de la vérification de l'existence du tag: %v", err)
    }
    defer rowsTag.Close()

    for rowsTag.Next() {
        var id int
        var name string
        var color string
        err := rowsTag.Scan(&id, &name, &color)
        if err != nil {
            return tag, fmt.Errorf("Erreur lors de la récupération du tag: %v", err)
        }

        tag = &Tag{
            ID: id,
            Name: name,
            Color: color,
        }

        return tag, nil
    }

    return tag, ErrTagUnknow
}

func GetAllTags(db *sql.DB) ([]*Tag, error) {
    tags := []*Tag{}

    rowsTag, err := db.Query("SELECT id, name, color FROM tag")
    if err != nil {
        return tags, fmt.Errorf("Erreur lors de la vérification de l'existence du tag: %v", err)
    }
    defer rowsTag.Close()

    for rowsTag.Next() {
        var id int
        var name string
        var color string
        err := rowsTag.Scan(&id, &name, &color)
        if err != nil {
            return tags, fmt.Errorf("Erreur lors de la récupération du tag: %v", err)
        }

        tag := &Tag{
            ID: id,
            Name: name,
            Color: color,
        }

        tags=append(tags, tag)
    }

    return tags, nil
}
//==========