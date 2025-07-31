package handlers

import (
	"database/sql"
	"forum/database"
	"html/template"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"

	"golang.org/x/crypto/bcrypt"
)

// RenderTemplate renders an HTML template with data
func RenderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	t, err := template.ParseFiles("templates/" + tmpl)
	if err != nil {
		http.Error(w, "Error loading template", http.StatusInternalServerError)
		return
	}
	t.Execute(w, data)
}

// validateEmail checks if the email format is valid
func validateEmail(email string) bool {
	
	// emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)+\.[a-zA-Z]{2,}$`)

	// Additional checks for common invalid patterns
	if strings.Contains(email, "..") {
		return false
	}
	if strings.HasPrefix(email, ".") || strings.HasSuffix(email, ".") {
		return false
	}
	if strings.Contains(email, "@.") || strings.Contains(email, ".@") {
		return false
	}

	// Check that there's at least one dot in the domain part
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return false
	}
	domain := parts[1]
	if !strings.Contains(domain, ".") {
		return false
	}

	// Check that the domain has at least one character before the first dot
	domainParts := strings.Split(domain, ".")
	if len(domainParts) < 2 || domainParts[0] == "" {
		return false
	}

	return len(email) <= 254
}

// validateUsername checks if username meets requirements
func validateUsername(username string) (bool, string) {
	if len(username) < 3 {
		return false, "Username must be at least 3 characters long."
	}
	if len(username) > 20 {
		return false, "Username must be no more than 20 characters long."
	}
	// Check for valid characters (letters, numbers, underscores, hyphens)
	usernameRegex := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !usernameRegex.MatchString(username) {
		return false, "Username can only contain letters, numbers, underscores, and hyphens."
	}
	return true, ""
}

// validatePassword checks if password meets requirements
func validatePassword(password string) (bool, string) {
	if len(password) < 8 {
		return false, "Password must be at least 8 characters long."
	}
	if len(password) > 50 {
		return false, "Password must be no more than 50 characters long."
	}
	return true, ""
}

// RegisterHandler handles GET and POST for /register
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		// Show the registration form
		RenderTemplate(w, "register.html", nil)
		return
	}

	if r.Method == http.MethodPost {
		email := strings.TrimSpace(r.FormValue("email"))
		username := strings.TrimSpace(r.FormValue("username"))
		password := r.FormValue("password")

		// Basic validation - check for empty fields
		if email == "" || username == "" || password == "" {
			RenderTemplate(w, "register.html", map[string]string{"Error": "All fields are required."})
			return
		}

		// Validate email format
		if !validateEmail(email) {
			RenderTemplate(w, "register.html", map[string]string{"Error": "Please enter a valid email address."})
			return
		}

		// Validate username
		if valid, errMsg := validateUsername(username); !valid {
			RenderTemplate(w, "register.html", map[string]string{"Error": errMsg})
			return
		}

		// Validate password
		if valid, errMsg := validatePassword(password); !valid {
			RenderTemplate(w, "register.html", map[string]string{"Error": errMsg})
			return
		}

		// Check if email or username already exists
		var exists int
		err := database.DB.QueryRow("SELECT COUNT(*) FROM users WHERE email = ? OR username = ?", email, username).Scan(&exists)
		if err != nil {
			RenderTemplate(w, "register.html", map[string]string{"Error": "Database error."})
			return
		}
		if exists > 0 {
			RenderTemplate(w, "register.html", map[string]string{"Error": "Email or username already taken."})
			return
		}

		// Hash the password
		hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			RenderTemplate(w, "register.html", map[string]string{"Error": "Error securing password."})
			return
		}

		// Insert the new user
		_, err = database.DB.Exec("INSERT INTO users (email, username, password_hash) VALUES (?, ?, ?)", email, username, string(hash))
		if err != nil {
			RenderTemplate(w, "register.html", map[string]string{"Error": "Failed to register user."})
			return
		}

		// Registration successful, redirect to login
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Method not allowed
	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

// LoginHandler handles GET and POST for /login
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		// Show the login form
		RenderTemplate(w, "login.html", nil)
		return
	}

	if r.Method == http.MethodPost {
		email := r.FormValue("email")
		password := r.FormValue("password")

		// Simple validation
		if email == "" || password == "" {
			RenderTemplate(w, "login.html", map[string]string{"Error": "All fields are required."})
			return
		}

		// Look up user by email
		var id int
		var username, passwordHash string
		err := database.DB.QueryRow("SELECT id, username, password_hash FROM users WHERE email = ?", email).Scan(&id, &username, &passwordHash)
		if err == sql.ErrNoRows {
			RenderTemplate(w, "login.html", map[string]string{"Error": "Invalid email or password."})
			return
		} else if err != nil {
			RenderTemplate(w, "login.html", map[string]string{"Error": "Database error."})
			return
		}

		// Compare password
		err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(password))
		if err != nil {
			RenderTemplate(w, "login.html", map[string]string{"Error": "Invalid email or password."})
			return
		}

		// Enforce one session per user: delete old sessions
		_, _ = database.DB.Exec("DELETE FROM sessions WHERE user_id = ?", id)

		// Create a new session (UUID)
		sessionToken := uuid.New().String()
		expiresAt := time.Now().Add(24 * time.Hour) // Session valid for 24 hours

		// Store session in DB
		_, err = database.DB.Exec("INSERT INTO sessions (user_id, session_token, expires_at) VALUES (?, ?, ?)", id, sessionToken, expiresAt)
		if err != nil {
			RenderTemplate(w, "login.html", map[string]string{"Error": "Failed to create session."})
			return
		}

		// Set session cookie
		cookie := &http.Cookie{
			Name:     "session_token",
			Value:    sessionToken,
			Expires:  expiresAt,
			HttpOnly: true,
			Path:     "/",
		}
		http.SetCookie(w, cookie)

		// Login successful, redirect to home
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	// Method not allowed
	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
}

// LogoutHandler logs the user out by deleting the session and clearing the cookie
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_token")
	if err == nil {
		// Delete session from DB
		database.DB.Exec("DELETE FROM sessions WHERE session_token = ?", cookie.Value)
		// Clear the cookie
		cleared := &http.Cookie{
			Name:     "session_token",
			Value:    "",
			Expires:  time.Unix(0, 0),
			HttpOnly: true,
			Path:     "/",
		}
		http.SetCookie(w, cleared)
	}
	// Redirect to homepage
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
