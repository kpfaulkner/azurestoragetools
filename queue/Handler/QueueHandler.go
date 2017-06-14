package Handler

import (
	"azure-sdk-for-go/storage"
	"errors"
	"sync"

	log "github.com/Sirupsen/logrus"
)

type QueueHandler struct {

	// creds.
	accountName        string
	accountKey         string
	queueStorageClient storage.QueueServiceClient
}

var wg sync.WaitGroup

// NewQueueHandler   create new instance of QueueHandler
func NewQueueHandler(accountName string, accountKey string) (*QueueHandler, error) {
	qh := new(QueueHandler)

	client, err := storage.NewBasicClient(accountName, accountKey)

	if err != nil {
		return nil, err
	}

	qh.accountName = accountName
	qh.accountKey = accountKey
	qh.queueStorageClient = client.GetQueueService()
	return qh, nil
}

// GenerateSASURL generates SAS URL for blob
func (qh QueueHandler) GenerateSASURL(queueName string, durationInSeconds int, permissions string) (string, error) {

	return "", nil
}

// CreateQueue creates a new queue
func (qh QueueHandler) CreateQueue(queueName string) error {

	log.Debugf("CreateQueue %s", queueName)
	queue := qh.queueStorageClient.GetQueueReference(queueName)
	err := queue.Create(nil)
	if err != nil {
		return err
	}
	return nil
}

// PushQueue creates a new queue message
func (qh QueueHandler) pushQueue(queueName string, message string, options *storage.PutMessageOptions) error {
	log.Debugf("PushQueue %s: %s", queueName, message)

	queue := qh.queueStorageClient.GetQueueReference(queueName)
	doesExist, err := queue.Exists()
	if err != nil {
		return err
	}

	if !doesExist {
		return errors.New("Queue does not exist")
	}

	msg := queue.GetMessageReference(message)
	err = msg.Put(options)
	if err != nil {
		return err
	}

	return nil
}

// PushQueueWithTimeouts creates a new queue message
func (qh QueueHandler) PushQueueWithTimeouts(queueName string, message string, timeToLive int, visibilityTimeout int) error {
	log.Debugf("PushQueue %s: %s : %d %d", queueName, message, timeToLive, visibilityTimeout)

	options := storage.PutMessageOptions{
		VisibilityTimeout: visibilityTimeout,
		MessageTTL:        timeToLive,
	}
	return qh.pushQueue(queueName, message, &options)

}

// PushQueue creates a new queue message
func (qh QueueHandler) PushQueue(queueName string, message string) error {
	log.Debugf("PushQueue %s: %s", queueName, message)
	return qh.pushQueue(queueName, message, nil)

}

// PopQueue creates a new queue
func (qh QueueHandler) ClearQueue(queueName string) error {
	queue := qh.queueStorageClient.GetQueueReference(queueName)
	doesExist, err := queue.Exists()
	if err != nil {
		return err
	}

	if !doesExist {
		return errors.New("Queue does not exist")
	}

	err = queue.ClearMessages(nil)
	if err != nil {
		return err
	}

	return nil
}

// PopQueue creates a new queue
func (qh QueueHandler) PopQueue(queueName string) (string, error) {
	queue := qh.queueStorageClient.GetQueueReference(queueName)
	doesExist, err := queue.Exists()
	if err != nil {
		return "", err
	}

	if !doesExist {
		return "", errors.New("Queue does not exist")
	}

	msgList, err := queue.GetMessages(&storage.GetMessagesOptions{NumOfMessages: 1})
	if err != nil {
		return "", err
	}

	if len(msgList) > 0 {
		// make sure its marked as read!
		msgList[0].Delete(nil)

		// just really interested in the content.
		return msgList[0].Text, nil
	}

	return "", nil
}

// PeekQueue creates a new queue
func (qh QueueHandler) PeekQueue(queueName string) (string, error) {
	queue := qh.queueStorageClient.GetQueueReference(queueName)
	doesExist, err := queue.Exists()
	if err != nil {
		return "", err
	}

	if !doesExist {
		return "", errors.New("Queue does not exist")
	}

	msgList, err := queue.PeekMessages(&storage.PeekMessagesOptions{NumOfMessages: 1})
	if err != nil {
		return "", err
	}
	return msgList[0].Text, nil
}

// QueueSize returns size of queue
func (qh QueueHandler) QueueSize(queueName string) (uint64, error) {
	queue := qh.queueStorageClient.GetQueueReference(queueName)
	doesExist, err := queue.Exists()
	if err != nil {
		return 0, err
	}

	if !doesExist {
		return 0, errors.New("Queue does not exist")
	}

	err = queue.GetMetadata(nil)
	if err != nil {
		return 0, err
	}

	log.Debugf("queue %v", queue)
	return queue.AproxMessageCount, nil
}
