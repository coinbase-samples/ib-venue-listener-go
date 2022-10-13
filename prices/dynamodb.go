package prices

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

var ddc *dynamodb.Client

func DynamoDbClient() *dynamodb.Client {

	if ddc != nil {
		return ddc
	}

	cfg, err := awsConfig.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatalf("Unable to load AWS config", err)
	}

	ddc = dynamodb.NewFromConfig(cfg)

	return ddc
}

func PutAsset(ctx context.Context, asset *Asset) error {

	item, err := attributevalue.MarshalMap(asset)
	if err != nil {
		return fmt.Errorf("Unable to marshal asset: %v", err)
	}

	if _, err = DynamoDbClient().PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(os.Getenv("PRODUCT_PRICE_TABLE_NAME")),
		Item:      item,
	}); err != nil {
		return fmt.Errorf("Unable to PutItem on DynamoDB - msg: %v", err)
	}

	return nil

}

func LoadAssets(ctx context.Context) ([]Asset, error) {
	var assets []Asset

	out, err := DynamoDbClient().Scan(ctx, &dynamodb.ScanInput{
		TableName: aws.String(os.Getenv("PRODUCT_PRICE_TABLE_NAME")),
		Limit:     aws.Int32(200), // TODO: Move away from scan and/or walk table
	})

	if err != nil {
		return assets, fmt.Errorf("Unable to scan/load assets: %v", err)
	}

	err = attributevalue.UnmarshalListOfMaps(out.Items, &assets)
	if err != nil {
		return assets, fmt.Errorf("Unable to unmarshal assets: %v", err)
	}

	return assets, nil
}
