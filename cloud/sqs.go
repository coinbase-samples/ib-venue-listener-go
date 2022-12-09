package cloud

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/coinbase-samples/ib-venue-listener-go/config"
	log "github.com/sirupsen/logrus"
)

var (
	sqsClient     *sqs.Client
	sqsClientLock sync.Mutex
)

func SqsSendMessage(
	ctx context.Context,
	app config.AppConfig,
	queueUrl,
	messageGroupId,
	data string,
) error {

	client, err := SqsClient(app)
	if err != nil {
		return err
	}

	log.WithFields(log.Fields{"url": queueUrl, "data": data}).Debug("Send SQS message")

	dedupId := md5.Sum([]byte(fmt.Sprintf("%s%s", queueUrl, data)))

	_, err = client.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:               aws.String(queueUrl),
		MessageBody:            aws.String(data),
		MessageDeduplicationId: aws.String(hex.EncodeToString(dedupId[:])),
		MessageGroupId:         aws.String(messageGroupId),
	})

	return err
}

func SqsClient(app config.AppConfig) (*sqs.Client, error) {
	sqsClientLock.Lock()
	defer sqsClientLock.Unlock()

	if sqsClient != nil {
		return sqsClient, nil
	}

	cfg, err := awsconfig.LoadDefaultConfig(context.TODO(), awsconfig.WithRegion(app.AwsRegion))
	if err != nil {
		return nil, fmt.Errorf("Unable to load AWS SDK config: %v", err)
	}

	if app.IsLocalEnv() {
		conn := fmt.Sprintf("http://%s:4566", app.LocalStackHostname)
		sqsClient = sqs.NewFromConfig(cfg, func(o *sqs.Options) {
			o.EndpointResolver = sqs.EndpointResolverFromURL(conn)
		})
	} else {
		sqsClient = sqs.NewFromConfig(cfg)
	}
	return sqsClient, nil

}
