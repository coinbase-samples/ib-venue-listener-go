package cloud

import (
	"context"
	"fmt"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/kinesis"
	"github.com/coinbase-samples/ib-venue-listener-go/config"
)

var (
	kinesisClient     *kinesis.Client
	kinesisClientLock sync.Mutex
)

func KdsPutRecord(
	ctx context.Context,
	app config.AppConfig,
	streamName,
	partitionKey string,
	data []byte,
) error {

	client, err := KdsClient(app)
	if err != nil {
		return err
	}

	_, err = client.PutRecord(ctx, &kinesis.PutRecordInput{
		PartitionKey: aws.String(partitionKey),
		StreamName:   aws.String(streamName),
		Data:         data,
	})

	return err
}

func KdsClient(app config.AppConfig) (*kinesis.Client, error) {
	kinesisClientLock.Lock()
	defer kinesisClientLock.Unlock()

	if kinesisClient != nil {
		return kinesisClient, nil
	}

	cfg, err := awsconfig.LoadDefaultConfig(context.TODO(), awsconfig.WithRegion(app.AwsRegion))
	if err != nil {
		return nil, fmt.Errorf("Unable to load AWS SDK config: %v", err)
	}

	kinesisClient = kinesis.NewFromConfig(cfg)

	return kinesisClient, nil

}
