package handlers

import (
	"fmt"
	"forum/database"
	"forum/utils"
	"net/http"
	"strconv"
)

// CommentHandler handles POST /comment
func CommentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.HandleError(w, 405, "Method Not Allowed", "This endpoint only accepts POST requests")
		return
	}

	userID, _ := utils.GetCurrentUser(r)

	postIDStr := r.FormValue("post_id")
	content := utils.SanitizeHTML(r.FormValue("content"))
	postID, err := strconv.Atoi(postIDStr)
	if err != nil || postID <= 0 || content == "" {
		utils.HandleError(w, 400, "Invalid Comment Data", "Please provide valid post ID and comment content")
		return
	}
	if len(content) > 500 {
		utils.HandleError(w, 400, "Comment Too Long", "Comments must be 500 characters or less")
		return
	}

	_, err = database.DB.Exec("INSERT INTO comments (post_id, user_id, content) VALUES (?, ?, ?)", postID, userID, content)
	if err != nil {
		utils.HandleError(w, 500, "Database Error", "Failed to add comment")
		return
	}

	// Redirect back to the post page
	http.Redirect(w, r, "/post?id="+postIDStr, http.StatusSeeOther)
}

// DeleteCommentHandler handles POST /delete_comment
func DeleteCommentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.HandleError(w, 405, "Method Not Allowed", "This endpoint only accepts POST requests")
		return
	}

	userID, _ := utils.GetCurrentUser(r)

	commentIDStr := r.FormValue("comment_id")
	commentID, err := strconv.Atoi(commentIDStr)
	if err != nil || commentID <= 0 {
		utils.HandleError(w, 400, "Invalid Comment ID", "The comment ID provided is not valid")
		return
	}

	// Check if the comment exists and belongs to the current user
	var commentUserID, postID int
	err = database.DB.QueryRow("SELECT user_id, post_id FROM comments WHERE id = ?", commentID).Scan(&commentUserID, &postID)
	if err != nil {
		utils.HandleError(w, 404, "Comment Not Found", "The comment you're trying to delete doesn't exist")
		return
	}

	if commentUserID != userID {
		utils.HandleError(w, 403, "Forbidden", "You can only delete your own comments")
		return
	}

	// Delete the comment
	result, err := database.DB.Exec("DELETE FROM comments WHERE id = ?", commentID)
	if err != nil {
		utils.HandleError(w, 500, "Database Error", "Failed to delete comment")
		return
	}

	// Check how many rows were affected
	rowsAffected, _ := result.RowsAffected()
	fmt.Printf("Deleted comment %d, rows affected: %d\n", commentID, rowsAffected)

	// Redirect back to the post page
	http.Redirect(w, r, "/post?id="+strconv.Itoa(postID), http.StatusSeeOther)
}
