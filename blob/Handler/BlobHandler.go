package Handler

import (
	"azure-sdk-for-go/storage"
	"sync"

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

// DownloadFile downloads blob to local filesystem
func (bh BlobHandler) DownloadFile(containerName string, blobName string, filePath string) error {
	return nil
}

// Delete a blob
func (bh BlobHandler) Delete(containerName string, blobName string) error {
	return nil
}

// GenerateSASURLForBlob generates SAS URL for blob
func (bh BlobHandler) GenerateSASURLForBlob(containerName string, blobName string, durationInMinutes int) (string, error) {
	return "", nil
}

// GenerateSASURLForContainer generates SAS URL for container
func (bh BlobHandler) GenerateSASURLForContainer(containerName string, durationInMinutes int) (string, error) {
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
