package handlers

import (
	"fmt"
	"forum/database"
	"forum/utils"
	"html/template"
	"net/http"
	"strconv"
	"time"
)

// CreatePostHandler handles GET and POST for /create_post
func CreatePostHandler(w http.ResponseWriter, r *http.Request) {
	userID, _ := utils.GetCurrentUser(r)

	if r.Method == http.MethodGet {
		// Show the form with categories
		cats, err := getAllCategories()
		if err != nil {
			utils.HandleError(w, 500, "Database Error", "Failed to load categories")
			return
		}
		tmpl, err := template.ParseFiles("templates/create_post.html")
		if err != nil {
			utils.HandleError(w, 500, "Template Error", "Failed to load create post template")
			return
		}
		err = tmpl.Execute(w, map[string]interface{}{
			"Categories": cats,
		})
		if err != nil {
			utils.HandleError(w, 500, "Template Error", "Failed to render create post page")
			return
		}
		return
	}

	if r.Method == http.MethodPost {

		title := utils.SanitizeTitle(r.FormValue("title"))
		content := utils.SanitizeHTML(r.FormValue("content"))
		categoryIDs := r.Form["category_id"]

		if title == "" || content == "" || len(categoryIDs) == 0 {
			cats, err := getAllCategories()
			if err != nil {
				utils.HandleError(w, 500, "Database Error", "Failed to load categories")
				return
			}
			tmpl, err := template.ParseFiles("templates/create_post.html")
			if err != nil {
				utils.HandleError(w, 500, "Template Error", "Failed to load create post template")
				return
			}
			errorMsg := "All fields are required."
			if len(categoryIDs) == 0 {
				errorMsg = "Please select at least one category."
			}
			err = tmpl.Execute(w, map[string]interface{}{
				"Error":      errorMsg,
				"Categories": cats,
			})
			if err != nil {
				utils.HandleError(w, 500, "Template Error", "Failed to render create post page")
				return
			}
			return
		}
		if len(title) > 100 || len(content) > 1000 {
			cats, err := getAllCategories()
			if err != nil {
				utils.HandleError(w, 500, "Database Error", "Failed to load categories")
				return
			}
			tmpl, err := template.ParseFiles("templates/create_post.html")
			if err != nil {
				utils.HandleError(w, 500, "Template Error", "Failed to load create post template")
				return
			}
			err = tmpl.Execute(w, map[string]interface{}{
				"Error":      "Title or content too long.",
				"Categories": cats,
			})
			if err != nil {
				utils.HandleError(w, 500, "Template Error", "Failed to render create post page")
				return
			}
			return
		}
		res, err := database.DB.Exec("INSERT INTO posts (user_id, title, content) VALUES (?, ?, ?)", userID, title, content)
		if err != nil {
			cats, err := getAllCategories()
			if err != nil {
				utils.HandleError(w, 500, "Database Error", "Failed to load categories")
				return
			}
			tmpl, err := template.ParseFiles("templates/create_post.html")
			if err != nil {
				utils.HandleError(w, 500, "Template Error", "Failed to load create post template")
				return
			}
			err = tmpl.Execute(w, map[string]interface{}{
				"Error":      "Failed to create post.",
				"Categories": cats,
			})
			if err != nil {
				utils.HandleError(w, 500, "Template Error", "Failed to render create post page")
				return
			}
			return
		}
		postID, _ := res.LastInsertId()

		// Save selected categories
		for _, catID := range categoryIDs {
			_, _ = database.DB.Exec("INSERT INTO post_categories (post_id, category_id) VALUES (?, ?)", postID, catID)
		}

		// Success: redirect to homepage
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Method not allowed
	utils.HandleError(w, 405, "Method Not Allowed", "This endpoint only accepts GET and POST requests")
}

// CommentView is used to display comments on a post
type CommentView struct {
	ID           int
	Content      string
	Author       string
	Created      string
	LikeCount    int
	DislikeCount int
	UserID       int
}

// ViewPostHandler handles GET /post?id=POST_ID
func ViewPostHandler(w http.ResponseWriter, r *http.Request) {
	// Get post ID from query
	idStr := r.URL.Query().Get("id")
	postID, err := strconv.Atoi(idStr)
	if err != nil || postID <= 0 {
		utils.HandleError(w, 400, "Invalid Post ID", "The post ID provided is not valid")
		return
	}

	// Fetch the post
	var postTitle, postContent, postAuthor, postCreated string
	var postUserID int
	err = database.DB.QueryRow(`
		SELECT posts.title, posts.content, users.username, posts.created_at, posts.user_id
		FROM posts
		JOIN users ON posts.user_id = users.id
		WHERE posts.id = ?
	`, postID).Scan(&postTitle, &postContent, &postAuthor, &postCreated, &postUserID)
	if err != nil {
		utils.HandleError(w, 404, "Post Not Found", "The post you're looking for doesn't exist")
		return
	}

	// Format post timestamp
	postTime, err := time.Parse("2006-01-02 15:04:05", postCreated)
	if err != nil {
		postTime, err = time.Parse("2006-01-02 15:04:05.000", postCreated)
	}
	if err != nil {
		postTime, err = time.Parse("2006-01-02T15:04:05Z", postCreated)
	}
	if err != nil {
		postTime, err = time.Parse(time.RFC3339, postCreated)
	}
	if err != nil {
		postTime = time.Now() // fallback
	}
	formattedPostTime := postTime.Format("January 2, 2006 15:04")

	// Fetch comments for the post
	rows, err := database.DB.Query(`
		SELECT comments.id, comments.content, users.username, comments.created_at, comments.user_id
		FROM comments
		JOIN users ON comments.user_id = users.id
		WHERE comments.post_id = ?
		ORDER BY comments.created_at ASC
	`, postID)
	if err != nil {
		utils.HandleError(w, 500, "Database Error", "Failed to load comments")
		return
	}
	defer rows.Close()

	var comments []CommentView
	for rows.Next() {
		var c CommentView
		var commentTimeStr string
		if err := rows.Scan(&c.ID, &c.Content, &c.Author, &commentTimeStr, &c.UserID); err != nil {
			continue
		}

		// Format comment timestamp
		commentTime, err := time.Parse("2006-01-02 15:04:05", commentTimeStr)
		if err != nil {
			commentTime, err = time.Parse("2006-01-02 15:04:05.000", commentTimeStr)
		}
		if err != nil {
			commentTime, err = time.Parse("2006-01-02T15:04:05Z", commentTimeStr)
		}
		if err != nil {
			commentTime, err = time.Parse(time.RFC3339, commentTimeStr)
		}
		if err != nil {
			commentTime = time.Now() // fallback
		}
		c.Created = commentTime.Format("January 2, 2006 15:04")

		// Fetch like and dislike counts for the comment
		likeCount := 0
		dislikeCount := 0
		_ = database.DB.QueryRow("SELECT COUNT(*) FROM likes WHERE comment_id = ? AND is_like = 1", c.ID).Scan(&likeCount)
		_ = database.DB.QueryRow("SELECT COUNT(*) FROM likes WHERE comment_id = ? AND is_like = 0", c.ID).Scan(&dislikeCount)
		c.LikeCount = likeCount
		c.DislikeCount = dislikeCount
		comments = append(comments, c)
	}

	// Check if user is logged in
	userID, username := utils.GetCurrentUser(r)

	// Fetch categories for the post
	catRows, err := database.DB.Query(`SELECT categories.name FROM categories JOIN post_categories ON categories.id = post_categories.category_id WHERE post_categories.post_id = ?`, postID)
	if err != nil {
		utils.HandleError(w, 500, "Database Error", "Failed to load post categories")
		return
	}
	var cats []string
	for catRows.Next() {
		var catName string
		catRows.Scan(&catName)
		cats = append(cats, catName)
	}
	catRows.Close()

	// Render the template
	tmpl, err := template.ParseFiles("templates/post.html")
	if err != nil {
		utils.HandleError(w, 500, "Template Error", "Failed to load post template")
		return
	}

	data := map[string]interface{}{
		"ID":         postID,
		"Title":      postTitle,
		"Content":    postContent,
		"Author":     postAuthor,
		"Created":    formattedPostTime,
		"Comments":   comments,
		"LoggedIn":   userID != 0,
		"Username":   username,
		"UserID":     userID,
		"PostUserID": postUserID,
		"Categories": cats,
	}
	err = tmpl.Execute(w, data)
	if err != nil {
		utils.HandleError(w, 500, "Template Error", "Failed to render post page")
		return
	}
}

// Category represents a forum category
type Category struct {
	ID   int
	Name string
}

// getAllCategories fetches all categories from the database
func getAllCategories() ([]Category, error) {
	rows, err := database.DB.Query("SELECT id, name FROM categories ORDER BY name ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var cats []Category
	for rows.Next() {
		var c Category
		if err := rows.Scan(&c.ID, &c.Name); err != nil {
			continue
		}
		cats = append(cats, c)
	}
	return cats, nil
}

// DeletePostHandler handles POST /delete_post
func DeletePostHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.HandleError(w, 405, "Method Not Allowed", "This endpoint only accepts POST requests")
		return
	}

	userID, _ := utils.GetCurrentUser(r)

	postIDStr := r.FormValue("post_id")
	postID, err := strconv.Atoi(postIDStr)
	if err != nil || postID <= 0 {
		utils.HandleError(w, 400, "Invalid Post ID", "The post ID provided is not valid")
		return
	}

	// Check if the post exists and belongs to the current user
	var postUserID int
	err = database.DB.QueryRow("SELECT user_id FROM posts WHERE id = ?", postID).Scan(&postUserID)
	if err != nil {
		utils.HandleError(w, 404, "Post Not Found", "The post you're trying to delete doesn't exist")
		return
	}

	if postUserID != userID {
		utils.HandleError(w, 403, "Forbidden", "You can only delete your own posts")
		return
	}

	// Delete the post (cascade will handle related data)
	result, err := database.DB.Exec("DELETE FROM posts WHERE id = ?", postID)
	if err != nil {
		utils.HandleError(w, 500, "Database Error", "Failed to delete post")
		return
	}

	// Check how many rows were affected
	rowsAffected, _ := result.RowsAffected()
	fmt.Printf("Deleted post %d, rows affected: %d\n", postID, rowsAffected)

	// Redirect back to homepage
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
