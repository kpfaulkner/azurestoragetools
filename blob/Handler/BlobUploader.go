package Handler

import (
	"azure-sdk-for-go/storage"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"

	log "github.com/Sirupsen/logrus"
	"github.com/satori/uuid"
)

// UploadFiles uploads file based off filepath.
// Can either be single file (ie filePath doesn't end with a / or \)
// or it can be a directory so, filePath ends with / or \
func (bh BlobHandler) UploadFiles(filePath string, containerName string) error {
	log.Debugf("UploadFiles %s %s", filePath, containerName)
	container := bh.blobStorageClient.GetContainerReference(containerName)
	doesExist, err := container.Exists()
	if err != nil {
		return err
	}

	if !doesExist {
		return fmt.Errorf("Container %s doesn't exist", containerName)
	}

	// channel to hold names of all local files to copy.
	filesChannel := make(chan string, 1000)

	bh.launchUploadGoRoutines(containerName, filePath, filesChannel)

	allFiles := bh.getLocalFiles(filePath)
	fmt.Printf("Copying %d files\n", len(allFiles))

	for _, file := range allFiles {
		filesChannel <- file
	}

	close(filesChannel)
	wg.Wait()
	return nil
}

// launchUploadGoRoutines starts a number of Go Routines used for uploading
func (bh BlobHandler) launchUploadGoRoutines(containerName string, localFilePrefix string, copyChannel chan string) {

	log.Debugf("launching %d goroutines", bh.concurrentFactor)
	for i := 0; i < int(bh.concurrentFactor); i++ {
		wg.Add(1)
		go bh.uploadFileFromChannel(containerName, localFilePrefix, copyChannel)
	}
}

// uploadFileFromChannel reads blob from channel and uploads to Azure.
func (bh BlobHandler) uploadFileFromChannel(containerName string, localFilePrefix string, copyChannel chan string) {

	defer wg.Done()

	for {
		fileName, ok := <-copyChannel
		if !ok {
			log.Debugf("Closing channel for uploadFileFromChannel")
			// closed...   so all writing is done?  Or what?
			return
		}
		fmt.Printf("uploading %s\n", fileName)

		// calculate blob name
		blobName := generateBlobName(fileName, localFilePrefix)
		//blobName = "foo/kdkd"

		// get container. Should just be able to get once!
		container := bh.blobStorageClient.GetContainerReference(containerName)

		// create blob
		blob := container.GetBlobReference(blobName)

		// upload file.
		uploadFile(fileName, blob)
	}

}

func uploadFile(fileName string, blob *storage.Blob) error {

	log.Debugf("uploadFile %s", fileName)
	// get stream to file.
	file, err := os.OpenFile(fileName, os.O_RDONLY, 0)
	if err != nil {
		log.Errorf("Error %s", err)
		return err
	}
	defer file.Close()

	buffer := make([]byte, 1024*100)
	numBytesRead := 0
	blockIDList := []string{}
	finishedProcessing := false
	for finishedProcessing == false {
		numBytesRead, err = file.Read(buffer)
		if err != nil {
			finishedProcessing = true
			continue
		}

		if numBytesRead <= 0 {
			finishedProcessing = true
			continue
		}
		blockID, err := writeMemoryToBlob(blob, buffer[:numBytesRead])
		if err != nil {
			log.Fatal("Unable to write memory to blob ", err)
		}

		blockIDList = append(blockIDList, blockID)
	}

	blockSlice := generateBlockSlice(blockIDList)

	log.Debugf("blockslice is %v", blockSlice)
	if err := blob.PutBlockList(blockSlice, nil); err != nil {
		log.Fatalf("putBlockIDList failed %s", err)
	}

	return nil
}

func generateBlockSlice(blockIDList []string) []storage.Block {
	blockSlice := []storage.Block{}
	for _, block := range blockIDList {
		b := storage.Block{}
		b.ID = block
		b.Status = storage.BlockStatusLatest
		blockSlice = append(blockSlice, b)
	}
	return blockSlice
}

func writeMemoryToBlob(blob *storage.Blob, buffer []byte) (string, error) {

	log.Debugf("writeMemoryToBlob buffer length %d", len(buffer))
	blockID := fmt.Sprintf("%s", uuid.NewV4())

	log.Debugf("1generate blockID is %s", blockID)
	blockID = base64.StdEncoding.EncodeToString([]byte(blockID))

	log.Debugf("2generate blockID is %s", blockID)
	err := blob.PutBlock(blockID, buffer, nil)
	if err != nil {
		log.Fatalf("Unable to PutBlock %s, %s ", blockID, err)
	}
	return blockID, nil
}

// generateBlobName  if localFilePrefix is a file, then just return the file portion of pathName
// if localFilePrefix is a directory, then return the portion of pathName that doesn't include localFilePrefix
func generateBlobName(pathName string, localFilePrefix string) string {

	file, err := os.OpenFile(localFilePrefix, os.O_RDONLY, 0)
	if err != nil {
		log.Fatalf("Error %s", err)
	}
	defer file.Close()

	fi, err := file.Stat()
	if err != nil {
		log.Fatalf("Error %s", err)
	}

	// if directory, prune the sucker.
	if !fi.IsDir() {
		log.Debugf("name is %s", fi.Name())
		return fi.Name()
	}

	v := pathName[len(localFilePrefix):]
	s := filepath.ToSlash(v)
	log.Debugf("pruned generated blob name is %s", s)
	return s

}

// getLocalFiles gets the local files and puts the names on the filesChannel.
func (bh BlobHandler) getLocalFiles(filePath string) []string {

	fileSlice := []string{}

	err := filepath.Walk(filePath, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}

		log.Debugf("file %s", path)
		fileSlice = append(fileSlice, path)
		//filesChannel <- path
		return nil
	})

	if err != nil {
		log.Debugf("ERROR during walk %s", err)
	}

	return fileSlice
}
