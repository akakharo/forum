package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"forum/database"
	"forum/handlers"
	"forum/utils"
	"html/template"
	"time"
)

// PostView is used to display posts on the homepage
type PostView struct {
	ID           int
	Title        string
	Content      string
	Author       string
	Created      time.Time
	LikeCount    int
	DislikeCount int
	Categories   []string
	UserID       int
}

// panicRecovery is a middleware that recovers from panics
func panicRecovery(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Panic recovered: %v", err)
				utils.HandleError(w, 500, "Internal Server Error", "The server encountered an unexpected error")
			}
		}()
		next(w, r)
	}
}

func main() {
	// Get database path from environment variable or use default
	dbPath := "dinoforum.db"
	if envPath := os.Getenv("DB_PATH"); envPath != "" {
		dbPath = envPath
	}

	// Initialize the database (creates database file and tables if needed)
	database.InitDB(dbPath, "database/schema.sql")

	// Insert default categories if none exist
	var count int
	err := database.DB.QueryRow("SELECT COUNT(*) FROM categories").Scan(&count)
	if err == nil && count == 0 {
		defaultCategories := []string{"General", "Fossils", "Dino News", "Questions", "Paleontology", "Dino Art", "Research", "Fun Facts"}
		for _, cat := range defaultCategories {
			_, _ = database.DB.Exec("INSERT INTO categories (name) VALUES (?)", cat)
		}
	}

	// Set up a handler for the root path with panic recovery
	http.HandleFunc("/", panicRecovery(func(w http.ResponseWriter, r *http.Request) {
		userID, username := utils.GetCurrentUser(r)

		// Fetch all categories for the filter UI
		catRows, err := database.DB.Query("SELECT id, name FROM categories ORDER BY name ASC")
		if err != nil {
			utils.HandleError(w, 500, "Database Error", "Failed to load categories")
			return
		}
		var allCategories []struct {
			ID   int
			Name string
		}
		for catRows.Next() {
			var id int
			var name string
			if err := catRows.Scan(&id, &name); err != nil {
				continue
			}
			allCategories = append(allCategories, struct {
				ID   int
				Name string
			}{id, name})
		}
		catRows.Close()

		// Check for category filter
		categoryFilter := r.URL.Query().Get("category_id")
		filter := r.URL.Query().Get("filter")
		var posts []PostView
		var rows *sql.Rows
		if filter == "my" && userID != 0 {
			rows, err = database.DB.Query(`
				SELECT posts.id, posts.title, posts.content, users.username, posts.created_at, posts.user_id
				FROM posts
				JOIN users ON posts.user_id = users.id
				WHERE posts.user_id = ?
				ORDER BY posts.created_at DESC
			`, userID)
		} else if filter == "liked" && userID != 0 {
			rows, err = database.DB.Query(`
				SELECT posts.id, posts.title, posts.content, users.username, posts.created_at, posts.user_id
				FROM posts
				JOIN users ON posts.user_id = users.id
				JOIN likes ON posts.id = likes.post_id
				WHERE likes.user_id = ? AND likes.is_like = 1
				ORDER BY posts.created_at DESC
			`, userID)
		} else if categoryFilter != "" {
			rows, err = database.DB.Query(`
				SELECT posts.id, posts.title, posts.content, users.username, posts.created_at, posts.user_id
				FROM posts
				JOIN users ON posts.user_id = users.id
				JOIN post_categories ON posts.id = post_categories.post_id
				WHERE post_categories.category_id = ?
				ORDER BY posts.created_at DESC
			`, categoryFilter)
		} else {
			rows, err = database.DB.Query(`
				SELECT posts.id, posts.title, posts.content, users.username, posts.created_at, posts.user_id
				FROM posts
				JOIN users ON posts.user_id = users.id
				ORDER BY posts.created_at DESC
			`)
		}
		if err != nil {
			utils.HandleError(w, 500, "Database Error", "Failed to load posts")
			return
		}
		defer rows.Close()

		for rows.Next() {
			var p PostView
			var createdStr string
			if err := rows.Scan(&p.ID, &p.Title, &p.Content, &p.Author, &createdStr, &p.UserID); err != nil {
				continue
			}
			// Try parsing with different formats
			p.Created, err = time.Parse("2006-01-02 15:04:05.000", createdStr)
			if err != nil {
				p.Created, err = time.Parse("2006-01-02 15:04:05", createdStr)
			}
			if err != nil {
				p.Created, err = time.Parse("2006-01-02T15:04:05Z", createdStr)
			}
			if err != nil {
				p.Created, err = time.Parse(time.RFC3339, createdStr)
			}
			if err != nil {
				p.Created = time.Now() // fallback to now if parsing fails
			}
			// Fetch like and dislike counts for the post
			likeCount := 0
			dislikeCount := 0
			_ = database.DB.QueryRow("SELECT COUNT(*) FROM likes WHERE post_id = ? AND is_like = 1", p.ID).Scan(&likeCount)
			_ = database.DB.QueryRow("SELECT COUNT(*) FROM likes WHERE post_id = ? AND is_like = 0", p.ID).Scan(&dislikeCount)
			p.LikeCount = likeCount
			p.DislikeCount = dislikeCount
			// Fetch categories for the post
			catRows, _ := database.DB.Query(`SELECT categories.name FROM categories JOIN post_categories ON categories.id = post_categories.category_id WHERE post_categories.post_id = ?`, p.ID)
			var cats []string
			for catRows.Next() {
				var catName string
				catRows.Scan(&catName)
				cats = append(cats, catName)
			}
			catRows.Close()
			p.Categories = cats
			posts = append(posts, p)
		}

		data := map[string]interface{}{
			"LoggedIn":        userID != 0,
			"Username":        username,
			"UserID":          userID,
			"Posts":           posts,
			"Categories":      allCategories,
			"CurrentCategory": categoryFilter,
			"CurrentFilter":   filter,
		}
		tmpl, err := template.ParseFiles("templates/index.html")
		if err != nil {
			utils.HandleError(w, 500, "Template Error", "Failed to load homepage template")
			return
		}
		err = tmpl.Execute(w, data)
		if err != nil {
			utils.HandleError(w, 500, "Template Error", "Failed to render homepage")
			return
		}
	}))

	// Registration route with panic recovery and guest-only access
	http.HandleFunc("/register", panicRecovery(utils.RequireGuest(handlers.RegisterHandler)))

	// Login route with panic recovery and guest-only access
	http.HandleFunc("/login", panicRecovery(utils.RequireGuest(handlers.LoginHandler)))

	// Logout route with panic recovery
	http.HandleFunc("/logout", panicRecovery(handlers.LogoutHandler))

	// Create Post route with panic recovery and authentication required
	http.HandleFunc("/create_post", panicRecovery(utils.RequireAuth(handlers.CreatePostHandler)))

	// Comment route with panic recovery and authentication required
	http.HandleFunc("/comment", panicRecovery(utils.RequireAuth(handlers.CommentHandler)))

	// View Post route with panic recovery (public access)
	http.HandleFunc("/post", panicRecovery(handlers.ViewPostHandler))

	// Like/Dislike route with panic recovery and authentication required
	http.HandleFunc("/like", panicRecovery(utils.RequireAuth(handlers.LikeHandler)))

	// Delete Post route with panic recovery and authentication required
	http.HandleFunc("/delete_post", panicRecovery(utils.RequireAuth(handlers.DeletePostHandler)))

	// Delete Comment route with panic recovery and authentication required
	http.HandleFunc("/delete_comment", panicRecovery(utils.RequireAuth(handlers.DeleteCommentHandler)))

	// Serve static files (CSS, JS, etc.)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// Custom 404 handler for unknown routes
	http.HandleFunc("/404", func(w http.ResponseWriter, r *http.Request) {
		utils.HandleError(w, 404, "Page Not Found", "The page you're looking for doesn't exist")
	})
	http.HandleFunc("/favicon.ico", http.NotFound)

	// Start the HTTP server on port 8080
	fmt.Println("Server started at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
