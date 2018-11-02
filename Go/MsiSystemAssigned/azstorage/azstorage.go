package azstorage

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/url"

	"github.com/Azure/azure-sdk-for-go/services/storage/mgmt/2017-06-01/storage"
	"github.com/Azure/azure-storage-blob-go/2016-05-31/azblob"
	"github.com/Azure/go-autorest/autorest/azure/auth"
)

var (
	blobFormatString = `https://%s.blob.core.windows.net`
)

// Client object to interact with azure storage
type Client struct {
	StorageAccountName   string
	ResourceGroupName    string
	SubscriptionID       string
	DefaultBlobName      string
	DefaultContainerName string
}

// NewClient creates a new client to interact with azure storage
func NewClient(storageAccountName, resourceGroupName, subscriptionID, defaultContainerName string) (*Client, error) {
	return &Client{
		StorageAccountName:   storageAccountName,
		ResourceGroupName:    resourceGroupName,
		SubscriptionID:       subscriptionID,
		DefaultContainerName: defaultContainerName,
	}, nil
}

// GetBlob downloads the specified blob contents
func (c *Client) GetBlob(ctx context.Context, containerName, blobName string) (string, error) {
	b := c.getBlobURL(ctx, containerName, blobName)

	resp, err := b.GetBlob(ctx, azblob.BlobRange{}, azblob.BlobAccessConditions{}, false)
	if err != nil {
		return "", err
	}
	defer resp.Body().Close()
	body, err := ioutil.ReadAll(resp.Body())
	return string(body), err
}

func (c *Client) getBlobURL(ctx context.Context, containerName, blobName string) azblob.BlobURL {
	container := c.getContainerURL(ctx, containerName)
	blob := container.NewBlobURL(blobName)
	return blob
}

func (c *Client) getContainerURL(ctx context.Context, containerName string) azblob.ContainerURL {
	key, err := c.getAccountPrimaryKey(ctx)
	if err != nil {
		log.Fatal(err)
	}

	cred := azblob.NewSharedKeyCredential(c.StorageAccountName, key)
	p := azblob.NewPipeline(cred, azblob.PipelineOptions{})

	// azblob.PipelineOptions{
	// 	Telemetry: azblob.TelemetryOptions{Value: config.UserAgent()},
	// }

	u, _ := url.Parse(fmt.Sprintf(blobFormatString, c.StorageAccountName))
	service := azblob.NewServiceURL(*u, p)
	container := service.NewContainerURL(containerName)
	return container
}

func (c *Client) getAccountPrimaryKey(ctx context.Context) (string, error) {
	accountsClient, err := c.getStorageAccountsClient()
	if err != nil {
		return "", err
	}

	result, err := accountsClient.ListKeys(ctx, c.ResourceGroupName, c.StorageAccountName)
	if err != nil {
		return "", err
	}

	return *(((*result.Keys)[0]).Value), nil
}

func (c *Client) getStorageAccountsClient() (*storage.AccountsClient, error) {
	storageAccountsClient := storage.NewAccountsClient(c.SubscriptionID)

	msiConfig := auth.NewMSIConfig()

	auth, err := msiConfig.Authorizer()
	if err != nil {
		return nil, err
	}

	storageAccountsClient.Authorizer = auth
	//storageAccountsClient.AddToUserAgent(config.UserAgent())
	return &storageAccountsClient, nil
}
