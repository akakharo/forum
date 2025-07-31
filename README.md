# Dino Forum

A secure, feature-rich forum application built with Go, featuring user authentication, post management, comments, likes/dislikes, and category filtering.

## Features

- 🔐 **Secure Authentication**: User registration and login with session management
- 📝 **Post Management**: Create, view, and delete posts with rich content
- 💬 **Comments**: Add comments to posts with threading support
- 👍 **Likes/Dislikes**: Interactive voting system for posts
- 🏷️ **Categories**: Organize posts with category filtering
- 🛡️ **Security**: CSRF protection, input sanitization, and security headers
- 🎨 **Modern UI**: Clean, responsive design with CSS styling
- 🗄️ **SQLite Database**: Lightweight, file-based database

## Quick Start with Docker

### Prerequisites
- Docker
- Docker Compose

### Running with Docker Compose (Recommended)

1. **Clone and navigate to the project:**
   ```bash
   cd forum
   ```

2. **Build and run with Docker Compose:**
   ```bash
   docker-compose up --build
   ```

3. **Access the application:**
   Open your browser and go to `http://localhost:8080`

4. **Stop the application:**
   ```bash
   docker-compose down
   ```

### Running with Docker directly

1. **Build the Docker image:**
   ```bash
   docker build -t dino-forum .
   ```

2. **Run the container:**
   ```bash
   docker run -p 8080:8080 -v forum_data:/app/data dino-forum
   ```

3. **Access the application:**
   Open your browser and go to `http://localhost:8080`

## Development Setup

### Prerequisites
- Go 1.21 or later
- SQLite3

### Local Development

1. **Clone the repository:**
   ```bash
   git clone <repository-url>
   cd forum
   ```

2. **Install dependencies:**
   ```bash
   go mod download
   ```

3. **Run the application:**
   ```bash
   go run main.go
   ```

4. **Access the application:**
   Open your browser and go to `http://localhost:8080`

## Project Structure

```
forum/
├── database/          # Database initialization and schema
├── handlers/          # HTTP request handlers
├── static/           # Static assets (CSS, JS)
├── templates/        # HTML templates
├── utils/            # Utility functions (auth, security, etc.)
├── main.go           # Application entry point
├── Dockerfile        # Docker configuration
├── docker-compose.yml # Docker Compose configuration
└── README.md         # This file
```

## Security Features

- **CSRF Protection**: All forms include CSRF tokens
- **Input Sanitization**: User inputs are sanitized to prevent XSS
- **Session Security**: Secure session management with expiration
- **Security Headers**: Comprehensive security headers
- **Request Validation**: Origin and method validation
- **SQL Injection Prevention**: Parameterized queries

## Database

The application uses SQLite for data storage. The database file is created automatically when the application starts.

### Schema
- **users**: User accounts and authentication
- **posts**: Forum posts with titles and content
- **comments**: Comments on posts
- **likes**: User likes/dislikes on posts
- **categories**: Post categories
- **post_categories**: Many-to-many relationship between posts and categories

## API Endpoints

- `GET /` - Homepage with posts listing
- `GET /register` - Registration page
- `POST /register` - User registration
- `GET /login` - Login page
- `POST /login` - User login
- `POST /logout` - User logout
- `GET /create_post` - Create post page
- `POST /create_post` - Create new post
- `GET /post?id=<id>` - View specific post
- `POST /comment` - Add comment to post
- `POST /like` - Like/dislike post
- `POST /delete_post` - Delete post (owner only)
- `POST /delete_comment` - Delete comment (owner only)

## Environment Variables

- `DB_PATH`: Database file path (default: `dinoforum.db`)
- `TZ`: Timezone (default: `UTC`)

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Test thoroughly
5. Submit a pull request

## License

This project is open source and available under the [MIT License](LICENSE). 