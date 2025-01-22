package routes

import (
	"encoding/json"
	"os"
	"net/http"
	"time"
	"golang.org/x/crypto/bcrypt"
	"github.com/golang-jwt/jwt/v5"
)

type User struct {
    Username string `json:"username"`
    Password string `json:"password"`
}

func readUsersFromFile(filePath string) ([]User, error) {
    file, err := os.ReadFile(filePath)
    if err != nil {
        return nil, err
    }

    var data struct {
        Users []User `json:"users"`
    }
    if err := json.Unmarshal(file, &data); err != nil {
        return nil, err
    }

    return data.Users, nil
}

func Login(w http.ResponseWriter, r *http.Request) {
    var credentials struct {
        Username string `json:"username"`
        Password string `json:"password"`
    }
    if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }

    users, err := readUsersFromFile("./users.json")
    if err != nil {
        http.Error(w, "Internal server error", http.StatusInternalServerError)
        return
    }

    for _, user := range users {
        if user.Username == credentials.Username {
            if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(credentials.Password)) == nil {
                token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
                    "username": credentials.Username,
                    "exp":      time.Now().Add(time.Hour * 24).Unix(), // Token expires in 24 hours
                })
                tokenString, err := token.SignedString([]byte(os.Getenv("SSH_PASSWORD")))
                if err != nil {
                    http.Error(w, "Internal server error", http.StatusInternalServerError)
                    return
                }

                w.Header().Set("Content-Type", "application/json")
                json.NewEncoder(w).Encode(map[string]string{"token": tokenString})
                return
            }
        }
    }

    http.Error(w, "Invalid credentials", http.StatusUnauthorized)
}
