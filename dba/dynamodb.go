package dba

import (
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/coinbase-samples/ib-venue-listener-go/config"
)

// Repo the repository used by dynamo
var Repo *Repository

// Repository is the repository type
type Repository struct {
	App   *config.AppConfig
	Svc   *dynamodb.Client
	Cache Cache
}

// NewRepo creates a new repository
func NewRepo(a *config.AppConfig) *Repository {
	svc := setupService(a)
	cache := NewCache()
	return &Repository{
		App:   a,
		Svc:   svc,
		Cache: cache,
	}
}

// NewDBA sets the repository for the handlers
func NewDBA(r *Repository) {
	Repo = r
}

func setupService(a *config.AppConfig) *dynamodb.Client {
	return dynamodb.NewFromConfig(a.AwsConfig, func(o *dynamodb.Options) {
		o.EndpointResolver = dynamodb.EndpointResolverFromURL(a.DatabaseEndpoint)
	})
}
