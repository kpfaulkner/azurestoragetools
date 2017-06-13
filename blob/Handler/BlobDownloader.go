package Handler

import (
	"azure-sdk-for-go/storage"
	"fmt"
	"io"
	"os"
	"path/filepath"

	log "github.com/Sirupsen/logrus"
)

// DownloadFiles downloads blob to local filesystem (filePath)
// blobPrefix might be a specific blob or just literally a prefix.
// Firing off goroutine for eachdownload.
// Will see how this goes, otherwise we'll just throttle to a certain limit?
func (bh BlobHandler) DownloadFiles(containerName string, blobPrefix string, filePath string) error {
	container := bh.blobStorageClient.GetContainerReference(containerName)

	blobList := bh.listBlobs(containerName, blobPrefix)

	fmt.Printf("Downloading %d blobs\n", len(blobList))

	for _, blobName := range blobList {
		blob := container.GetBlobReference(blobName)
		sr, err := blob.Get(nil)
		if err != nil {
			log.Fatal(err)
			return err
		}
		defer sr.Close()

		localName := generateLocalName(filePath, blobName)

		// read it!
		fmt.Printf("reading %s\n", blobName)
		go bh.downloadFile(sr, localName)
	}

	return nil
}

func (bh *BlobHandler) downloadFile(sr io.ReadCloser, filePath string) error {

	dirPart := filepath.Dir(filePath)
	os.MkdirAll(dirPart, 0700)

	log.Debugf("downloading to file %s", filePath)
	file, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		log.Fatalf("create file error %s", err)
		return err
	}
	defer file.Close()

	// 100k buffer... way too small?
	buffer := make([]byte, 1024*100)

	finishedProcessing := false
	for finishedProcessing == false {
		numBytesRead, err := sr.Read(buffer)
		if err != nil {
			finishedProcessing = true
		}

		if numBytesRead <= 0 {
			finishedProcessing = true
			continue
		}

		_, err = file.Write(buffer[:numBytesRead])
		if err != nil {
			log.Fatal(err)
			return err
		}

	}
	return nil
}

// generateLocalFileName generates the complete path that a blob will be downloaded to.
// filePath is basically the location that it will be downloaded... blobName needs to be appended.
// But... blobName will need to be converted to the appropriate format for the OS.
func generateLocalName(filePath string, blobName string) string {
	return fmt.Sprintf("%s%s", filePath, filepath.FromSlash(blobName))
}

func (bh BlobHandler) listBlobs(containerName string, blobPrefix string) []string {

	// get all the blobs in the container with the given prefix!
	params := storage.ListBlobsParameters{Prefix: blobPrefix}
	container := bh.blobStorageClient.GetContainerReference(containerName)
	blobListResponse, err := container.ListBlobs(params)
	if err != nil {
		log.Fatal("Error")
	}

	blobList := []string{}

	for _, blob := range blobListResponse.Blobs {
		blobList = append(blobList, blob.Name)
	}

	return blobList
}
