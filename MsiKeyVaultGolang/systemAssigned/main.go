package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/Azure/go-autorest/autorest/azure/auth"

	"github.com/Azure/azure-sdk-for-go/services/keyvault/2016-10-01/keyvault"
	"github.com/Azure/go-autorest/autorest/azure"
)

func main() {
	vaultName, ok := os.LookupEnv("KEYVAULT_VAULT_NAME")
	if !ok {
		log.Fatal("KEYVAULT_VAULT_NAME must be set.")
	}

	secretName, ok := os.LookupEnv("KEYVAULT_SECRET_NAME")
	if !ok {
		log.Fatal("KEYVAULT_SECRET_NAME must be set.")
	}

	keyClient, err := NewKeyVaultClient(vaultName)
	if err != nil {
		log.Fatal(err)
	}

	secret, err := keyClient.GetSecret(secretName)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Retrieved secret %s from keyvault using MSI", secret)
}

type KeyVault struct {
	client   *keyvault.BaseClient
	vaultURL string
}

func NewKeyVaultClient(vaultName string) (*KeyVault, error) {
	msiKeyConfig := &auth.MSIConfig{
		Resource: strings.TrimSuffix(azure.PublicCloud.KeyVaultEndpoint, "/"),
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

func (k *KeyVault) GetSecret(keyName string) (string, error) {
	ctx := context.Background()

	keyBundle, err := k.client.GetSecret(ctx, k.vaultURL, keyName, "")
	if err != nil {
		return "", err
	}

	return *keyBundle.Value, nil
}
