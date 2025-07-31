package utils

import (
	"forum/database"
	"net/http"
)

// GetCurrentUser checks the session_token cookie and returns the user's id and username if logged in.
// Returns (0, "") if not logged in or session is invalid/expired.
func GetCurrentUser(r *http.Request) (int, string) {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		return 0, ""
	}

	var userID int
	var username string
	// Join sessions and users to get username
	err = database.DB.QueryRow(`
		SELECT users.id, users.username
		FROM sessions
		JOIN users ON sessions.user_id = users.id
		WHERE sessions.session_token = ? AND sessions.expires_at > datetime('now')
	`, cookie.Value).Scan(&userID, &username)
	if err != nil {
		return 0, ""
	}
	return userID, username
}

// RequireAuth middleware ensures user is logged in
func RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, _ := GetCurrentUser(r)
		if userID == 0 {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next(w, r)
	}
}

// RequireGuest middleware ensures user is NOT logged in (for login/register pages)
func RequireGuest(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, _ := GetCurrentUser(r)
		if userID != 0 {
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		next(w, r)
	}
}
