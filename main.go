package main

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/joho/godotenv"
	"github.com/tablelandnetwork/basin-cli/pkg/signing"
)

type Event struct {
	CID         string `json:"cid"`
	Timestamp   int64  `json:"timestamp"`
	IsArchived  bool   `json:"is_archived"`
	CacheExpiry string `json:"cache_expiry"`
}

// createVault creates a vault via the Basin API.
func createVault(vaultID, account string, cache *int) error {
	data := url.Values{}
	data.Set("account", account)
	if cache != nil {
		data.Set("cache", fmt.Sprintf("%d", *cache))
	}

	url := "https://basin.tableland.xyz/vaults/" + vaultID
	resp, err := http.Post(url, "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
	if err != nil {
		fmt.Printf("Error creating vault: %v\n", err)
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	fmt.Printf("Create response: %s\n", body)
	return nil
}

// writeEvent writes a file to a vault via the Basin API.
func writeEvent(vaultId, filename, timestamp, signature string) error {
	url := fmt.Sprintf("https://basin.tableland.xyz/vaults/%s/events?timestamp=%s&signature=%s", vaultId, timestamp, signature)

	file, err := os.Open(filename)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		return err
	}
	defer file.Close()

	fileData, err := io.ReadAll(file)
	if err != nil {
		fmt.Printf("Error reading file: %v\n", err)
		return err
	}

	req, err := http.NewRequest("POST", url, bytes.NewReader(fileData))
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		return err
	}

	req.Header.Set("filename", filename)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error sending request: %v\n", err)
		return err
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	fmt.Printf("Write response: %s\n", responseBody)
	return nil
}

// listEvents lists the events in a vault via the Basin API.
func listEvents(vaultId string) ([]Event, error) {
	url := fmt.Sprintf("https://basin.tableland.xyz/vaults/%s/events", vaultId)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to list events: %s, response: %s", resp.Status, string(bodyBytes))
	}

	var events []Event
	if err := json.NewDecoder(resp.Body).Decode(&events); err != nil {
		return nil, err
	}

	return events, nil
}

// downloadEvent downloads an event from the cache and saves it to a file.
func downloadEvent(eventID, outputPath string) error {
	url := fmt.Sprintf("https://basin.tableland.xyz/events/%s", eventID)

	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("error creating file: %v", err)
	}
	defer outFile.Close()

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("error making GET request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("received non-200 response status: %d %s", resp.StatusCode, resp.Status)
	}

	_, err = io.Copy(outFile, resp.Body)
	if err != nil {
		return fmt.Errorf("error writing to file: %v", err)
	}

	fmt.Println("Event downloaded successfully")
	return nil
}

// getPubKey returns the public key from a private key
func getPubKey(privateKey *ecdsa.PrivateKey) (string, error) {
	pubKey := privateKey.Public()
	pubKeyECDSA, ok := pubKey.(*ecdsa.PublicKey)
	if !ok {
		fmt.Println("Error casting public key to ECDSA")
		return "", errors.New("error casting public key to ECDSA")
	}

	address := crypto.PubkeyToAddress(*pubKeyECDSA)
	return address.Hex(), nil
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
	pk := os.Getenv("PRIVATE_KEY")
	vaultId := os.Getenv("VAULT_ID")

	// Set up our private key, account/address, and signer
	privateKey, _ := signing.LoadPrivateKey(pk)
	account, _ := getPubKey(privateKey)
	signer := signing.NewSigner(privateKey)

	// Set up our file to sign/write to a vault
	filePath := "test.txt"
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("Error loading reading file: %v", err)
	}
	filename := file.Name()
	defer file.Close()

	// Sign the file
	signature, err := signer.SignFile(filename)
	if err != nil {
		log.Fatalf("Error signing file: %v", err)
	}
	fmt.Printf("Signature: %v\n", signature)

	// Create a test vault via API
	fmt.Printf("Creating vault '%s' for account: %s\n", vaultId, account)
	cache := 10800 // minutes
	err = createVault(vaultId, account, &cache)
	if err != nil {
		log.Fatalf("Error creating vault: %v", err)
	}

	// Write an event to the vault
	fmt.Printf("Writing to vault '%s'\n", vaultId)
	timestamp := time.Now().Unix()
	timestampStr := strconv.FormatInt(timestamp, 10)
	err = writeEvent(vaultId, filename, timestampStr, signature)
	if err != nil {
		log.Fatalf("Error writing event: %v", err)
	}

	// Wait a couple of seconds for the event to be written
	time.Sleep(5 * time.Second)

	// List the events in the vault
	fmt.Printf("Getting vault '%s' events\n", vaultId)
	events, err := listEvents(vaultId)
	if err != nil {
		log.Fatalf("Error listing events: %v", err)
	}
	// Log or process the events
	fmt.Print("Events:\n")
	for _, event := range events {
		fmt.Printf("  CID: %s\n  Timestamp: %d\n  IsArchived: %t\n  CacheExpiry: %s\n", event.CID, event.Timestamp, event.IsArchived, event.CacheExpiry)
	}

	// Download the first event from the cache to a file
	cid := events[0].CID
	fmt.Printf("Downloading event '%s'\n", events[0].CID)
	outputPath := "test-download.txt"
	if err := downloadEvent(cid, outputPath); err != nil {
		fmt.Printf("Error downloading event: %v\n", err)
	}
}
