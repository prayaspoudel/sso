package main

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"sso/config"
	"sso/handlers"
	"sso/middleware"
	"sso/repository"
	"sso/services"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Connect to database
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.DBName,
		cfg.Database.SSLMode,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Test database connection
	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}
	log.Println("✓ Database connected successfully")

	// Create sqlx.DB wrapper for repositories that need it
	dbx := sqlx.NewDb(db, "postgres")

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)
	sessionRepo := repository.NewSessionRepository(db)
	tokenRepo := repository.NewTokenRepository(db)
	userMgmtRepo := repository.NewUserManagementRepository(db)
	notificationRepo := repository.NewNotificationRepository(dbx)

	// Initialize WebSocket hub and start it
	wsHub := services.NewWebSocketHub()
	go wsHub.Run()
	log.Println("✓ WebSocket hub started")

	// Initialize services
	authService := services.NewAuthService(cfg, userRepo, sessionRepo, tokenRepo)

	// For now, create a nil email service (TODO: initialize properly with config)
	var emailService *services.EmailService = nil

	userMgmtService := services.NewUserManagementService(userMgmtRepo, userRepo, emailService)
	companyMgmtService := services.NewCompanyManagementService(db)
	auditLogService := services.NewAuditLogService(db)
	notificationService := services.NewNotificationService(notificationRepo, wsHub)

	// Start scheduled notification cleanup
	notificationService.ScheduleCleanup()
	log.Println("✓ Notification cleanup scheduler started")

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService)
	userMgmtHandler := handlers.NewUserManagementHandler(userMgmtService)
	companyMgmtHandler := handlers.NewCompanyManagementHandler(companyMgmtService)
	auditLogHandler := handlers.NewAuditLogHandler(auditLogService)
	wsHandler := handlers.NewWebSocketHandler(notificationService)

	// Setup Gin router
	if cfg.Server.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	// Apply middleware
	router.Use(middleware.CORSMiddleware(cfg.Server.AllowedOrigins))
	router.Use(middleware.LoggerMiddleware())

	// API version 1
	v1 := router.Group("/api/v1")

	// Public routes
	public := v1.Group("/auth")
	{
		public.POST("/register", authHandler.Register)
		public.POST("/login", authHandler.Login)
		public.POST("/refresh", authHandler.RefreshToken)
		public.POST("/logout", authHandler.Logout)
		public.GET("/validate", authHandler.ValidateToken)
	}

	// Protected routes (require authentication)
	protected := v1.Group("/auth")
	protected.Use(middleware.AuthMiddleware(authService))
	{
		protected.GET("/me", authHandler.Me)
		protected.POST("/change-password", authHandler.ChangePassword)
		protected.POST("/logout-all", authHandler.LogoutAll)
	}

	// User Management routes (protected)
	users := v1.Group("/users")
	users.Use(middleware.AuthMiddleware(authService))
	{
		users.GET("", userMgmtHandler.ListUsers)
		users.GET("/:id", userMgmtHandler.GetUser)
		users.POST("", userMgmtHandler.CreateUser)
		users.PUT("/:id", userMgmtHandler.UpdateUser)
		users.DELETE("/:id", userMgmtHandler.DeleteUser)
		users.GET("/stats", userMgmtHandler.GetUserStats)
	}

	// Company Management routes (protected)
	companies := v1.Group("/companies")
	companies.Use(middleware.AuthMiddleware(authService))
	{
		companies.GET("", companyMgmtHandler.ListCompanies)
		companies.GET("/:id", companyMgmtHandler.GetCompany)
		companies.POST("", companyMgmtHandler.CreateCompany)
		companies.PUT("/:id", companyMgmtHandler.UpdateCompany)
		companies.DELETE("/:id", companyMgmtHandler.DeleteCompany)
		companies.GET("/:id/users", companyMgmtHandler.GetCompanyUsers)
		companies.POST("/:id/users", companyMgmtHandler.AddUserToCompany)
		companies.DELETE("/:id/users/:user_id", companyMgmtHandler.RemoveUserFromCompany)
		companies.GET("/stats", companyMgmtHandler.GetCompanyStats)
	}

	// Audit Log routes (protected)
	auditLogs := v1.Group("/audit-logs")
	auditLogs.Use(middleware.AuthMiddleware(authService))
	{
		auditLogs.GET("", auditLogHandler.ListAuditLogs)
		auditLogs.GET("/:id", auditLogHandler.GetAuditLog)
		auditLogs.GET("/stats", auditLogHandler.GetAuditLogStats)
		auditLogs.POST("/timeline", auditLogHandler.GetAuditTimeline)
		auditLogs.POST("/export", auditLogHandler.ExportAuditLogs)
		auditLogs.POST("/cleanup", auditLogHandler.CleanupOldLogs)
		auditLogs.GET("/actions", auditLogHandler.GetDistinctActions)
		auditLogs.GET("/resources", auditLogHandler.GetDistinctResources)
		auditLogs.GET("/compare", auditLogHandler.CompareAuditLogs)
	}

	// WebSocket endpoint (requires authentication via query param or upgrade)
	v1.GET("/ws", wsHandler.HandleWebSocket)

	// Notification routes (protected)
	notifications := v1.Group("/notifications")
	notifications.Use(middleware.AuthMiddleware(authService))
	{
		notifications.GET("", wsHandler.ListNotifications)
		notifications.GET("/unread-count", wsHandler.GetUnreadCount)
		notifications.GET("/stats", wsHandler.GetNotificationStats)
		notifications.GET("/:id", wsHandler.GetNotification)
		notifications.POST("", wsHandler.CreateNotification)
		notifications.POST("/broadcast", wsHandler.BroadcastNotification)
		notifications.PUT("/:id/read", wsHandler.MarkAsRead)
		notifications.POST("/read", wsHandler.MarkMultipleAsRead)
		notifications.POST("/read-all", wsHandler.MarkAllAsRead)
		notifications.DELETE("/:id", wsHandler.DeleteNotification)
		notifications.POST("/delete", wsHandler.DeleteMultipleNotifications)
		notifications.GET("/preferences", wsHandler.GetPreference)
		notifications.PUT("/preferences", wsHandler.UpdatePreference)
		notifications.GET("/connections", wsHandler.GetConnectedUsers)
		notifications.POST("/disconnect/:id", wsHandler.DisconnectUser)
		notifications.POST("/test", wsHandler.SendTestNotification)
	}

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "healthy",
			"service": "sso",
			"version": "1.0.0",
		})
	})

	// Root endpoint
	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"service": "Union Products SSO",
			"version": "2.0.0",
			"endpoints": map[string]interface{}{
				"health": "/health",
				"auth": map[string]string{
					"register": "/api/v1/auth/register",
					"login":    "/api/v1/auth/login",
					"refresh":  "/api/v1/auth/refresh",
					"logout":   "/api/v1/auth/logout",
					"validate": "/api/v1/auth/validate",
					"me":       "/api/v1/auth/me (protected)",
				},
				"users": map[string]string{
					"list":   "/api/v1/users (protected)",
					"create": "/api/v1/users (protected)",
					"get":    "/api/v1/users/:id (protected)",
					"update": "/api/v1/users/:id (protected)",
					"delete": "/api/v1/users/:id (protected)",
					"stats":  "/api/v1/users/stats (protected)",
				},
				"companies": map[string]string{
					"list":   "/api/v1/companies (protected)",
					"create": "/api/v1/companies (protected)",
					"get":    "/api/v1/companies/:id (protected)",
					"stats":  "/api/v1/companies/stats (protected)",
				},
				"audit-logs": map[string]string{
					"list":  "/api/v1/audit-logs (protected)",
					"stats": "/api/v1/audit-logs/stats (protected)",
				},
				"notifications": map[string]string{
					"websocket": "/api/v1/ws (protected)",
					"list":      "/api/v1/notifications (protected)",
					"stats":     "/api/v1/notifications/stats (protected)",
				},
			},
		})
	})

	// Start server
	port := cfg.Server.Port
	log.Printf("✓ Starting SSO server on port %s", port)
	log.Printf("✓ Environment: %s", cfg.Server.Environment)
	log.Printf("✓ Allowed origins: %v", cfg.Server.AllowedOrigins)

	if err := router.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
