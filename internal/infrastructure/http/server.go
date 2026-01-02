package http

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/bsko/casino-transaction-system/internal/config"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type HttpServer struct {
	postTransactionsMessageHandler postTransactionsMessageHandler
	server                         *http.Server
	port                           int
}

func NewHttpServer(postTransactionsMessageHandler postTransactionsMessageHandler, conf *config.Http) *HttpServer {
	return &HttpServer{
		postTransactionsMessageHandler: postTransactionsMessageHandler,
		port:                           conf.Port,
	}
}

func (s *HttpServer) Start(ctx context.Context) error {
	router := chi.NewRouter()

	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.RequestID)
	router.Use(middleware.Timeout(60 * time.Second))

	router.Get("/health", s.handleHealthCheck)
	router.Post("/transactions", s.handlePostTransactions)

	s.server = &http.Server{
		Addr:         fmt.Sprintf(":%d", s.port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Printf("Starting HTTP server on port %d", s.port)

	if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("failed to start HTTP server: %w", err)
	}

	return nil
}

func (s *HttpServer) Shutdown(ctx context.Context) error {
	if s.server == nil {
		return nil
	}

	log.Println("Shutting down HTTP server...")
	shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := s.server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("failed to shutdown HTTP server: %w", err)
	}

	log.Println("HTTP server stopped successfully")
	return nil
}

func (s *HttpServer) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := map[string]string{
		"status":    "healthy",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Failed to encode health check response: %v", err)
	}
}

func (s *HttpServer) handlePostTransactions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req TransactionSearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}
	defer func() { _ = r.Body.Close() }()

	filter, err := TransformRequestToFilter(req)
	if err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid filter parameters", err.Error())
		return
	}

	transactions, err := s.postTransactionsMessageHandler.GetListByFilter(filter)
	if err != nil {
		log.Printf("Failed to get transactions: %v", err)
		s.writeError(w, http.StatusInternalServerError, "Failed to retrieve transactions", "")
		return
	}

	response := TransformTransactionsToResponse(transactions, filter.Limit, filter.Offset)

	w.WriteHeader(http.StatusOK)
	if err = json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Failed to encode response: %v", err)
	}
}

func (s *HttpServer) writeError(w http.ResponseWriter, statusCode int, error, message string) {
	w.WriteHeader(statusCode)
	errResponse := ErrorResponse{
		Error:   error,
		Message: message,
	}
	if err := json.NewEncoder(w).Encode(errResponse); err != nil {
		log.Printf("Failed to encode error response: %v", err)
	}
}
