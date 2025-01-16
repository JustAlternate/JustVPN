package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strconv"
)

func TerraformRessource(){
	cmd := exec.Command("terraform apply", fmt.Sprintf("-var=\"endpoint=%s\" -var-file=\"secrets.tfvars\" --auto-approve"))
	cmd.Stdout = os.Stdout
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
}

func getStart(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "Creating the ressource...\n")
	ip := r.Form.Get("IP")
	parsedIp := net.ParseIP(ip)
	if parsedIp == nil {
		io.WriteString(w, "Not a valid ip address.\n")
		return
	}
	timeWantedBeforeDeletion := r.Form.Get("timeWantedBeforeDeletion")
	parsedTimeBeforeDeletion, err := strconv.Atoi(timeWantedBeforeDeletion)
	if err != nil {
		io.WriteString(w, "Not a valid time wanted address.\n")
		return
	}
}

func main() {
	http.HandleFunc("/start", getStart)
	log.Fatal(http.ListenAndServe(":8081", nil))
}
