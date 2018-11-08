package main

import (
	"context"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/services/keyvault/2016-10-01/keyvault"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/azure/auth"
)

const (
	getSecretRetires      = 10
	cosmosDBURISecretName = "cosmosDBConnectionString"
)

func main() {
	vaultName, ok := os.LookupEnv("VAULT_NAME")
	if !ok {
		log.Fatal("VAULT_NAME must be set.")
	}

	clientID, ok := os.LookupEnv("MSI_CLIENTID")
	if !ok {
		log.Fatal("MSI_CLIENTID must be set.")
	}

	keyClient, err := NewKeyVaultClient(vaultName, clientID)
	if err != nil {
		log.Fatal(err)
	}

	count := 0
	var dbURI string
	for {
		dbURI, err = keyClient.GetSecret(cosmosDBURISecretName)
		if err != nil {
			if count > getSecretRetires {
				log.Fatalf("Failed to get secret within retries with err: %v", err)
			}

			log.Printf("Retrying GetSecret: %d", count)
			count++

			time.Sleep(time.Second)
			continue
		}

		log.Println("Got DBURI")
		break
	}

	db := NewDB(dbURI, "users")
	users, err := db.GetUsers()
	if err != nil {
		log.Fatal(err)
	}

	// if theres no users in the DB, generate some and add them in
	if len(users) == 0 {
		err := db.PopulateWithUsers(10)
		if err != nil {
			log.Fatal(err)
		}

		users, err = db.GetUsers()
		if err != nil {
			log.Fatal(err)
		}
	}

	tmpl := template.Must(template.ParseFiles("index.html"))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		data := IndexPageData{
			PageTitle: "All the Users",
			Users:     users,
		}
		tmpl.Execute(w, data)
	})

	log.Println("Serving on port 80")
	log.Fatal(http.ListenAndServe("0.0.0.0:80", nil))
}

// IndexPageData holds the data to populate index.html
type IndexPageData struct {
	PageTitle string
	Users     []User
}

// KeyVault holds the information for a keyvault instance
type KeyVault struct {
	client   *keyvault.BaseClient
	vaultURL string
}

// NewKeyVaultClient creates a new keyvault client
func NewKeyVaultClient(vaultName, clientID string) (*KeyVault, error) {
	msiKeyConfig := &auth.MSIConfig{
		Resource: strings.TrimSuffix(azure.PublicCloud.KeyVaultEndpoint, "/"),
		ClientID: clientID,
	}

	auth, err := msiKeyConfig.Authorizer()
	if err != nil {
		return nil, err
	}

	keyClient := keyvault.New()
	keyClient.Authorizer = auth

	k := &KeyVault{
		vaultURL: fmt.Sprintf("https://%s.%s", vaultName, azure.PublicCloud.KeyVaultDNSSuffix),
		client:   &keyClient,
	}

	return k, nil
}

// GetSecret retrieves a secret from keyvault
func (k *KeyVault) GetSecret(keyName string) (string, error) {
	ctx := context.Background()

	keyBundle, err := k.client.GetSecret(ctx, k.vaultURL, keyName, "")
	if err != nil {
		return "", err
	}

	return *keyBundle.Value, nil
}
