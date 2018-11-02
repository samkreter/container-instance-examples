package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/samkreter/container-instance-examples/Go/MsiSystemAssigned/azstorage"
)

const (
	getBlobRetryCount = 30
)

func main() {
	subID := getEnv("SUBID")
	resourceGroup := getEnv("RESOURCE_GROUP")
	storageAccountName := getEnv("ACCOUNT_NAME")

	azStorage, err := azstorage.NewClient(storageAccountName, resourceGroup, subID, "")
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()

	count := 0
	var blobContents string
	for {
		blobContents, err = azStorage.GetBlob(ctx, "democontainer", "kubernetes-acsiiart.txt")
		if err == nil {
			break
		}

		if count >= getBlobRetryCount {
			log.Fatal("Exceded retry attempts")
		}

		log.Println("Retrying get blob")
		time.Sleep(time.Second * 3)

		count++
	}

	log.Println("Download Blob Contents:")
	time.Sleep(time.Second * 1)
	fmt.Println(blobContents)

	blocker := make(chan struct{})
	<-blocker
}

func getEnv(envName string) string {
	val, ok := os.LookupEnv(envName)
	if !ok {
		log.Fatalf("%s must be set.", envName)
	}

	return val
}
