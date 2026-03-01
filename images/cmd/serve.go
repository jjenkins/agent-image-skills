package cmd

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/jjenkins/labnocturne/images/internal/handlers"
	mcpinternal "github.com/jjenkins/labnocturne/images/internal/mcp"
	"github.com/jjenkins/labnocturne/images/internal/middleware"
	"github.com/jjenkins/labnocturne/images/internal/ratelimit"
	"github.com/jjenkins/labnocturne/images/internal/store"
	mcpserver "github.com/mark3labs/mcp-go/server"
	"github.com/spf13/cobra"
	"github.com/stripe/stripe-go/v76"
)

var port string

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the Image API server",
	Long:  `Start the web server for the Lab Nocturne Image Storage API.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Use PORT env var if set, otherwise use flag value
		if envPort := os.Getenv("PORT"); envPort != "" && port == "8080" {
			port = envPort
		}

		// Database connection is required
		dsn := os.Getenv("DATABASE_URL")
		if dsn == "" {
			log.Fatal("DATABASE_URL environment variable is required")
		}

		db, err := store.NewDB(dsn)
		if err != nil {
			log.Fatalf("Failed to connect to database: %v", err)
		}
		defer db.Close()
		log.Printf("Successfully connected to database")

		// Initialize AWS S3 client
		baseURL := os.Getenv("BASE_URL")
		if baseURL == "" {
			log.Fatal("BASE_URL environment variable is required")
		}

		s3Bucket := os.Getenv("AWS_S3_BUCKET")
		if s3Bucket == "" {
			log.Fatal("AWS_S3_BUCKET environment variable is required")
		}

		awsRegion := os.Getenv("AWS_REGION")
		if awsRegion == "" {
			log.Fatal("AWS_REGION environment variable is required")
		}

		cfg, err := config.LoadDefaultConfig(context.Background(),
			config.WithRegion(awsRegion),
		)
		if err != nil {
			log.Fatalf("Failed to load AWS config: %v", err)
		}

		s3Client := s3.NewFromConfig(cfg)
		log.Printf("Successfully initialized S3 client for bucket: %s", s3Bucket)

		// Initialize Stripe
		stripeSecretKey := os.Getenv("STRIPE_SECRET_KEY")
		if stripeSecretKey != "" {
			stripe.Key = stripeSecretKey
			log.Printf("Stripe integration enabled")
		} else {
			log.Printf("Warning: STRIPE_SECRET_KEY not set, checkout endpoints will not work")
		}

		app := fiber.New(fiber.Config{
			AppName:   "Lab Nocturne Images API",
			BodyLimit: 110 * 1024 * 1024, // 110MB (above live key limit of 100MB)
		})

		// Initialize rate limiter
		rateLimiter := ratelimit.NewLimiter()

		// Start cleanup goroutine
		go func() {
			ticker := time.NewTicker(5 * time.Minute)
			defer ticker.Stop()
			for range ticker.C {
				rateLimiter.Cleanup(1 * time.Hour)
			}
		}()

		// Middleware
		app.Use(logger.New())
		app.Use(middleware.SecurityHeadersMiddleware())
		app.Use(middleware.RateLimitMiddleware(rateLimiter, db))

		// SEO Routes
		app.Get("/robots.txt", handlers.RobotsTxtHandler())
		app.Get("/sitemap.xml", handlers.SitemapXMLHandler())
		app.Get("/site.webmanifest", handlers.SiteWebmanifestHandler())

		// Static Files (OG images, favicons)
		app.Static("/", "./static", fiber.Static{
			Compress:      true,
			ByteRange:     true,
			Browse:        false,
			Index:         "",
			CacheDuration: 24 * time.Hour,
			MaxAge:        86400, // 1 day in seconds
		})

		// Routes
		app.Get("/", handlers.HomeHandler())
		app.Get("/docs", handlers.DocsHandler())
		app.Get("/api", handlers.APIInfoHandler())
		app.Get("/health", handlers.HealthHandler())
		app.Get("/key",
			middleware.RateLimitByIP(rateLimiter, "key generation", 5, time.Hour), // 5 test keys per hour per IP
			handlers.GenerateKeyHandler(db))
		app.Post("/upload", handlers.UploadHandler(db, s3Client, baseURL, s3Bucket))
		app.Get("/i/:ulid.:ext", handlers.GetFileHandler(db, s3Client, baseURL, s3Bucket))
		app.Delete("/i/:id", handlers.DeleteFileHandler(db))
		app.Get("/files", handlers.ListFilesHandler(db))
		app.Get("/stats", handlers.StatsHandler(db, rateLimiter))

		// Stripe Checkout routes
		app.Post("/checkout", handlers.CheckoutHandler(db, baseURL))
		app.Get("/key/retrieve", handlers.RetrieveKeyHandler(db, baseURL))
		app.Post("/webhook", handlers.WebhookHandler(db, baseURL))
		app.Get("/success", handlers.SuccessHandler())

		// MCP (Model Context Protocol) server - SSE transport
		mcpDeps := &mcpinternal.Deps{
			DB:          db,
			S3Client:    s3Client,
			S3Bucket:    s3Bucket,
			BaseURL:     baseURL,
			RateLimiter: rateLimiter,
		}
		mcpSrv := mcpinternal.NewMCPServer(mcpDeps)
		sseServer := mcpserver.NewSSEServer(mcpSrv,
			mcpserver.WithBaseURL(baseURL),
			mcpserver.WithStaticBasePath("/mcp"),
			mcpserver.WithSSEContextFunc(mcpinternal.SSEContextFunc(db)),
		)
		app.Get("/mcp/sse", adaptor.HTTPHandler(sseServer.SSEHandler()))
		app.Post("/mcp/messages", adaptor.HTTPHandler(sseServer.MessageHandler()))

		log.Printf("Starting Lab Nocturne Images API on :%s", port)
		if err := app.Listen(":" + port); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	},
}

func init() {
	rootCmd.AddCommand(serveCmd)
	serveCmd.Flags().StringVarP(&port, "port", "p", "8080", "Port to run the server on")
}
