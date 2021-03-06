package Handler

import (
	"azure-sdk-for-go/storage"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
)

// Can copy between local filesystem and Azure Blob Storage.
// Doesn't go across different Azure accounts yet.
type BlobHandler struct {

	// creds.
	accountName       string
	accountKey        string
	concurrentFactor  int
	blobStorageClient storage.BlobStorageClient
}

var wg sync.WaitGroup

// NewBlobHandler   create new instance of BlobHandler
func NewBlobHandler(accountName string, accountKey string, concurrentFactor int) (*BlobHandler, error) {
	bh := new(BlobHandler)

	client, err := storage.NewBasicClient(accountName, accountKey)

	if err != nil {
		return nil, err
	}

	bh.concurrentFactor = concurrentFactor
	bh.accountName = accountName
	bh.accountKey = accountKey
	bh.blobStorageClient = client.GetBlobService()
	return bh, nil
}

// Delete a blob
func (bh BlobHandler) Delete(containerName string, blobName string) error {
	return nil
}

// GenerateSASURLForBlob generates SAS URL for blob
func (bh BlobHandler) GenerateSASURLForBlob(containerName string, blobName string, durationInSeconds int, permissions string) (string, error) {
	container := bh.blobStorageClient.GetContainerReference(containerName)
	blob := container.GetBlobReference(blobName)

	//expiry := time.Now().UTC().Add(time.Second * time.Duration(durationInSeconds))
	expiry := time.Now().UTC().Add(time.Hour)

	log.Debugf("now %s", time.Now().UTC())
	log.Debugf("expiry %s", expiry)
	u, err := blob.GetSASURI(expiry, permissions)
	if err != nil {
		return "", err
	}

	return u, nil
}

// GenerateSASURLForContainer generates SAS URL for container
// Doesn't exist yet in SDK!
func (bh BlobHandler) GenerateSASURLForContainer(containerName string, durationInSeconds int) (string, error) {
	return "", nil
}

// CreateContainer creates a new container
func (bh BlobHandler) CreateContainer(containerName string) error {
	container := bh.blobStorageClient.GetContainerReference(containerName)

	_, err := container.CreateIfNotExists(nil)
	if err != nil {
		return err
	}
	return nil
}

// ListBlobsInContainer lists the blobs in a container
func (bh BlobHandler) ListBlobsInContainer(containerName string) ([]storage.Blob, error) {
	log.Debugf("ListBlobsInContainer %s", containerName)
	container := bh.blobStorageClient.GetContainerReference(containerName)
	seen := []storage.Blob{}
	marker := ""
	for {
		resp, err := container.ListBlobs(storage.ListBlobsParameters{
			MaxResults: 100,
			Marker:     marker})

		if err != nil {
			return nil, err
		}

		for _, v := range resp.Blobs {
			seen = append(seen, v)
		}

		marker = resp.NextMarker
		if marker == "" || len(resp.Blobs) == 0 {
			break
		}
	}

	return seen, nil
}

// ListContainers lists the blobs in a container
func (bh BlobHandler) ListContainers() ([]storage.Container, error) {
	log.Debugf("ListContainers start")

	seen := []storage.Container{}
	marker := ""
	for {
		containerResponse, err := bh.blobStorageClient.ListContainers(storage.ListContainersParameters{MaxResults: 100, Marker: marker})
		if err != nil {
			return nil, err
		}

		for _, v := range containerResponse.Containers {
			seen = append(seen, v)
		}

		marker = containerResponse.NextMarker
		if marker == "" || len(containerResponse.Containers) == 0 {
			break
		}
	}

	return seen, nil
}
