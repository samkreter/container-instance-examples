package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"strings"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/Azure/azure-sdk-for-go/services/keyvault/2016-10-01/keyvault"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/azure/auth"
)

func main() {
	// vaultName, ok := os.LookupEnv("KEYVAULT_VAULT")
	// if !ok {
	// 	log.Fatal("KEYVAULT_VAULT must be set.")
	// }

	// secretName, ok := os.LookupEnv("KEYVAULT_SECRET")
	// if !ok {
	// 	log.Fatal("KEYVAULT_SECRET must be set.")
	// }

	// clientID := os.Getenv("MSI_USER_ASSIGNED_CLIENTID")

	// keyClient, err := NewKeyVaultClient(vaultName, clientID)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// dbURI, err := keyClient.GetSecret(secretName)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// get users

	tmpl := template.Must(template.ParseFiles("index.html"))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		data := IndexPageData{
			PageTitle: "All the Users",
			Users: []User{
				{Name: "Task 1", Email: "test@gmail.com"},
				{Name: "Task 2", Email: "test2@gmail.com"},
				{Name: "Task 3", Email: "test3@gmail.com"},
			},
		}
		tmpl.Execute(w, data)
	})

	log.Println("Serving on port 8080")
	log.Fatal(http.ListenAndServe("0.0.0.0:8080", nil))
}

type IndexPageData struct {
	PageTitle string
	Users     []User
}

type User struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type DB struct {
	connURI   string
	Container string
}

func NewDB(connURI string) *DB {
	return &DB{
		connURI: connURI,
	}
}

func (db *DB) getConn() *mgo.Session {
	dialInfo, err := mgo.ParseURL(db.connURI)
	if err != nil {
		log.Fatal(err)
	}

	// //Below part is similar to above.
	dialInfo.DialServer = func(addr *mgo.ServerAddr) (net.Conn, error) {
		return tls.Dial("tcp", addr.String(), &tls.Config{})
	}

	session, err := mgo.DialWithInfo(dialInfo)
	if err != nil {
		log.Fatal(err)
	}

	session.SetSafe(&mgo.Safe{})

	return session
}

func (db *DB) InsertUsers(users []User) error {
	session := db.getConn()
	defer session.Close()

	c := session.DB(db.Container).C(db.Container)

	log.Println("Adding Recored to Databases")

	errOccured := false

	for _, user := range users {
		//Container started the work
		err := c.Insert(user)
		if err != nil {
			errOccured = true
			log.Println("Error while inserting user: '%s' error: %v", err)
		}
	}

	if errOccured {
		return fmt.Errorf("1 or more errors occured while inserting users into db")
	}

	return nil
}

func (db *DB) GetUsers() ([]User, error) {
	session := db.getConn()
	defer session.Close()

	c := session.DB(db.Container).C(db.Container)

	log.Println("Getting Users from Databases")

	var users []User
	err := c.Find(bson.M{}).All(&users)
	if err != nil {
		return nil, err
	}

	return users, nil
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
