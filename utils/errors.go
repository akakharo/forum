package utils

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
)

// ErrorData represents data passed to error templates
type ErrorData struct {
	StatusCode int
	Message    string
	Details    string
}

// HandleError renders an appropriate error page based on the status code
func HandleError(w http.ResponseWriter, statusCode int, message string, details string) {
	log.Printf("Error %d: %s - %s", statusCode, message, details)

	// Set the status code first
	w.WriteHeader(statusCode)

	var tmpl *template.Template
	var err error

	// Try to load the appropriate template
	switch statusCode {
	case 404:
		tmpl, err = template.ParseFiles("templates/404.html")
	case 500:
		tmpl, err = template.ParseFiles("templates/500.html")
	default:
		// For other status codes, try 500 template first, then fallback
		tmpl, err = template.ParseFiles("templates/500.html")
	}

	// If template loading fails, try 404 template as fallback
	if err != nil {
		tmpl, err = template.ParseFiles("templates/404.html")
	}

	// If both templates fail, send a simple HTML error response
	if err != nil {
		htmlError := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <title>Error %d - DinoForum</title>
    <style>
        body { font-family: Arial, sans-serif; text-align: center; padding: 50px; }
        .error { color: #d32f2f; font-size: 2rem; margin-bottom: 20px; }
        .message { color: #666; font-size: 1.1rem; margin-bottom: 30px; }
        .back { background: #388e3c; color: white; padding: 10px 20px; text-decoration: none; border-radius: 5px; }
    </style>
</head>
<body>
    <div class="error">🦖 Error %d</div>
    <div class="message">%s</div>
    <a href="/" class="back">🦕 Back to Home</a>
</body>
</html>`, statusCode, statusCode, message)

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, htmlError)
		return
	}

	// Execute the template
	data := ErrorData{
		StatusCode: statusCode,
		Message:    message,
		Details:    details,
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		// If template execution fails, send a simple error response
		htmlError := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <title>Error %d - DinoForum</title>
    <style>
        body { font-family: Arial, sans-serif; text-align: center; padding: 50px; }
        .error { color: #d32f2f; font-size: 2rem; margin-bottom: 20px; }
        .message { color: #666; font-size: 1.1rem; margin-bottom: 30px; }
        .back { background: #388e3c; color: white; padding: 10px 20px; text-decoration: none; border-radius: 5px; }
    </style>
</head>
<body>
    <div class="error">🦖 Error %d</div>
    <div class="message">%s</div>
    <a href="/" class="back">🦕 Back to Home</a>
</body>
</html>`, statusCode, statusCode, message)

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, htmlError)
		return
	}
}

// HandlePanic recovers from panics and shows a 500 error
func HandlePanic(w http.ResponseWriter, r *http.Request) {
	if err := recover(); err != nil {
		log.Printf("Panic recovered: %v", err)
		HandleError(w, 500, "Internal Server Error", "The server encountered an unexpected error")
	}
}
