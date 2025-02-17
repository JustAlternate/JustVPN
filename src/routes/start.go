package routes

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"
	"slices"
	"time"

	"JustVPN/src/terraform"
)

type LogWriter struct {
	w http.ResponseWriter
	flusher http.Flusher
}

func (lw *LogWriter) Write(p []byte) (n int, err error) {
	data := struct {
		Type string `json:"type"`
		Message string `json:"message"`
	}{
		Type: "log",
		Message: string(p),
	}
	
	jsonData, err := json.Marshal(data)
	if err != nil {
		return 0, err
	}
	
	fmt.Fprintf(lw.w, "data: %s\n\n", jsonData)
	lw.flusher.Flush()
	return len(p), nil
}

func GetStart(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	// Set headers for SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	// Create custom log writer
	logWriter := &LogWriter{w: w, flusher: flusher}
	log.SetOutput(logWriter)

	terraformService := terraform.NewTerraformService()
	log.Println("Creating TerraformService and Init...")

	log.Println("Parsing Response information...")
	err := r.ParseForm()
	if err != nil {
		log.Printf("Error parsing the response: %v\n", err)
		return
	}

	ip := r.Form.Get("IP")
	parsedIp := net.ParseIP(ip)
	if parsedIp == nil {
		log.Println("Not a valid ip address.")
		return
	}
	parsedIpString := parsedIp.String()
	
	timeWantedBeforeDeletion := r.Form.Get("timeWantedBeforeDeletion")
	timeBeforeDeletion, err := strconv.Atoi(timeWantedBeforeDeletion)
	if err != nil {
		log.Printf("Not a valid time wanted address: %v\n", err)
		return
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
	if !slices.Contains(availableRegion, region) {
		log.Println("Not a valid region")
		return
	}

	log.Printf("Terraform Apply for %s %s...\n", parsedIpString, region)
	err = terraformService.Apply(parsedIpString, region)
	if err != nil {
		log.Printf("Error: %v\n", err)
		return
	}

	log.Printf("Getting hostIp for %s %s...\n", parsedIpString, region)
	hostIp, err := terraformService.GetOutput()
	if err != nil {
		log.Printf("Error when retrieving the host ip: %s\n", err)
		return
	}

	log.Printf("Getting PubKey for %s %s...\n", parsedIpString, region)
	pubkey, err := terraformService.GetPubKey(hostIp)
	maxRetries := 5

	for attempt := 1; attempt <= maxRetries || maxRetries == 0; attempt++ {
		if err == nil {
			break
		}
		log.Printf("Attempt %d: Error occurred when fetching pubkey, retrying in 10 seconds...\n", attempt)
		time.Sleep(10 * time.Second)
		pubkey, err = terraformService.GetPubKey(hostIp)
	}

	if err != nil {
		log.Printf("Error when retrieving the pub key: %s\n", err)
		return
	}

	log.Printf("Creating the response for %s %s...\n", parsedIpString, region)
	response := map[string]string{
		"host_endpoint": hostIp,
		"public_key":    pubkey,
	}

	// Send final response
	finalResponse := struct {
		Type string `json:"type"`
		Data map[string]string `json:"data"`
	}{
		Type: "result",
		Data: response,
	}
	
	jsonData, err := json.Marshal(finalResponse)
	if err != nil {
		log.Printf("Error encoding final response: %v\n", err)
		return
	}
	
	fmt.Fprintf(w, "data: %s\n\n", jsonData)
	flusher.Flush()

	log.Printf("Launching timer before destroy for %s %s...\n", parsedIpString, region)
	go terraformService.Destroy(parsedIpString, timeBeforeDeletion)
}
