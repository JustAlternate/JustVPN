package routes

import (
	"net/http"
)

func GetHealth(w http.ResponseWriter, r *http.Request){
	_, _ = w.Write([]byte("ok"))
}
