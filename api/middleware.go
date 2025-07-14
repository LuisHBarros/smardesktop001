package api

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

// CORSMiddleware adiciona headers CORS permitindo apenas localhost e 127.0.0.1
func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		
		// Permite apenas origens localhost e 127.0.0.1
		if strings.HasPrefix(origin, "http://localhost") || strings.HasPrefix(origin, "http://127.0.0.1") {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}
		
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		
		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		
		next.ServeHTTP(w, r)
	})
}

// LoggingMiddleware registra todas as requisições
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		// Cria um ResponseWriter customizado para capturar o status code
		lrw := &loggingResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		
		next.ServeHTTP(lrw, r)

		duration := time.Since(start)

		// Gera log apenas para rotas específicas da API
		apiRoutes := []string{"/status", "/escreve_arquivo", "/move_arquivo", "/executar_terceiros"}
		shouldLog := false
		for _, route := range apiRoutes {
			if r.RequestURI == route || strings.HasPrefix(r.RequestURI, route+"?") {
				shouldLog = true
				break
			}
		}

		if shouldLog {
			logMessage := fmt.Sprintf("[%s] %s %s - %d - %v",
				r.Method,
				r.RequestURI,
				r.RemoteAddr,
				lrw.statusCode,
				duration)

			log.Print(logMessage)
			AddLogEntry(logMessage)
		}
	})
}

// loggingResponseWriter é um wrapper para capturar o status code
type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}
