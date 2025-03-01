package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"JustVPN/routes"

	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
)

// authMiddleware validates the JWT token
func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get the token from the Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Unauthorized: No token provided", http.StatusUnauthorized)
			return
		}

		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || strings.ToLower(tokenParts[0]) != "bearer" {
			http.Error(w, "Unauthorized: Invalid token format", http.StatusUnauthorized)
			return
		}
		tokenString := tokenParts[1]

		// Parse and validate the token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(os.Getenv("JWT_SECRET")), nil
		})

		if err != nil {
			http.Error(w, fmt.Sprintf(`{"error": "Unauthorized: %v"}`, err), http.StatusUnauthorized)
			return
		}

		// Check if the token is valid
		if !token.Valid {
			http.Error(w, `{"error": "Unauthorized: Invalid token"}`, http.StatusUnauthorized)
			return
		}

		// Verify the token's expiration time
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, `{"error": "Unauthorized: Invalid token claims"}`, http.StatusUnauthorized)
			return
		}

		exp, err := claims.GetExpirationTime()
		if err != nil || exp == nil {
			http.Error(w, `{"error": "Unauthorized: Missing or invalid expiration time"}`, http.StatusUnauthorized)
			return
		}

		if exp.Before(time.Now()) {
			http.Error(w, `{"error": "Unauthorized: Token expired"}`, http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found, using existing environment variables")
	}

	// Initialize routes with environment variables
	routes.Initialize()

	// Set up HTTP handlers
	http.Handle("/start", authMiddleware(http.HandlerFunc(routes.GetStart)))
	http.Handle("/init", http.HandlerFunc(routes.InitSession))
	http.HandleFunc("/health", routes.GetHealth)
	http.HandleFunc("/login", routes.Login)
	http.HandleFunc("/ws", routes.ServeWs)

	// Get port from environment variable or use default
	port := os.Getenv("API_PORT")
	if port == "" {
		port = "8081"
	}

	log.Printf("Starting server on port %s...\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
