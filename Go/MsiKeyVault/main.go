package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/Azure/azure-sdk-for-go/services/keyvault/2016-10-01/keyvault"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/azure/auth"
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

	clientID := os.Getenv("MSI_USER_ASSIGNED_CLIENTID")

	keyClient, err := NewKeyVaultClient(vaultName, clientID)
	if err != nil {
		log.Fatal(err)
	}

	secret, err := keyClient.GetSecret(secretName)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Retrieved secret '%s' from keyvault using MSI", secret)
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

// type MSIConfig struct {
// 	Resource string
// 	ClientID string
// }

// func NewCustomMSIConfig(clientID string) *MSIConfig {
// 	return &MSIConfig{
// 		Resource: strings.TrimSuffix(azure.PublicCloud.KeyVaultEndpoint, "/"),
// 		ClientID: clientID,
// 	}
// }

// func (mc *MSIConfig) Authorizer() (autorest.Authorizer, error) {
// 	msiEndpoint, err := adal.GetMSIVMEndpoint()
// 	if err != nil {
// 		return nil, err
// 	}

// 	var spToken *adal.ServicePrincipalToken
// 	if mc.ClientID == "" {
// 		log.Println("Using System Assined Identity to get token")
// 		spToken, err = adal.NewServicePrincipalTokenFromMSI(msiEndpoint, mc.Resource)
// 		if err != nil {
// 			return nil, fmt.Errorf("failed to get oauth token for system Assigned MSI: %v", err)
// 		}
// 	} else {
// 		log.Println("Using User assigned Idenetity to get token.")
// 		spToken, err = adal.NewServicePrincipalTokenFromMSIWithUserAssignedID(msiEndpoint, mc.Resource, mc.ClientID)
// 		if err != nil {
// 			return nil, fmt.Errorf("failed to get oauth token for user assigned MSI: %v", err)
// 		}
// 	}

// 	return autorest.NewBearerAuthorizer(spToken), nil
// }
