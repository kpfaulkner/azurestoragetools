package main

import (
	"azurestoragetools/blob/Handler"
	"azurestoragetools/common"
	"flag"
	"fmt"
	"os"

	log "github.com/Sirupsen/logrus"
)

var Version string

// getCommand. Naive way to determine what the actual user wants to do. Copy, list etc etc.
// rework when it gets more complex.
func getCommand(uploadCommand bool, downloadCommand bool, listCommand bool, createContainerCommand bool, listContainersCommand bool) int {

	if !uploadCommand && !downloadCommand && !listCommand && !createContainerCommand && !listContainersCommand {
		fmt.Println("No command given")
		os.Exit(1)
	}

	if listContainersCommand {
		return common.CommandListContainers
	}

	if uploadCommand {
		return common.CommandUpload
	}

	if downloadCommand {
		return common.CommandDownload
	}

	if listCommand {
		return common.CommandListBlobs
	}

	if createContainerCommand {
		return common.CommandCreateContainer
	}

	log.Fatal("unsure of command to use")
	return common.CommandUnknown
}

func setupConfiguration() *common.CloudConfig {
	config := common.NewCloudConfig()

	var concurrentCount = flag.Uint("cc", 5, "Concurrent Count. How many blobs are copied concurrently")

	var version = flag.Bool("version", false, "Display Version")
	var localFilesystem = flag.String("local", "", "Path for local filesystem")
	var debug = flag.Bool("debug", false, "Debug output")
	var upload = flag.Bool("upload", false, "Upload from local filesystem to Azure")
	var download = flag.Bool("download", false, "Download to local filesystem from Azure")
	var listCommand = flag.Bool("list", false, "List blobs in container")
	var listContainersCommand = flag.Bool("listcontainers", false, "List available containers")
	var createContainerCommand = flag.Bool("createcontainer", false, "Create container for Azure")
	var containerName = flag.String("container", "", "Container used for command")
	var blobPrefix = flag.String("blobprefix", "", "Optional: BlobPrefix for download command. This can either be entire blob name or just a prefix.")

	var azureDefaultAccountName = flag.String("AzureDefaultAccountName", "", "Default Azure Account Name")
	var azureDefaultAccountKey = flag.String("AzureDefaultAccountKey", "", "Default Azure Account Key")
	flag.Parse()

	config.Version = *version
	config.Debug = *debug
	if !*version {

		if *concurrentCount > 1000 {
			fmt.Printf("Maximum number for concurrent count is 1000")
			os.Exit(1)
		}

		config.Command = getCommand(*upload, *download, *listCommand, *createContainerCommand, *listContainersCommand)
		config.Configuration[common.Local] = *localFilesystem
		config.Configuration[common.Container] = *containerName
		config.Configuration[common.BlobPrefix] = *blobPrefix
		config.ConcurrentCount = *concurrentCount

		config.Configuration[common.AzureDefaultAccountName] = os.Getenv("ACCOUNT_NAME")
		config.Configuration[common.AzureDefaultAccountKey] = os.Getenv("ACCOUNT_KEY")

		// passed params trumps env vars.
		if *azureDefaultAccountName != "" {
			config.Configuration[common.AzureDefaultAccountName] = *azureDefaultAccountName
		}

		if *azureDefaultAccountKey != "" {
			config.Configuration[common.AzureDefaultAccountKey] = *azureDefaultAccountKey
		}

	}

	return config
}

// "so it begins"
func main() {

	config := setupConfiguration()

	if !config.Debug {
		log.SetLevel(log.InfoLevel)
	} else {
		log.SetLevel(log.DebugLevel)
	}
	log.Debug("after config setup")

	// if display version, then display then exit
	if config.Version {
		fmt.Println("Version: " + Version)
		return
	}

	bh, err := Handler.NewBlobHandler(config.Configuration[common.AzureDefaultAccountName], config.Configuration[common.AzureDefaultAccountKey], 5)
	if err != nil {
		log.Debugf("Unable to create BlobHandler")
		return
	}

	switch config.Command {
	case common.CommandUpload:
		err := bh.UploadFiles(config.Configuration[common.Local], config.Configuration[common.Container])
		if err != nil {
			log.Fatal(err)
		}
		break

	case common.CommandDownload:
		err := bh.DownloadFiles(config.Configuration[common.Container], config.Configuration[common.BlobPrefix], config.Configuration[common.Local])
		if err != nil {
			log.Fatal(err)
		}
		break

	case common.CommandListBlobs:
		blobList, err := bh.ListBlobsInContainer(config.Configuration[common.Container])
		if err != nil {
			log.Fatal(err)
		}

		for _, b := range blobList {
			fmt.Printf("%s\n", b.Name)
		}
		break

	case common.CommandListContainers:
		containerList, err := bh.ListContainers()
		if err != nil {
			log.Fatal(err)
		}

		for _, c := range containerList {
			fmt.Printf("%s\n", c.Name)
		}
		break

	case common.CommandCreateContainer:
		err := bh.CreateContainer(config.Configuration[common.Container])
		if err != nil {
			log.Fatal(err)
		}

	case common.CommandUnknown:
		log.Fatal("Unsure of command to execute")
	}

}
