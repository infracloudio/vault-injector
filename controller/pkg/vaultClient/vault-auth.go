package vaultClient

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/hashicorp/vault/api"
)

var (
	vaultClient *api.Client
	vaultAddr   string
)

type vaultAuthReq struct {
	JWT  string `json:"jwt"`
	Role string `json:"role"`
}

type vaultAuthResp struct {
	Auth `json:"auth"`
}

type Auth struct {
	ClientToken string   `json:"client_token"`
	Policies    []string `json:"policies"`
}

func init() {
	if vaultToken := os.Getenv("VAULT_TOKEN"); vaultToken == "" {
		log.Fatal("VAULT_TOKEN must be set and non-empty")
	}

	vaultAddr = os.Getenv("VAULT_ADDR")
	if vaultAddr == "" {
		log.Fatal("VAULT_ADDR must be set and non-empty")
	}

	if os.Getenv("VAULT_WRAP_TTL") == "" {
		os.Setenv("VAULT_WRAP_TTL", "120")
	}

	var err error
	config := api.DefaultConfig()
	vaultClient, err = api.NewClient(config)
	if err != nil {
		log.Fatal(err)
	}
}

func VaultAuthToken(saToken, role string) (vaultToken string, err error) {
	vaultReq := &vaultAuthReq{
		JWT:  saToken,
		Role: role,
	}
	body, err := json.Marshal(vaultReq)
	if err != nil {
		log.Printf("ERROR: marshalling vault request failed\n")
	}
	//	body := strings.NewReader(`{"jwt": "your_service_account_jwt", "role": "demo"}`)
	req, err := http.NewRequest("POST", "http://"+vaultAddr+":8200/v1/auth/kubernetes/login", bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	vaultResp := vaultAuthResp{}
	json.NewDecoder(resp.Body).Decode(&vaultResp)

	log.Printf("Vault Token Request Response %+v\n", vaultResp)
	return vaultResp.ClientToken, nil
}
