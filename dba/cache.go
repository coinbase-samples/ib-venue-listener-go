package dba

import (
	"context"
	"sync"
	"time"

	"github.com/coinbase-samples/ib-venue-listener-go/model"
	log "github.com/sirupsen/logrus"

	"github.com/pkg/errors"
)

// How long before we requery to update
var defaultExpiry = 30 * time.Second

// SetDefaultExpiry update the default expiry for all cached parameters
//
// Note this will update expires value on the next refresh of entries.
func SetDefaultExpiry(expires time.Duration) {
	defaultExpiry = expires
}

// Entry an SSM entry in the cache
type Entry struct {
	value   []model.Asset
	expires time.Time
}

// Cache SSM cache which provides read access to parameters
type Cache interface {
	GetAssets(ctx context.Context) ([]model.Asset, error)
}

type cache struct {
	ssm       sync.Mutex
	ssmValues map[string]*Entry
}

// New new SSM cache
func NewCache() Cache {
	return &cache{
		ssmValues: make(map[string]*Entry),
	}
}

// GetKey retrieve a parameter from SSM and cache it.
func (ssc *cache) GetAssets(ctx context.Context) ([]model.Asset, error) {

	ssc.ssm.Lock()
	defer ssc.ssm.Unlock()

	ent, ok := ssc.ssmValues["assets"]
	if !ok {
		// record is missing
		return ssc.updateAssets(ctx)
	}

	if time.Now().After(ent.expires) {
		// we have expired and need to refresh
		log.Debugln("expired cache refreshing value")

		return ssc.updateAssets(ctx)
	}

	// return the value
	return ent.value, nil
}

func (ssc *cache) updateAssets(ctx context.Context) ([]model.Asset, error) {
	log.Debugln("updating assets from dynamo")

	assets, err := Repo.LoadAssets(ctx)
	if err != nil {
		return []model.Asset{}, errors.Wrapf(err, "failed to retrieve assets from dynamodb")
	}

	ssc.ssmValues["assets"] = &Entry{
		value:   assets,
		expires: time.Now().Add(defaultExpiry), // reset the expiry
	}

	log.Debugln("assets values refreshed from ssm at:", time.Now())

	return assets, nil
}
