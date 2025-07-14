package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"locals3/internal/auth"
	"locals3/internal/config"
	"locals3/internal/handlers"
	"locals3/internal/storage"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logrus.Fatalf("Failed to load configuration: %v", err)
	}

	// Setup logging
	setupLogging(cfg.LogLevel)

	// Initialize storage backend
	storageBackend := storage.NewFileSystemStorage(cfg.DataDir)

	// Initialize auth provider
	authProvider := auth.NewAWSV4Auth(cfg.AccessKey, cfg.SecretKey, cfg.Region)

	// Initialize handlers
	handlerConfig := &handlers.Config{
		Storage:     storageBackend,
		Auth:        authProvider,
		Region:      cfg.Region,
		BaseDomain:  cfg.BaseDomain,
		DisableAuth: cfg.DisableAuth,
	}
	h := handlers.New(handlerConfig)

	// Setup router
	router := setupRouter(h)

	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	// Start server
	go func() {
		logrus.Infof("Starting LocalS3 server on port %d", cfg.Port)
		logrus.Infof("Data directory: %s", cfg.DataDir)
		logrus.Infof("Region: %s", cfg.Region)
		logrus.Infof("Access Key: %s", cfg.AccessKey)

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	logrus.Info("Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logrus.Errorf("Server shutdown error: %v", err)
	}

	logrus.Info("Server stopped")
}

func setupLogging(level string) {
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	switch level {
	case "debug":
		logrus.SetLevel(logrus.DebugLevel)
	case "info":
		logrus.SetLevel(logrus.InfoLevel)
	case "warn":
		logrus.SetLevel(logrus.WarnLevel)
	case "error":
		logrus.SetLevel(logrus.ErrorLevel)
	default:
		logrus.SetLevel(logrus.InfoLevel)
	}
}

func setupRouter(h *handlers.Handler) *mux.Router {
	router := mux.NewRouter()

	// Health check endpoint
	router.HandleFunc("/health", h.HealthCheck).Methods("GET")

	// S3 API endpoints
	s3Router := router.PathPrefix("/").Subrouter()

	// Bucket operations
	s3Router.HandleFunc("/", h.ListBuckets).Methods("GET")
	s3Router.HandleFunc("/{bucket}", h.CreateBucket).Methods("PUT")
	s3Router.HandleFunc("/{bucket}", h.DeleteBucket).Methods("DELETE")
	s3Router.HandleFunc("/{bucket}", h.ListObjects).Methods("GET")
	s3Router.HandleFunc("/{bucket}/", h.ListObjects).Methods("GET")

	// Object operations
	s3Router.HandleFunc("/{bucket}/{key:.*}", h.PutObject).Methods("PUT")
	s3Router.HandleFunc("/{bucket}/{key:.*}", h.GetObject).Methods("GET")
	s3Router.HandleFunc("/{bucket}/{key:.*}", h.DeleteObject).Methods("DELETE")
	s3Router.HandleFunc("/{bucket}/{key:.*}", h.HeadObject).Methods("HEAD")

	// Multipart upload operations
	s3Router.HandleFunc("/{bucket}/{key:.*}", h.InitiateMultipartUpload).Methods("POST").Queries("uploads", "")
	s3Router.HandleFunc("/{bucket}/{key:.*}", h.UploadPart).Methods("PUT").Queries("partNumber", "{partNumber}", "uploadId", "{uploadId}")
	s3Router.HandleFunc("/{bucket}/{key:.*}", h.CompleteMultipartUpload).Methods("POST").Queries("uploadId", "{uploadId}")
	s3Router.HandleFunc("/{bucket}/{key:.*}", h.AbortMultipartUpload).Methods("DELETE").Queries("uploadId", "{uploadId}")

	// CORS middleware
	router.Use(corsMiddleware)

	return router
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, HEAD, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, Cache-Control, X-Requested-With, x-amz-*")
		w.Header().Set("Access-Control-Expose-Headers", "ETag, x-amz-*")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
