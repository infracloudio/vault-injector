package serve

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/infracloudio/vault-injector/controller/pkg/vaultClient"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type secretReq struct {
	SecretName     string
	Namespace      string
	VaultPath      string
	VaultKey       string
	ServiceAccount string
	Role           string
}

func getServiceAccountSecret(sa, ns string) (string, error) {
	saClient := clientset.CoreV1().ServiceAccounts(ns)
	sacc, err := saClient.Get(sa, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	log.Printf("Service account:: %+v\n", sacc)
	return sacc.Secrets[0].Name, nil
}

func getSAToken(saSecret, ns string) (token string, err error) {
	saSecretClient := clientset.CoreV1().Secrets(ns)
	secret, err := saSecretClient.Get(saSecret, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	return string(secret.Data["token"]), err
}

func createK8SSecret(name, key, value, ns string) (err error) {

	saSecretClient := clientset.CoreV1().Secrets(ns)
	newSecret := v1.Secret{
		TypeMeta:   metav1.TypeMeta{Kind: "Secret", APIVersion: "v1"},
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: ns},
		StringData: map[string]string{key: value},
	}
	_, err = saSecretClient.Create(&newSecret)
	if err != nil {
		return err
	}
	return nil
}

func secretRequestHandler(w io.Writer, r *http.Request) (int, error) {
	log.Printf("received secret request from %s\n", r.RemoteAddr)
	req := secretReq{}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return 400, fmt.Errorf("missing or empty name parameter from %s", r.RemoteAddr)
	}

	log.Printf("received payload: %+v\n", req)

	// get service account secret
	log.Println("getting service account secret")
	saSecret, err := getServiceAccountSecret(req.ServiceAccount, req.Namespace)
	if err != nil {
		return 400, fmt.Errorf("error in getting service account %s in %s ns, error=%s\n", req.ServiceAccount, req.Namespace, err.Error())
	}

	// get service account jwt token
	log.Println("getting service account token")
	saToken, err := getSAToken(saSecret, req.Namespace)
	if err != nil {
		return 400, fmt.Errorf("error in getting service account %s in %s ns, error=%s\n", req.ServiceAccount, req.Namespace, err.Error())
	}

	// get auth token from vault
	log.Println("getting vault token")
	vaultToken, err := vaultClient.VaultAuthToken(saToken, req.Role)
	if err != nil {
		log.Println("ERROR:", err)
		return 400, fmt.Errorf("error in getting vault auth token: %s\n", err)
	}

	// get secret from vault
	log.Println("getting vault secret")
	secretValue, err := vaultClient.GetVaultSecret(req.VaultPath, req.VaultKey, vaultToken)
	if err != nil {
		return 400, fmt.Errorf("error in getting secret from vault: %s\n", err)
	}

	// create k8s secret with value
	log.Println("generating k8s secret")
	err = createK8SSecret(req.SecretName, req.VaultKey, secretValue, req.Namespace)
	if err != nil {
		return 400, fmt.Errorf("error in getting generating k8s secret: %s\n", err)
	}
	log.Println("Done!!")
	return 200, nil
}
