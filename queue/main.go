package main

import (
	"azurestoragetools/common"
	"azurestoragetools/queue/Handler"
	"flag"
	"fmt"
	"os"
	"strconv"

	log "github.com/Sirupsen/logrus"
)

var Version string

// getCommand. Naive way to determine what the actual user wants to do. Copy, list etc etc.
// rework when it gets more complex.
func getCommand(push bool, pop bool, peek bool, size bool, createQueueCommand bool, generateQueueSASCommand bool, clearQueueCommand bool) int {

	if !push && !pop && !peek && !size && !createQueueCommand && !generateQueueSASCommand && !clearQueueCommand {
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

	if clearQueueCommand {
		return common.CommandClearQueue
	}

	if generateQueueSASCommand {
		return common.CommandGenerateQueueSAS
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
	var size = flag.Bool("size", false, "Get approximate size of queue")
	var clear = flag.Bool("clear", false, "Clear queue")
	var createQueueCommand = flag.Bool("createqueue", false, "Create queue for Azure")
	var generateQueueSASCommand = flag.Bool("queuesas", false, "Generate Queue SAS URL")

	var queueName = flag.String("queue", "", "Queue used for command")
	var sastimeout = flag.String("sastimeout", "60", "Optional: Timeout in seconds for generating SAS URL. Defaults to 60 seconds")
	var visibilityTimeout = flag.String("vtimeout", "0", "Optional: visibility time for queue messsage")
	var ttl = flag.String("ttl", "0", "Optional: Time to live for queue messsage.")
	var perms = flag.String("sasperms", "r", "Optional: SAS permissions. Combination of rw")

	var azureDefaultAccountName = flag.String("AzureDefaultAccountName", "", "Default Azure Account Name")
	var azureDefaultAccountKey = flag.String("AzureDefaultAccountKey", "", "Default Azure Account Key")
	flag.Parse()

	config.Version = *version
	config.Debug = *debug
	if !*version {

		config.Command = getCommand(*push, *pop, *peek, *size, *createQueueCommand, *generateQueueSASCommand, *clear)
		config.Configuration[common.Queue] = *queueName
		config.Configuration[common.QueueMessage] = *msg
		config.Configuration[common.VisibilityTimeout] = *visibilityTimeout
		config.Configuration[common.TTL] = *ttl
		config.Configuration[common.SASTimeout] = *sastimeout
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

		validateConfig(config)
	}

	return config
}

// validateConfig checks if required params are met.
// if not, we mark the config as invalid and the program will explain and stop.
// Probably should do it in main where we have other switch statement, but will keep it here for now.
func validateConfig(config *common.CloudConfig) {

	if config.Configuration[common.Queue] == "" {
		fmt.Printf("Missing queue name\n")
		config.ValidConfig = false
	}

	switch config.Command {

	case common.CommandPushQueue:
		// make sure message exists!
		if config.Configuration[common.QueueMessage] == "" {
			fmt.Printf("Error: Require message and queue when pushing!")
			config.ValidConfig = false
		}
		break

	}

	config.ValidConfig = true
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

	qh, err := Handler.NewQueueHandler(config.Configuration[common.AzureDefaultAccountName], config.Configuration[common.AzureDefaultAccountKey])
	if err != nil {
		log.Debugf("Unable to create QueueHandler %s", err)
		return
	}

	switch config.Command {

	case common.CommandSizeQueue:
		sz, err := qh.QueueSize(config.Configuration[common.Queue])
		if err != nil {
			fmt.Printf("ERROR: %s", err)
		}

		fmt.Printf("%d", sz)
		break

	case common.CommandCreateQueue:
		err := qh.CreateQueue(config.Configuration[common.Queue])
		if err != nil {
			fmt.Printf("ERROR: %s", err)
		}
		break

	case common.CommandPushQueue:
		ttl, err := strconv.Atoi(config.Configuration[common.TTL])
		if err != nil {
			log.Fatal(err)
		}

		visibilityTimeout, err := strconv.Atoi(config.Configuration[common.VisibilityTimeout])
		if err != nil {
			log.Fatal(err)
		}

		if ttl > 0 && visibilityTimeout > 0 {
			err = qh.PushQueueWithTimeouts(config.Configuration[common.Queue], config.Configuration[common.QueueMessage], ttl, visibilityTimeout)
		} else {
			err = qh.PushQueue(config.Configuration[common.Queue], config.Configuration[common.QueueMessage])
		}

		if err != nil {
			log.Fatal(err)
		}
		break

	case common.CommandPopQueue:
		msg, err := qh.PopQueue(config.Configuration[common.Queue])
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("%s", msg)
		break

	case common.CommandClearQueue:
		err := qh.ClearQueue(config.Configuration[common.Queue])
		if err != nil {
			log.Fatal(err)
		}

		break
	case common.CommandUnknown:
		log.Fatal("Unsure of command to execute")
	}

}
