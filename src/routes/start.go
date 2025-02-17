package routes

import (
	"encoding/json"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"JustVPN/src/terraform"
	"github.com/gorilla/websocket"
	"github.com/google/uuid"
)

var LogChannel = make(chan string, 100)

var sessionChannels = make(map[string]chan string)
var sessionMu sync.Mutex

type channelWriter struct {
	ch chan string
}

func (cw channelWriter) Write(p []byte) (n int, err error) {
	msg := string(p)
	cw.ch <- msg
	return len(p), nil
}

func GetStart(w http.ResponseWriter, r *http.Request) {
	sessionID := uuid.New().String()
	ch := make(chan string, 100)
	sessionMu.Lock()
	sessionChannels[sessionID] = ch
	sessionMu.Unlock()

	mw := io.MultiWriter(os.Stdout, channelWriter{ch})
	logger := log.New(mw, "", log.LstdFlags)

	logger.Printf("GetStart: Starting request handling")
	logger.Println("Creating TerraformService and Init...")
	terraformService := terraform.NewTerraformService()

	logger.Println("Parsing Response information...")
	err := r.ParseForm()
	if err != nil {
		logger.Fatalf("Error parsing the response: %v", err)
	}
	ip := r.Form.Get("IP")
	parsedIp := net.ParseIP(ip)
	if parsedIp == nil {
		logger.Fatalf("Not a valid ip address.")
	}
	parsedIpString := parsedIp.String()
	timeWantedBeforeDeletion := r.Form.Get("timeWantedBeforeDeletion")
	timeBeforeDeletion, err := strconv.Atoi(timeWantedBeforeDeletion)
	if err != nil {
		logger.Fatalf("Not a valid time wanted address: %v", err)
	}

	region := r.Form.Get("region")
	availableRegion := []string{
		"eu-central",
		"fr-par",
		"gb-lon",
		"it-mil",
		"nl-ams",
		"se-sto",
		"us-central",
		"us-east",
		"us-west",
		"ca-central",
		"jp-osa",
		"jp-tyo-3",
		"au-mel",
		"br-gru",
	}
	if !slices.Contains(availableRegion, region){
		logger.Fatalf("Not a valid region")
	}

	logger.Printf("Terraform Apply for %s %s...\n", parsedIpString, region)
	err = terraformService.Apply(parsedIpString, region)
	if err != nil {
		logger.Fatal(err)
	}

	logger.Printf("Getting hostIp for %s %s...\n", parsedIpString, region)
	hostIp, err := terraformService.GetOutput()
	if err != nil {
		logger.Fatalf("Error when retrieving the host ip: %s", err)
	}

	logger.Printf("Getting PubKey for %s %s...\n", parsedIpString, region)
	pubkey, err := terraformService.GetPubKey(hostIp)
	maxRetries := 5

	for attempt := 1; attempt <= maxRetries || maxRetries == 0; attempt++ {
			if err == nil {
					break
			}
			logger.Printf("Attempt %d: Error occurred when fetching pubkey, retrying in 10 seconds...\n", attempt)
			time.Sleep(10 * time.Second)
			pubkey, err = terraformService.GetPubKey(hostIp)
	}

	if err != nil {
		logger.Fatalf("Error when retrieving the pub key: %s", err)
	}

	logger.Printf("Creating the response for %s %s...\n", parsedIpString, region)
	response := map[string]string{
			"host_endpoint": hostIp,
			"public_key":    pubkey,
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	json.NewEncoder(w).Encode(response)

	logger.Printf("Launching timer before destroy for %s %s...\n", parsedIpString, region)
	go terraformService.Destroy(parsedIpString, timeBeforeDeletion)

	logger.Printf("Finished handling request")

	go func(id string, c chan string) {
		time.Sleep(2 * time.Second)
		sessionMu.Lock()
		delete(sessionChannels, id)
		sessionMu.Unlock()
		close(c)
	}(sessionID, ch)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"sessionID": sessionID})
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func ServeWs(w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Query().Get("session")
	if sessionID == "" {
		http.Error(w, "Missing session id", http.StatusBadRequest)
		return
	}

	sessionMu.Lock()
	ch, ok := sessionChannels[sessionID]
	sessionMu.Unlock()
	if !ok {
		http.Error(w, "Session not found", http.StatusNotFound)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Failed to upgrade to WebSocket", http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	for msg := range ch {
		err := conn.WriteMessage(websocket.TextMessage, []byte(msg))
		if err != nil {
			break
		}
	}
}
