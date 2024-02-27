package main

import (
	"bytes"
	"crypto/ecdsa"
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

// Create a vault via the Basin API
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

// Write a file to a vault via the Basin API
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

// List the events in a vault via the Basin API
func listEvents(vaultId string) (string, error) {
	url := fmt.Sprintf("https://basin.tableland.xyz/vaults/%s/events", vaultId)

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

// Get the public key from a private key
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
	fmt.Printf("Signature: %v", signature)

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
	fmt.Printf("Events: %v", events)
}
