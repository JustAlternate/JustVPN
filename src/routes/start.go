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
	"slices"
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

// setCorsHeaders sets the CORS headers for the response
func setCorsHeaders(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
}

// InitSession creates a new session and returns its ID
func InitSession(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		setCorsHeaders(w)
		return
	}

	sessionID := uuid.New().String()
	ch := make(chan string, 100)
	
	sessionMu.Lock()
	sessionChannels[sessionID] = ch
	sessionMu.Unlock()

	// Set up automatic cleanup after 30 minutes
	go func() {
		time.Sleep(15 * time.Minute)
		sessionMu.Lock()
		if ch, exists := sessionChannels[sessionID]; exists {
			delete(sessionChannels, sessionID)
			close(ch)
		}
		sessionMu.Unlock()
	}()

	w.Header().Set("Content-Type", "application/json")
	setCorsHeaders(w)
	json.NewEncoder(w).Encode(map[string]string{"sessionID": sessionID})
}

func GetStart(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		setCorsHeaders(w)
		return
	}

	logger := log.New(os.Stdout, "", log.LstdFlags)
	logger.Printf("GetStart: Starting request handling")
	
	// Get session ID from form data
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Error parsing form data", http.StatusBadRequest)
		return
	}
	
	sessionID := r.Form.Get("sessionID")
	if sessionID == "" {
		http.Error(w, "No session ID provided", http.StatusBadRequest)
		return
	}

	// Get the channel for this session
	sessionMu.Lock()
	ch, exists := sessionChannels[sessionID]
	sessionMu.Unlock()
	
	if !exists {
		http.Error(w, "Invalid or expired session ID", http.StatusBadRequest)
		return
	}

	// Create a multi-writer to write logs to both stdout and the session channel
	mw := io.MultiWriter(os.Stdout, channelWriter{ch})
	logger = log.New(mw, "", log.LstdFlags)

	logger.Println("Creating TerraformService and Init...")
	terraformService := terraform.NewTerraformService()

	logger.Println("Parsing Response information...")
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
		logger.Print(err)
		return;
	}

	logger.Printf("Getting hostIp for %s %s...\n", parsedIpString, region)
	hostIp, err := terraformService.GetOutput()
	if err != nil {
		logger.Printf("Error when retrieving the host ip: %s", err)
		return;
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
		logger.Printf("Error when retrieving the pub key: %s", err)
		return;
	}

	logger.Printf("Creating the response for %s %s...\n", parsedIpString, region)
	response := map[string]string{
		"host_endpoint": hostIp,
		"public_key":    pubkey,
	}
	w.Header().Set("Content-Type", "application/json")
	setCorsHeaders(w)
	json.NewEncoder(w).Encode(response)

	logger.Printf("Launching timer before destroy for %s %s...\n", parsedIpString, region)
	go terraformService.Destroy(parsedIpString, timeBeforeDeletion)

	logger.Printf("Finished handling request")

	// Clean up the session after a delay
	go func(id string, c chan string) {
		time.Sleep(2 * time.Second)
		sessionMu.Lock()
		delete(sessionChannels, id)
		sessionMu.Unlock()
		close(c)
	}(sessionID, ch)
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// In production, you might want to restrict this
		return true
	},
	ReadBufferSize:  2048,
	WriteBufferSize: 2048,
}

func ServeWs(w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Query().Get("session")
	if sessionID == "" {
		http.Error(w, "Session ID is required", http.StatusBadRequest)
		return
	}

	sessionMu.Lock()
	ch, exists := sessionChannels[sessionID]
	sessionMu.Unlock()

	if !exists {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	// Set CORS headers for WebSocket handshake
	setCorsHeaders(w)

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Error upgrading to WebSocket: %v", err)
		return
	}
	defer ws.Close()

	// Send a connected message
	err = ws.WriteMessage(websocket.TextMessage, []byte("WebSocket connected successfully"))
	if err != nil {
		log.Printf("Error sending welcome message: %v", err)
		return
	}

	// Read from channel and write to WebSocket
	for msg := range ch {
		err := ws.WriteMessage(websocket.TextMessage, []byte(msg))
		if err != nil {
			log.Printf("Error writing to WebSocket: %v", err)
			break
		}
	}
}
