package models

import (
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

// ===== BASICS =====
func InitializePostsDB(db *sql.DB) error {
	createTablePostsSQL := `
        CREATE TABLE IF NOT EXISTS posts (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            community INTEGER NOT NULL,
            title VARCHAR(256) NOT NULL,
            author INTEGER,
            content TEXT NOT NULL,
            tags VARCHAR(255) NOT NULL,
            createdAt INT NOT NULL
        );
    `
	_, err := db.Exec(createTablePostsSQL)
	if err != nil {
		return fmt.Errorf("Erreur lors de la création de la table des posts: %v", err)
	}

	return nil
}

//==========

// ===== STRUCTS =====
type Post struct {
	ID                  int
	Community           *Community
	Title               string
	Author              int
	AuthorData          *User
	Content             string
	Tags                []*Tag
	CreatedAt           int
	CreatedAtFormatted  string
	CommentsLength      int
	LikesNumber         int
	Liked               bool
	IsCurrentUserAuthor bool
}

//==========

// ===== Messages d'erreurs =====
var (
	ErrPostUnknow     = errors.New("Post inconnu")
	ErrTooShortTitle  = errors.New("Le titre est trop court")
	ErrTooLongTitle   = errors.New("Le titre est trop long (max 256 caractères)")
	ErrInvalidContent = errors.New("Le contenu est invalide")
	ErrTooLongContent = errors.New("Le contenu est trop long (max 4000 caractères)")
)

//==========

//===== FONCTIONS =====

// Fonction pour poster un post
func PostPost(db *sql.DB, community string, title string, author int, content string, tags []string) (int, error) {
	defaultPostId := -1

	if len(title) == 0 {
		return defaultPostId, ErrTooShortTitle
	}
	if len(title) >= 256 {
		return defaultPostId, ErrTooLongTitle
	}
	if len(content) == 0 {
		return defaultPostId, ErrInvalidContent
	}
	if len(content) > 4000 {
		return defaultPostId, ErrTooLongContent
	}

	communityData, err := GetCommunityFromName(db, community)
	if err != nil {
		return defaultPostId, ErrCommuUnknow
	}

	tagsStringified := strings.Join(tags, ",")

	insertCommentSQL := `
        INSERT INTO posts (community, title, author, content, tags, createdAt) VALUES (?, ?, ?, ?, ?, ?);
    `
	result, err := db.Exec(insertCommentSQL, communityData.ID, title, author, content, tagsStringified, getCurrentTimestamp())
	if err != nil {
		return defaultPostId, fmt.Errorf("Erreur lors de la publication du post: %v", err)
	}

	postId, err := result.LastInsertId()
	if err != nil {
		return defaultPostId, fmt.Errorf("Erreur lors de la récupération de l'ID du post: %v", err)
	}

	return int(postId), nil
}

func GetPost(db *sql.DB, id string, currentUserId int) (*Post, error) {
	rowsPost, err := db.Query("SELECT id, community, title, author, content, tags, createdAt FROM posts WHERE id = ?", id)
	if err != nil {
		return nil, fmt.Errorf("Erreur lors de la vérification de l'existence du post: %v", err)
	}
	defer rowsPost.Close()

	if rowsPost.Next() {
		var id int
		var community int
		var title string
		var author int
		var content string
		var tags string
		var createdAt int

		err := rowsPost.Scan(&id, &community, &title, &author, &content, &tags, &createdAt)
		if err != nil {
			return nil, fmt.Errorf("Erreur lors de la récupération du post: %v", err)
		} else {
			timeObj := time.Unix(int64(createdAt), 0)
			formattedDateTime := timeObj.Format("02/01 15:04")

			tagsArray := strings.Split(tags, ",")
			tagsArrayFormatted := []*Tag{}
			for _, tag := range tagsArray {
				tagFormatted, err := GetTag(db, tag)
				if err == nil {
					tagsArrayFormatted = append(tagsArrayFormatted, tagFormatted)
				}
			}

			communityData, err := GetCommunity(db, strconv.Itoa(community))
			if err != nil {
				communityData = nil
			}

			commentsLength, err := GetCountCommentsOfPost(db, id)
			if err != nil {
				commentsLength = 0
			}

			likesNumber, err := GetCountLikesOfPost(db, id, false)
			if err != nil {
				likesNumber = 0
			}

			post := &Post{
				ID:                  id,
				Community:           communityData,
				Title:               title,
				Author:              author,
				Content:             content,
				Tags:                tagsArrayFormatted,
				CreatedAt:           createdAt,
				CreatedAtFormatted:  formattedDateTime,
				IsCurrentUserAuthor: author == currentUserId,
				LikesNumber:         likesNumber,
				CommentsLength:      commentsLength,
			}

			return post, nil
		}
	}
	return nil, ErrPostUnknow
}

func GetAllPosts(db *sql.DB, filter string, tagFilters []string, communityId string) ([]*Post, error) {
	posts := []*Post{}

	var rowsPost *sql.Rows
	var err error

	if communityId != "" {
		rowsPost, err = db.Query("SELECT id, community, title, author, tags, content, createdAt FROM posts WHERE community = ?", communityId)
	} else {
		rowsPost, err = db.Query("SELECT id, community, title, author, tags, content, createdAt FROM posts")
	}
	if err != nil {
		return posts, fmt.Errorf("Error while checking post existence: %v", err)
	}
	defer rowsPost.Close()

	for rowsPost.Next() {
		var id int
		var community int
		var title string
		var author int
		var tags string
		var content string
		var createdAt int
		err := rowsPost.Scan(&id, &community, &title, &author, &tags, &content, &createdAt)
		if err != nil {
			return posts, fmt.Errorf("Error while retrieving post: %v", err)
		}
		timeObj := time.Unix(int64(createdAt), 0)
		formattedDateTime := timeObj.Format("02/01 15:04")

		tagsArray := strings.Split(tags, ",")
		tagsArrayFormatted := []*Tag{}

		for _, tag := range tagsArray {
			tagFormatted, err := GetTag(db, tag)
			if err == nil {
				tagsArrayFormatted = append(tagsArrayFormatted, tagFormatted)
			}
		}

		post := &Post{
			Tags: tagsArrayFormatted,
		}

		if !postContainsAllTags(post, tagFilters) {
			continue
		}

		authorData, err := GetUser(db, author)
		if err != nil {
			fmt.Println(err)
			authorData = &User{}
		}

		if len(content) > 50 {
			content = content[:50] + "..."
		}

		commentsLength, err := GetCountCommentsOfPost(db, id)
		if err != nil {
			commentsLength = 0
		}

		likesNumber, err := GetCountLikesOfPost(db, id, false)
		if err != nil {
			likesNumber = 0
		}

		communityData, err := GetCommunity(db, strconv.Itoa(community))
		if err != nil {
			communityData = nil
		}

		post = &Post{
			ID:                 id,
			Community:          communityData,
			Title:              title,
			Author:             author,
			AuthorData:         authorData,
			Content:            content,
			Tags:               tagsArrayFormatted,
			CreatedAt:          createdAt,
			CreatedAtFormatted: formattedDateTime,
			CommentsLength:     commentsLength,
			LikesNumber:        likesNumber,
		}

		posts = append(posts, post)
	}

	if filter == "" || filter == "date-up" {
		sort.Slice(posts, func(i, j int) bool {
			return posts[i].CreatedAt > posts[j].CreatedAt
		})

	} else if filter == "date-down" {
		sort.Slice(posts, func(i, j int) bool {
			return posts[i].CreatedAt < posts[j].CreatedAt
		})

	} else if filter == "likes" {
		sort.Slice(posts, func(i, j int) bool {
			return posts[i].LikesNumber > posts[j].LikesNumber
		})

	} else if filter == "comments" {
		sort.Slice(posts, func(i, j int) bool {
			return posts[i].CommentsLength > posts[j].CommentsLength
		})
	}

	return posts, nil
}

func postContainsAllTags(post *Post, tags []string) bool {
	postTags := make(map[string]bool)
	for _, tag := range post.Tags {
		postTags[tag.Name] = true
	}

	for _, tag := range tags {
		if !postTags[tag] {
			return false
		}
	}

	return true
}

func GetAllPostsFromCommunity(db *sql.DB, communityId int, filter string, tagFilters []string) ([]*Post, error) {
	posts := []*Post{}

	rowsPost, err := db.Query("SELECT id, title, author, tags, content, createdAt FROM posts WHERE community = ?", communityId)
	if err != nil {
		return posts, fmt.Errorf("Error while checking post existence: %v", err)
	}
	defer rowsPost.Close()

	for rowsPost.Next() {
		var id int
		var title string
		var author int
		var tags string
		var content string
		var createdAt int
		err := rowsPost.Scan(&id, &title, &author, &tags, &content, &createdAt)
		if err != nil {
			return posts, fmt.Errorf("Error while retrieving post: %v", err)
		}
		timeObj := time.Unix(int64(createdAt), 0)
		formattedDateTime := timeObj.Format("02/01 15:04")

		tagsArray := strings.Split(tags, ",")
		tagsArrayFormatted := []*Tag{}
		for _, tag := range tagsArray {
			tagFormatted, err := GetTag(db, tag)
			if err == nil {
				tagsArrayFormatted = append(tagsArrayFormatted, tagFormatted)
			}
		}

		post := &Post{
			Tags: tagsArrayFormatted,
		}

		if !postContainsAllTags(post, tagFilters) {
			continue
		}

		authorData, err := GetUser(db, author)
		if err != nil {
			authorData = nil
		}

		if len(content) > 50 {
			content = content[:50] + "..."
		}

		commentsLength, err := GetCountCommentsOfPost(db, id)
		if err != nil {
			commentsLength = 0
		}

		likesNumber, err := GetCountLikesOfPost(db, id, false)
		if err != nil {
			likesNumber = 0
		}

		post = &Post{
			ID:                 id,
			Title:              title,
			Author:             author,
			AuthorData:         authorData,
			Content:            content,
			Tags:               tagsArrayFormatted,
			CreatedAt:          createdAt,
			CreatedAtFormatted: formattedDateTime,
			CommentsLength:     commentsLength,
			LikesNumber:        likesNumber,
		}

		posts = append(posts, post)
	}

	if filter == "" || filter == "date-up" {
		sort.Slice(posts, func(i, j int) bool {
			return posts[i].CreatedAt > posts[j].CreatedAt
		})

	} else if filter == "date-down" {
		sort.Slice(posts, func(i, j int) bool {
			return posts[i].CreatedAt < posts[j].CreatedAt
		})

	} else if filter == "likes" {
		sort.Slice(posts, func(i, j int) bool {
			return posts[i].LikesNumber > posts[j].LikesNumber
		})

	} else if filter == "comments" {
		sort.Slice(posts, func(i, j int) bool {
			return posts[i].CommentsLength > posts[j].CommentsLength
		})
	}

	return posts, nil
}

//==========

func GetAllPostsFromAuthor(db *sql.DB, authorID int) ([]*Post, error) {
	posts := []*Post{}

	rowsPost, err := db.Query("SELECT id, community, title, author, tags, content, createdAt FROM posts WHERE author = ?", authorID)

	if err != nil {
		return posts, fmt.Errorf("Error while checking post existence: %v", err)
	}
	defer rowsPost.Close()

	for rowsPost.Next() {
		var id int
		var community int
		var title string
		var author int
		var tags string
		var content string
		var createdAt int
		err := rowsPost.Scan(&id, &community, &title, &author, &tags, &content, &createdAt)
		if err != nil {
			return posts, fmt.Errorf("Error while retrieving post: %v", err)
		}
		timeObj := time.Unix(int64(createdAt), 0)
		formattedDateTime := timeObj.Format("02/01 15:04")

		tagsArray := strings.Split(tags, ",")
		tagsArrayFormatted := []*Tag{}

		for _, tag := range tagsArray {
			tagFormatted, err := GetTag(db, tag)
			if err == nil {
				tagsArrayFormatted = append(tagsArrayFormatted, tagFormatted)
			}
		}

		post := &Post{
			Tags: tagsArrayFormatted,
		}

		authorData, err := GetUser(db, author)
		if err != nil {
			fmt.Println(err)
			authorData = &User{}
		}

		if len(content) > 50 {
			content = content[:50] + "..."
		}

		commentsLength, err := GetCountCommentsOfPost(db, id)
		if err != nil {
			commentsLength = 0
		}

		likesNumber, err := GetCountLikesOfPost(db, id, false)
		if err != nil {
			likesNumber = 0
		}

		communityData, err := GetCommunity(db, strconv.Itoa(community))
		if err != nil {
			communityData = nil
		}

		post = &Post{
			ID:                 id,
			Community:          communityData,
			Title:              title,
			Author:             author,
			AuthorData:         authorData,
			Content:            content,
			Tags:               tagsArrayFormatted,
			CreatedAt:          createdAt,
			CreatedAtFormatted: formattedDateTime,
			CommentsLength:     commentsLength,
			LikesNumber:        likesNumber,
		}

		posts = append(posts, post)
	}

	return posts, nil
}

func EditPostTitle(db *sql.DB, userID int, postID string, newTitle string) error {
	_, err := GetPost(db, postID, -1)
	if err != nil {
		return fmt.Errorf("erreur lors de la vérification de l'existence de l'utilisateur: %v", err)
	}

	_, err = db.Exec("UPDATE posts SET title = ? WHERE id = ? AND author = ?", newTitle, postID, userID)
	if err != nil {
		return fmt.Errorf("erreur lors de la mise à jour de la description de l'utilisateur: %v", err)
	}

	return nil
}

func EditPostContent(db *sql.DB, userID int, postID string, newcontent string) error {
	_, err := GetPost(db, postID, -1)
	if err != nil {
		return fmt.Errorf("erreur lors de la vérification de l'existence de l'utilisateur: %v", err)
	}

	_, err = db.Exec("UPDATE posts SET content = ? WHERE id = ? AND author = ?", newcontent, postID, userID)
	if err != nil {
		return fmt.Errorf("erreur lors de la mise à jour de la description de l'utilisateur: %v", err)
	}

	return nil
}
