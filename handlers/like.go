package handlers

import (
	"forum/database"
	"forum/utils"
	"net/http"
	"strconv"
)

// LikeHandler handles POST /like for posts
func LikeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		utils.HandleError(w, 405, "Method Not Allowed", "This endpoint only accepts POST requests")
		return
	}

	userID, _ := utils.GetCurrentUser(r)

	postIDStr := r.FormValue("post_id")
	commentIDStr := r.FormValue("comment_id")
	isLikeStr := r.FormValue("is_like")
	postID, _ := strconv.Atoi(postIDStr)
	commentID, _ := strconv.Atoi(commentIDStr)
	isLike, _ := strconv.Atoi(isLikeStr)
	if (postID <= 0 && commentID <= 0) || (isLike != 0 && isLike != 1) {
		utils.HandleError(w, 400, "Invalid Like Data", "Invalid post/comment ID or like value")
		return
	}

	if commentID > 0 {
		// Verify comment exists and belongs to a valid post
		var postID int
		err := database.DB.QueryRow("SELECT post_id FROM comments WHERE id = ?", commentID).Scan(&postID)
		if err != nil {
			utils.HandleError(w, 404, "Comment Not Found", "The comment you're trying to like doesn't exist")
			return
		}

		// Like/dislike for a comment
		var existingID int
		err = database.DB.QueryRow("SELECT id FROM likes WHERE user_id = ? AND comment_id = ?", userID, commentID).Scan(&existingID)
		if err == nil {
			_, _ = database.DB.Exec("UPDATE likes SET is_like = ? WHERE id = ?", isLike, existingID)
		} else {
			_, _ = database.DB.Exec("INSERT INTO likes (user_id, comment_id, is_like) VALUES (?, ?, ?)", userID, commentID, isLike)
		}
	} else {
		// Verify post exists
		var postExists int
		err := database.DB.QueryRow("SELECT COUNT(*) FROM posts WHERE id = ?", postID).Scan(&postExists)
		if err != nil || postExists == 0 {
			utils.HandleError(w, 404, "Post Not Found", "The post you're trying to like doesn't exist")
			return
		}

		// Like/dislike for a post
		var existingID int
		err = database.DB.QueryRow("SELECT id FROM likes WHERE user_id = ? AND post_id = ? AND comment_id IS NULL", userID, postID).Scan(&existingID)
		if err == nil {
			_, _ = database.DB.Exec("UPDATE likes SET is_like = ? WHERE id = ?", isLike, existingID)
		} else {
			_, _ = database.DB.Exec("INSERT INTO likes (user_id, post_id, is_like) VALUES (?, ?, ?)", userID, postID, isLike)
		}
	}

	// Redirect back to the referring page
	ref := r.Referer()
	if ref == "" {
		ref = "/"
	}
	http.Redirect(w, r, ref, http.StatusSeeOther)
}
