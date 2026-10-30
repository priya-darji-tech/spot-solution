package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"spot-solution/internal/config"
	"spot-solution/internal/database"
	"spot-solution/internal/email"
	"spot-solution/internal/handlers"
	"spot-solution/internal/middleware"
	"spot-solution/internal/models"
	"spot-solution/internal/repository"
	"spot-solution/internal/services"
)

func main() {
	// 1. Load Configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load application configuration: %v", err)
	}

	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 2. Initialize Database & Migrations
	db, err := database.NewConnection(cfg)
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}
	defer db.Close()

	if err := db.RunMigrations(); err != nil {
		log.Fatalf("Failed to execute migrations: %v", err)
	}

	if err := db.SeedDefaultAdmin(cfg); err != nil {
		log.Printf("Warning: failed to seed default admin: %v", err)
	}

	// 3. Initialize Repositories
	contactRepo := repository.NewContactRepository(db)
	leadRepo := repository.NewLeadRepository(db)
	userRepo := repository.NewUserRepository(db)
	outboxRepo := repository.NewOutboxRepository(db)

	// 4. Initialize Services & Mailer
	mailer := email.NewMailer(cfg)
	crmService := services.NewLeadService(leadRepo)
	contactService := services.NewContactService(cfg, db, contactRepo, crmService, outboxRepo, mailer)
	authService := services.NewAuthService(cfg, userRepo)

	// 5. Start Background Outbox Worker
	workerCtx, workerCancel := context.WithCancel(context.Background())
	defer workerCancel()
	go mailer.StartOutboxWorker(workerCtx, outboxRepo)

	// 6. Initialize HTTP Handlers
	pageHandler := handlers.NewPageHandler(cfg)
	contactHandler := handlers.NewContactHandler(contactService)
	leadHandler := handlers.NewLeadHandler(crmService)
	authHandler := handlers.NewAuthHandler(cfg, authService)
	seoHandler := handlers.NewSEOHandler(cfg)

	// 7. Initialize Gin Engine & Routes
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(gin.Logger())

	// Security & Rate Limiting Middleware
	r.Use(middleware.SecurityHeaders(cfg))
	r.Use(middleware.MaxBodySize(cfg.MaxBodySizeBytes))
	r.Use(middleware.IPRateLimiter(cfg.RateLimitRPS, cfg.RateLimitBurst))
	r.Use(middleware.CSRFMiddleware(cfg))
	r.Use(middleware.AuthMiddleware(authService))

	// Static Assets & Templates
	r.Static("/static", "./web/static")
	r.LoadHTMLGlob("web/templates/**/*")

	// SEO & Health Routes
	r.GET("/robots.txt", seoHandler.RobotsTxt)
	r.GET("/sitemap.xml", seoHandler.SitemapXML)
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"app":       cfg.AppName,
			"timestamp": time.Now().UTC(),
		})
	})

	// Public Marketing Routes
	r.GET("/", pageHandler.Home)
	r.GET("/about", pageHandler.About)
	r.GET("/about-us", pageHandler.About)
	r.GET("/careers", pageHandler.Careers)
	r.GET("/services", pageHandler.Services)
	r.GET("/services/:slug", pageHandler.ServiceDetail)
	r.GET("/industries", pageHandler.Industries)
	r.GET("/why-us", pageHandler.WhyUs)
	r.GET("/portfolio", pageHandler.Portfolio)
	r.GET("/blog", pageHandler.Blog)
	r.GET("/contact", pageHandler.Contact)
	r.GET("/contact-us", pageHandler.Contact)
	r.GET("/privacy", pageHandler.Privacy)
	r.GET("/privacy-policy", pageHandler.Privacy)
	r.GET("/terms", pageHandler.Terms)
	r.GET("/spot-academy", pageHandler.Services)
	r.GET("/spot-academy/*sub", pageHandler.Services)
	r.GET("/mobile", func(c *gin.Context) { c.Redirect(http.StatusMovedPermanently, "/services/mobile") })
	r.GET("/web-applications", func(c *gin.Context) { c.Redirect(http.StatusMovedPermanently, "/services/web-applications") })
	r.GET("/ai", func(c *gin.Context) { c.Redirect(http.StatusMovedPermanently, "/services/ai") })
	r.GET("/sharepoint", func(c *gin.Context) { c.Redirect(http.StatusMovedPermanently, "/services/sharepoint") })
	r.GET("/services/change-management", func(c *gin.Context) { c.Redirect(http.StatusMovedPermanently, "/services/change-managment") })
	r.GET("/services/cloud-solutions", func(c *gin.Context) { c.Redirect(http.StatusMovedPermanently, "/services/hosting") })
	r.GET("/services/business-process-automation", func(c *gin.Context) { c.Redirect(http.StatusMovedPermanently, "/services/sharepoint") })
	r.GET("/services/ai-automation", func(c *gin.Context) { c.Redirect(http.StatusMovedPermanently, "/services/ai") })
	r.GET("/services/api-integration", func(c *gin.Context) { c.Redirect(http.StatusMovedPermanently, "/services/custom-software") })
	r.GET("/services/it-consulting", func(c *gin.Context) { c.Redirect(http.StatusMovedPermanently, "/services/requirements") })
	r.GET("/services/digital-transformation", func(c *gin.Context) { c.Redirect(http.StatusMovedPermanently, "/services/change-managment") })
	r.GET("/services/m365", func(c *gin.Context) { c.Redirect(http.StatusMovedPermanently, "/services/miscrosoft-365-experts") })
	r.GET("/services/power-apps", func(c *gin.Context) { c.Redirect(http.StatusMovedPermanently, "/services/powerplatform") })
	r.GET("/services/power-platform", func(c *gin.Context) { c.Redirect(http.StatusMovedPermanently, "/services/powerplatform") })

	// Public API
	r.POST("/api/contact", contactHandler.SubmitContact)

	// Auth Endpoints
	r.GET("/admin/login", authHandler.ShowLoginPage)
	r.POST("/api/admin/login", authHandler.Login)
	r.POST("/api/admin/logout", authHandler.Logout)

	// Direct /admin redirect
	r.GET("/admin", func(c *gin.Context) {
		c.Redirect(http.StatusSeeOther, "/admin/leads")
	})

	// Protected Admin Dashboard & CRM Routes
	adminGroup := r.Group("/admin")
	adminGroup.Use(middleware.RequireAuth())
	{
		adminGroup.GET("/leads", leadHandler.AdminDashboard)
	}

	apiLeadsGroup := r.Group("/api/leads")
	apiLeadsGroup.Use(middleware.RequireAuth())
	{
		apiLeadsGroup.GET("", leadHandler.GetLeads)
		apiLeadsGroup.GET("/:id", leadHandler.GetLeadByID)
		apiLeadsGroup.POST("", middleware.RequireRole(models.RoleAdmin, models.RoleManager), leadHandler.CreateLead)
		apiLeadsGroup.PATCH("/:id/status", middleware.RequireRole(models.RoleAdmin, models.RoleManager), leadHandler.UpdateStatus)
		apiLeadsGroup.DELETE("/:id", middleware.RequireRole(models.RoleAdmin), leadHandler.DeleteLead)
	}

	// 404 Handler
	r.NoRoute(pageHandler.NotFound)

	// 8. Server Setup with Graceful Shutdown
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("%s server running at %s (Port :%s)", cfg.AppName, cfg.BaseURL, cfg.Port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server listen error: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server gracefully...")

	workerCancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting")
}
