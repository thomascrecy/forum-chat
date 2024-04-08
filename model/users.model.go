package models

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const UsersImageUploadPath = "./uploads/users"
const DefaultProfileImageFilename = "profile.jpg"

// ===== BASICS =====
func InitializeUsersDB(db *sql.DB) error {
	createTableUsersSQL := `
        CREATE TABLE IF NOT EXISTS users (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            username VARCHAR(24) NOT NULL UNIQUE,
            email VARCHAR(320) NOT NULL UNIQUE,
            password VARCHAR(255) NOT NULL,
            salt VARCHAR(16) NOT NULL,
            description VARCHAR(512) DEFAULT "J'aime beaucoup les chats",
            createdAt INT NOT NULL
        );
    `
	_, err := db.Exec(createTableUsersSQL)
	if err != nil {
		return fmt.Errorf("erreur lors de la création de la table des utilisateurs: %v", err)
	}

	return nil
}

//==========

// ===== CONST =====
const USERNAME_MAX_LENGTH = 64
const USERNAME_REGEX = `^[a-zA-Z0-9]+$`

const EMAIL_MAX_LENGTH = 320
const EMAIL_REGEX = `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`

const PASSWORD_MAX_LENGTH = 255
const PASSWORD_REGEX = `^[A-Za-z\d!@#$&*./]{8,}$`

//==========

// ===== STRUCTS =====
type User struct {
	ID                 int
	Username           string
	Description        string
	CreatedAt          int
	CreatedAtFormatted string
	HasProfilePicture  bool
}

//==========

//===== UTILS =====

// Permet de récupérer le timestamp actuel
func getCurrentTimestamp() int64 {
	location, err := time.LoadLocation("Europe/Paris")
	if err != nil {
		fmt.Println("Erreur lors du chargement du fuseau horaire:", err)
		return 0
	}

	timeInParis := time.Now().In(location)
	timestamp := timeInParis.Unix()

	return timestamp
}

// Permet de hash un string en sha-256 via un salt
func hashPassword(salt string, password string) string {
	hasher := sha256.New()
	hasher.Write([]byte(salt + password))
	hashedBytes := hasher.Sum(nil)
	hashedString := fmt.Sprintf("%x", hashedBytes)
	return hashedString
}

// Permet de générer un salt
func generateSalt() (string, error) {
	salt := make([]byte, 16)
	_, err := rand.Read(salt)
	if err != nil {
		return "", err
	}
	saltString := hex.EncodeToString(salt)
	return saltString, nil
}

// Permet de vérifier si un username est valide par rapport à son regex
func ValidateUsername(username string) error {
	if username == "" || len(username) > USERNAME_MAX_LENGTH {
		return fmt.Errorf("Le nom d'utilisateur doit faire entre 1 et 24 caractères")
	}
	usernameRegex, err := regexp.Compile(USERNAME_REGEX)
	if err != nil {
		return fmt.Errorf("Erreur de validation du nom d'utilisateur: %v", err)
	}
	if !usernameRegex.MatchString(username) {
		return fmt.Errorf("Le nom d'utilisateur ne doit contenir que des lettres et des chiffres")
	}
	return nil
}

// Permet de vérifier si une adresse email est valide par rapport à son regex
func ValidateEmail(email string) error {
	email = strings.ToLower(email)
	if email == "" {
		return fmt.Errorf("Une adresse email doit être spécifiée")
	}
	if len(email) > EMAIL_MAX_LENGTH {
		return fmt.Errorf("L'adresse email doit faire moins de 320 caractères")
	}
	emailRegex, err := regexp.Compile(EMAIL_REGEX)
	if err != nil {
		return fmt.Errorf("Erreur de validation de l'adresse email: %v", err)
	}
	if !emailRegex.MatchString(email) {
		return fmt.Errorf("L'adresse email doit être valide")
	}
	return nil
}

// Permet de vérifier si un mot de passe est valide par rapprot à son regex
func ValidatePassword(password string) error {
	if password == "" {
		return fmt.Errorf("Un mot de passe doit être spécifié")
	}
	if len(password) > PASSWORD_MAX_LENGTH {
		return fmt.Errorf("Le mot de passe ne doit pas dépasser les 255 caractères")
	}

	hasUppercase := strings.ContainsAny(password, "ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	hasLowercase := strings.ContainsAny(password, "abcdefghijklmnopqrstuvwxyz")
	hasDigit := strings.ContainsAny(password, "0123456789")
	hasSpecial := strings.ContainsAny(password, "!@#$&*./")

	if !hasUppercase || !hasLowercase || !hasDigit || !hasSpecial {
		return fmt.Errorf("Le mot de passe doit avoir au moins 1 majuscule, 1 minuscule, 1 chiffre et 1 caractère spécial (!@#$&*./)")
	}
	passwordRegex := regexp.MustCompile(PASSWORD_REGEX)
	if !passwordRegex.MatchString(password) {
		return fmt.Errorf("Le mot de passe doit avoir au moins 1 majuscule, 1 minuscule, 1 chiffre et 1 caractère spécial (!@#$&*./)")
	}

	return nil
}

//==========

// ===== Messages d'erreurs =====
var (
	ErrInvalidUsername = errors.New("Le nom d'utilisateur doit faire entre 1 et 64 caractères et ne doit contenir que des lettres et des chiffres")
	ErrInvalidEmail    = errors.New("L'adresse email doit être valide")
	ErrInvalidPassword = errors.New("Le mot de passe doit avoir au moins 1 majuscule, 1 minuscule, 1 chiffre et 1 caractère spécial")
	ErrUserExists      = errors.New("Ce nom d'utilisateur ou cette adresse email est déjà utilisé(e)")

	ErrInvalidAuth = errors.New("Erreur d'authentification")
)

//==========

// ===== FONCTIONS =====
func GetUser(db *sql.DB, id int) (*User, error) {

	rowsUser, err := db.Query("SELECT id, username, description, createdAt FROM users WHERE id = ?", id)
	if err != nil {
		return nil, fmt.Errorf("Erreur lors de la vérification de l'existence de l'utilisateur: %v", err)
	}
	defer rowsUser.Close()

	if rowsUser.Next() {
		var id int
		var username string
		var description string
		var createdAt int

		err := rowsUser.Scan(&id, &username, &description, &createdAt)
		if err != nil {
			return nil, fmt.Errorf("Erreur lors de la récupération de l'utilisateur: %v", err)
		} else {
			CreatedAtFormatted := time.Unix(int64(createdAt), 0).Format("15:04 - 02/01/2006")

			file, profilePictureExistErr := os.Open(UsersImageUploadPath + "/" + strconv.Itoa(id) + "/" + DefaultProfileImageFilename)
			defer file.Close()

			user := &User{
				ID:                 id,
				Username:           username,
				Description:        description,
				CreatedAt:          createdAt,
				CreatedAtFormatted: CreatedAtFormatted,
				HasProfilePicture:  (profilePictureExistErr == nil),
			}

			return user, nil
		}
	}
	return nil, nil
}

// Fonction pour enregister un utilisateur dans la base de donnée
func RegisterUser(db *sql.DB, username, email, password string) error {
	if err := ValidateUsername(username); err != nil {
		return ErrInvalidUsername
	}
	if err := ValidateEmail(email); err != nil {
		return ErrInvalidEmail
	}
	if err := ValidatePassword(password); err != nil {
		return ErrInvalidPassword
	}

	var count int
	row := db.QueryRow("SELECT COUNT(*) FROM users WHERE username = ? OR email = ?", username, email)
	if err := row.Scan(&count); err != nil {
		return fmt.Errorf("Erreur lors de la vérification de l'existence de l'utilisateur: %v", err)
	}
	if count > 0 {
		return ErrUserExists
	}

	salt, err := generateSalt()
	if err != nil {
		return fmt.Errorf("Erreur lors de la génération du sel: %v", err)
	}
	hashedPassword := hashPassword(salt, password)

	_, err = db.Exec("INSERT INTO users (username, email, password, salt, createdAt) VALUES (?, ?, ?, ?, ?)",
		username, email, hashedPassword, salt, getCurrentTimestamp())
	if err != nil {
		return fmt.Errorf("Erreur lors de l'insertion de l'utilisateur dans la base de données: %v", err)
	}

	return nil
}

// Fonction pour vérifier le login d'un utilisateur
func LoginUser(db *sql.DB, email, password string) (int, error) {
	email = strings.ToLower(email)

	userId := -1

	if err := ValidateEmail(email); err != nil {
		return userId, ErrInvalidEmail
	}

	rowsEmail, err := db.Query("SELECT id, salt, password FROM users WHERE email = ?", email)
	if err != nil {
		return userId, fmt.Errorf("erreur lors de la vérification de l'existence de l'utilisateur: %v", err)
	}
	defer rowsEmail.Close()

	if rowsEmail.Next() {
		var salt string
		var goodPassword string
		err := rowsEmail.Scan(&userId, &salt, &goodPassword)
		if err != nil {
			return userId, fmt.Errorf("erreur lors de la récupération des données de l'utilisateur: %v", err)
		} else {
			hashedPassword := hashPassword(salt, password)
			if hashedPassword != goodPassword {
				return userId, ErrInvalidAuth
			}
		}
	} else {
		return userId, ErrInvalidAuth
	}

	return userId, nil
}

func EditUserBio(db *sql.DB, userID int, newBio string) error {
	_, err := GetUser(db, userID)
	if err != nil {
		return fmt.Errorf("erreur lors de la vérification de l'existence de l'utilisateur: %v", err)
	}

	_, err = db.Exec("UPDATE users SET description = ? WHERE id = ?", newBio, userID)
	if err != nil {
		return fmt.Errorf("erreur lors de la mise à jour de la description de l'utilisateur: %v", err)
	}

	return nil
}

func EditUserName(db *sql.DB, userID int, newName string) error {
	_, err := GetUser(db, userID)
	if err != nil {
		return fmt.Errorf("erreur lors de la vérification de l'existence de l'utilisateur: %v", err)
	}

	_, err = db.Exec("UPDATE users SET username = ? WHERE id = ?", newName, userID)
	if err != nil {
		return fmt.Errorf("erreur lors de la mise à jour du nom de l'utilisateur: %v", err)
	}

	return nil
}

//==========
