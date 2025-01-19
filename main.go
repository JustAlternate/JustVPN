package main

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"strconv"
	"slices"
	"time"

	"JustVPN/src/terraform"
)

func getHealth(w http.ResponseWriter, r *http.Request){
	_, _ = w.Write([]byte("ok"))
}

func getStart(w http.ResponseWriter, r *http.Request) {
	log.Println("Creating TerraformService and Init...")
	terraformService := terraform.NewTerraformService()

	log.Println("Parsing Response information...")
	err := r.ParseForm()
	if err != nil {
		log.Fatalf("Error parsing the response: %v", err)
	}
	ip := r.Form.Get("IP")
	parsedIp := net.ParseIP(ip)
	if parsedIp == nil {
		log.Fatalf("Not a valid ip address.")
	}
	parsedIpString := parsedIp.String()
	timeWantedBeforeDeletion := r.Form.Get("timeWantedBeforeDeletion")
	timeBeforeDeletion, err := strconv.Atoi(timeWantedBeforeDeletion)
	if err != nil {
		log.Fatalf("Not a valid time wanted address: %v", err)
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
		log.Fatalf("Not a valid region")
	}

	log.Printf("Terraform Apply for %s %s...\n", parsedIpString, region)
	err = terraformService.Apply(parsedIpString, region)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Getting hostIp for %s %s...\n", parsedIpString, region)
	hostIp, err := terraformService.GetOutput()
if err != nil {
		log.Fatalf("Error when retrieving the host ip: %s", err)
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
		log.Fatalf("Error when retrieving the pub key: %s", err)
	}


	log.Printf("Creating the response for %s %s...\n", parsedIpString, region)
	response := map[string]string{
			"host_endpoint": hostIp,
			"public_key":    pubkey,
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	json.NewEncoder(w).Encode(response)

	log.Printf("Launching timer before destroy for %s %s...\n", parsedIpString, region)
	go terraformService.Destroy(parsedIpString, timeBeforeDeletion)
}

func main() {
	http.HandleFunc("/start", getStart)
	http.HandleFunc("/health", getHealth)
	log.Fatal(http.ListenAndServe(":8081", nil))
}
