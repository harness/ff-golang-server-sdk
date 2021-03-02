package cache

import (
	"github.com/mitchellh/mapstructure"
	"github.com/wings-software/ff-client-sdk-go/dto"
	"github.com/wings-software/ff-client-sdk-go/evaluation"
	"github.com/wings-software/ff-client-sdk-go/logger"
	"github.com/wings-software/ff-client-sdk-go/storage"
)

type Persistence struct {
	store  storage.Storage
	cache  Cache
	logger logger.Logger
}

func NewPersistence(store storage.Storage, cache Cache, logger logger.Logger) Persistence {
	return Persistence{store: store, cache: cache, logger: logger}
}

func (p Persistence) SaveToStore() error {
	if p.cache.Updated().Before(p.store.PersistedAt()) {
		return nil
	}
	p.logger.Info("Persisting cache data to the store")
	keys := p.cache.Keys()
	temp := make(map[string]interface{})
	for _, key := range keys {
		keyObject := key.(dto.Key)
		val, ok := p.cache.Get(key)
		if ok {

			if _, ok := temp[keyObject.Type]; !ok {
				temp[keyObject.Type] = make(map[string]interface{})
			}
			nameValue := temp[keyObject.Type].(map[string]interface{})
			nameValue[keyObject.Name] = val
		}
	}

	for key, val := range temp {
		p.store.Set(key, val)
	}
	err := p.store.Persist()
	if err != nil {
		return err
	}
	return nil
}

func (p *Persistence) LoadFromStore() error {
	p.logger.Info("Loading cache data from store")
	err := p.store.Load()
	if err != nil {
		return err
	}

	flags, ok := p.store.Get(dto.KeyFeature)
	if ok {
		for key, value := range flags.(map[string]interface{}) {
			keyData := dto.Key{
				Type: dto.KeyFeature,
				Name: key,
			}
			flag := evaluation.FeatureConfig{}
			err := mapstructure.Decode(value, &flag)
			if err != nil {
				return err
			}
			p.cache.Set(keyData, flag)
		}
	}

	segments, ok := p.store.Get(dto.KeySegment)
	if ok {
		for key, value := range segments.(map[string]interface{}) {
			keyData := dto.Key{
				Type: dto.KeySegment,
				Name: key,
			}
			segment := evaluation.Segment{}
			err := mapstructure.Decode(value, &segment)
			if err != nil {
				return err
			}
			p.cache.Set(keyData, value)
		}
	}
	return nil
}
