package vaultClient

import (
	"encoding/json"
	"log"
	"net/http"
)

type vaultSecretValue struct {
	Value string
}

func GetVaultSecret(path, key, token string) (value string, err error) {
	//	body := strings.NewReader(`{"jwt": "your_service_account_jwt", "role": "demo"}`)
	req, err := http.NewRequest("GET", "http://"+vaultAddr+":8200/v1/"+path+"/"+key, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("X-Vault-Token", token)
	log.Printf("Vault request %+v\n", req)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	vaultResp := vaultSecretValue{}
	json.NewDecoder(resp.Body).Decode(&vaultResp)

	log.Printf("Vault Secret Request Response %+v\n", vaultResp)
	return vaultResp.Value, nil
}
