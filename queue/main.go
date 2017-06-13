package main

import (
	"azurestoragetools/blob/Handler"
	"azurestoragetools/common"
	"flag"
	"fmt"
	"os"
	"strconv"

	log "github.com/Sirupsen/logrus"
)

var Version string

// getCommand. Naive way to determine what the actual user wants to do. Copy, list etc etc.
// rework when it gets more complex.
func getCommand(push bool, pop bool, peek bool, size bool, createQueueCommand bool, generateQueueSASCommand bool) int {

	if !push && !pop && !peek && !size && !createQueueCommand && !generateQueueSASCommand {
		fmt.Println("No command given")
		os.Exit(1)
	}

	if push {
		return common.CommandPushQueue
	}

	if pop {
		return common.CommandPopQueue
	}

	if peek {
		return common.CommandPeekQueue
	}

	if size {
		return common.CommandSizeQueue
	}

	if createQueueCommand {
		return common.CommandCreateQueue
	}

	if generateQueueSASCommand {
		return common.CommandGernateQueueSAS
	}

	log.Fatal("unsure of command to use")
	return common.CommandUnknown
}

func setupConfiguration() *common.CloudConfig {
	config := common.NewCloudConfig()

	var version = flag.Bool("version", false, "Display Version")
	var debug = flag.Bool("debug", false, "Debug output")
	var push = flag.Bool("push", false, "Push message to queue")
	var msg = flag.String("message", "", "Message to push")
	var pop = flag.Bool("pop", false, "Pop message from queue")
	var peek = flag.Bool("peek", false, "Peek message at from of queue")
	var size = flag.Bool("size", false, "Get size of queue")
	var createQueueCommand = flag.Bool("createqueue", false, "Create queue for Azure")
	var generateQueueSASCommand = flag.Bool("queuesas", false, "Generate Queue SAS URL")

	var queueName = flag.String("queue", "", "Queue used for command")
	var timeout = flag.String("sastimeout", "60", "Optional: Timeout in seconds for generating SAS URL. Defaults to 60 seconds.")
	var perms = flag.String("sasperms", "r", "Optional: SAS permissions. Combination of rw")

	var azureDefaultAccountName = flag.String("AzureDefaultAccountName", "", "Default Azure Account Name")
	var azureDefaultAccountKey = flag.String("AzureDefaultAccountKey", "", "Default Azure Account Key")
	flag.Parse()

	config.Version = *version
	config.Debug = *debug
	if !*version {

		config.Command = getCommand(*push, *pop, *peek, *size, *createQueueCommand, *generateQueueSASCommand)
		config.Configuration[common.Queue] = *queueName
		config.Configuration[common.QueueMessage] = *msg
		config.Configuration[common.Timeout] = *timeout
		config.Configuration[common.SASPermissions] = *perms

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

	case common.CommandSASURLBlob:
		timeout, _ := strconv.Atoi(config.Configuration[common.Timeout])
		url, err := bh.GenerateSASURLForBlob(config.Configuration[common.Container], config.Configuration[common.BlobPrefix], timeout, config.Configuration[common.SASPermissions])
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("SAS URL %s", url)
		break

	case common.CommandSASURLContainer:
		timeout, _ := strconv.Atoi(config.Configuration[common.Timeout])
		url, err := bh.GenerateSASURLForContainer(config.Configuration[common.Container], timeout)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("SAS URL %s", url)
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
