package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"incidenthub/backend/internal/handlers"
	"incidenthub/backend/internal/middleware"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/jmoiron/sqlx"
)

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://postgres:postgres@db:5432/incidenthub?sslmode=disable"
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "default-secret-change-in-production"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	frontendURL := os.Getenv("FRONTEND_URL")
	if frontendURL == "" {
		frontendURL = "http://localhost:3000"
	}

	sqlDB, err := sqlx.Open("postgres", databaseURL)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer sqlDB.Close()

	if err := sqlDB.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}

	db := sqlDB

	runMigrations(db)

	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{frontendURL},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	authHandler := handlers.NewAuthHandler(db, jwtSecret)
	incidentHandler := handlers.NewIncidentHandler(db)
	commentHandler := handlers.NewCommentHandler(db)
	dashboardHandler := handlers.NewDashboardHandler(db)

	api := r.Group("/api")
	{
		api.POST("/auth/register", authHandler.Register)
		api.POST("/auth/login", authHandler.Login)

		auth := api.Group("", middleware.AuthRequired(jwtSecret))
		{
			auth.GET("/dashboard/stats", dashboardHandler.GetStats)

			auth.GET("/incidents", incidentHandler.List)
			auth.POST("/incidents", incidentHandler.Create)
			auth.GET("/incidents/:id", incidentHandler.Get)
			auth.PUT("/incidents/:id", incidentHandler.Update)
			auth.DELETE("/incidents/:id", incidentHandler.Delete)

			auth.GET("/incidents/:id/comments", commentHandler.List)
			auth.POST("/incidents/:id/comments", commentHandler.Create)
		}
	}

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	fmt.Printf("Server starting on port %s\n", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func runMigrations(db *sqlx.DB) {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id VARCHAR(36) PRIMARY KEY,
			email VARCHAR(255) UNIQUE NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			name VARCHAR(255) NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);
		CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);`,
		`CREATE TABLE IF NOT EXISTS incidents (
			id VARCHAR(36) PRIMARY KEY,
			title VARCHAR(500) NOT NULL,
			description TEXT NOT NULL,
			severity VARCHAR(20) NOT NULL DEFAULT 'LOW',
			status VARCHAR(20) NOT NULL DEFAULT 'OPEN',
			assignee_id VARCHAR(36),
			created_by VARCHAR(36) NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			resolved_at TIMESTAMP WITH TIME ZONE,
			CONSTRAINT fk_incidents_assignee FOREIGN KEY (assignee_id) REFERENCES users(id) ON DELETE SET NULL,
			CONSTRAINT fk_incidents_creator FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE CASCADE
		);
		CREATE INDEX IF NOT EXISTS idx_incidents_status ON incidents(status);
		CREATE INDEX IF NOT EXISTS idx_incidents_severity ON incidents(severity);
		CREATE INDEX IF NOT EXISTS idx_incidents_assignee ON incidents(assignee_id);
		CREATE INDEX IF NOT EXISTS idx_incidents_created_at ON incidents(created_at DESC);
		CREATE INDEX IF NOT EXISTS idx_incidents_created_by ON incidents(created_by);`,
		`CREATE TABLE IF NOT EXISTS comments (
			id VARCHAR(36) PRIMARY KEY,
			incident_id VARCHAR(36) NOT NULL,
			user_id VARCHAR(36) NOT NULL,
			body TEXT NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
			CONSTRAINT fk_comments_incident FOREIGN KEY (incident_id) REFERENCES incidents(id) ON DELETE CASCADE,
			CONSTRAINT fk_comments_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
		);
		CREATE INDEX IF NOT EXISTS idx_comments_incident ON comments(incident_id);
		CREATE INDEX IF NOT EXISTS idx_comments_user ON comments(user_id);`,
		`CREATE OR REPLACE FUNCTION update_updated_at_column()
		RETURNS TRIGGER AS $$
		BEGIN
			NEW.updated_at = NOW();
			RETURN NEW;
		END;
		$$ language 'plpgsql';
		DROP TRIGGER IF EXISTS trigger_users_updated_at ON users;
		CREATE TRIGGER trigger_users_updated_at BEFORE UPDATE ON users
			FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
		DROP TRIGGER IF EXISTS trigger_incidents_updated_at ON incidents;
		CREATE TRIGGER trigger_incidents_updated_at BEFORE UPDATE ON incidents
			FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
		DROP TRIGGER IF EXISTS trigger_comments_updated_at ON comments;
		CREATE TRIGGER trigger_comments_updated_at BEFORE UPDATE ON comments
			FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();`,
	}

	for i, migration := range migrations {
		if _, err := db.Exec(migration); err != nil {
			log.Printf("Warning: migration %d failed: %v", i+1, err)
		}
	}
}