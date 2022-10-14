package dba

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/coinbase-samples/ib-venue-listener-go/model"
)

func (m *Repository) PutAsset(ctx context.Context, asset *model.Asset) error {

	item, err := attributevalue.MarshalMap(asset)
	if err != nil {
		return fmt.Errorf("unable to marshal asset: %v", err)
	}

	if _, err = m.Svc.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(m.App.AssetTableName),
		Item:      item,
	}); err != nil {
		return fmt.Errorf("unable to PutItem on DynamoDB - msg: %v", err)
	}

	return nil

}

func (m *Repository) LoadAssets(ctx context.Context) ([]model.Asset, error) {
	var assets []model.Asset

	out, err := m.Svc.Scan(ctx, &dynamodb.ScanInput{
		TableName: aws.String(m.App.AssetTableName),
		Limit:     aws.Int32(200), // TODO: Move away from scan and/or walk table
	})

	if err != nil {
		return assets, fmt.Errorf("unable to scan/load assets: %v", err)
	}

	err = attributevalue.UnmarshalListOfMaps(out.Items, &assets)
	if err != nil {
		return assets, fmt.Errorf("unable to unmarshal assets: %v", err)
	}

	return assets, nil
}
