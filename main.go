package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"JustVPN/src/routes"

	"github.com/golang-jwt/jwt/v5"
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

        // Extract the token (format: "Bearer <token>")
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
            return []byte("your_secret_key"), nil // Use the same secret key used to sign the token
        })

        if err != nil || !token.Valid {
            http.Error(w, "Unauthorized: Invalid token", http.StatusUnauthorized)
            return
        }

        // Token is valid, proceed to the next handler
        next.ServeHTTP(w, r)
    })
}

func main() {
	http.Handle("/start", authMiddleware(http.HandlerFunc(routes.GetStart)))
	http.HandleFunc("/health", routes.GetHealth)
	http.HandleFunc("/login", routes.Login)
	log.Fatal(http.ListenAndServe(":8081", nil))
}
