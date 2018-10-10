package serve

import (
	"fmt"
	"io"
	"log"
	"net/http"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var clientset *kubernetes.Clientset

func Serve() {
	//        if vaultToken = os.Getenv("VAULT_TOKEN"); vaultToken == "" {
	//                log.Fatal("VAULT_TOKEN must be set and non-empty")
	//        }
	//        if os.Getenv("VAULT_WRAP_TTL") == "" {
	//                os.Setenv("VAULT_WRAP_TTL", "120")
	//        }
	//        if vaultAddr = os.Getenv("VAULT_ADDR") == "" {
	//                log.Fatal("VAULT_ADDR must be set and non-empty")
	//        }
	//
	//	config := api.DefaultConfig()
	//	vaultClient, err := api.NewClient(config)
	kubeconfig, err := rest.InClusterConfig()
	if err != nil {
		log.Println("ERROR:", err)
		panic(err.Error())
	}

	clientset, err = kubernetes.NewForConfig(kubeconfig)
	if err != nil {
		log.Println("ERROR:", err)
		log.Fatal(err)
	}

	log.Println("Starting server on 80 port")
	http.Handle("/secret", handler{secretRequestHandler})

	// Start server
	log.Fatal(http.ListenAndServe(":80", nil))
}

type handler struct {
	f func(io.Writer, *http.Request) (int, error)
}

func (h handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	code, err := h.f(w, r)
	w.WriteHeader(code)
	if err != nil {
		log.Printf("%v", err)
		fmt.Fprintf(w, "%v", err)
	}
}
