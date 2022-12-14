package queue

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/coinbase-samples/ib-venue-listener-go/config"
	log "github.com/sirupsen/logrus"
)

var Repo *Repository

type Repository struct {
	Svc     *sqs.Client
	SvcLock sync.Mutex
}

func NewRepo(a *config.AppConfig) *Repository {
	svc := setupService(a)
	return &Repository{
		Svc: svc,
	}
}

func NewQueue(r *Repository) {
	Repo = r
}

func SqsSendMessage(
	ctx context.Context,
	app config.AppConfig,
	queueUrl,
	messageGroupId,
	data string,
) error {
	Repo.SvcLock.Lock()
	defer Repo.SvcLock.Unlock()
	log.WithFields(log.Fields{"url": queueUrl, "data": data}).Debug("Send SQS message")

	dedupId := md5.Sum([]byte(fmt.Sprintf("%s%s", queueUrl, data)))

	_, err := Repo.Svc.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:               aws.String(queueUrl),
		MessageBody:            aws.String(data),
		MessageDeduplicationId: aws.String(hex.EncodeToString(dedupId[:])),
		MessageGroupId:         aws.String(messageGroupId),
	})

	return err
}

func setupService(a *config.AppConfig) *sqs.Client {
	return sqs.NewFromConfig(a.AwsConfig)
}
