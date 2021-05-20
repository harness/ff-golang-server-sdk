package ffgosdk

import (
	"encoding/json"
	"fmt"
	log "github.com/sirupsen/logrus"

	harness "github.com/drone/ff-golang-server-sdk/client"
	"github.com/drone/ff-golang-server-sdk/evaluation"
	"github.com/drone/ff-golang-server-sdk/test_wrapper/wrapperconfig"
	"github.com/drone/ff-golang-server-sdk/types"
)

// SDK .
type SDK struct {
	sdkClient *harness.CfClient
}

// NewSDK .
func NewSDK(config wrapperconfig.Config) (SDK, error) {
	client, err := harness.NewCfClient(config.SdkKey,
		harness.WithURL(config.BaseURL),
		harness.WithStreamEnabled(config.EnableStreaming),
	)

	if err != nil {
		return SDK{}, err
	}

	return SDK{
		sdkClient: client,
	}, nil
}

// Close .
func (s SDK) Close() error {
	return s.sdkClient.Close()
}

// GetVariant .
func (s SDK) GetVariant(kind string, flagKey string, targetMap *map[string]string) (string, error) {
	out := ""
	t := &evaluation.Target{}
	if targetMap != nil {
		att := make(map[string]interface{})
		for k, v := range *targetMap {
			att[k] = v
		}
		t.Attributes = &att
	}

	var err error
	switch kind {
	case "boolean":
		log.Infof("get boolean variation: %s", flagKey)
		var variation bool
		variation, err = s.sdkClient.BoolVariation(flagKey, t, false)
		log.Info(variation)
		out = fmt.Sprintf("%t", variation)
	case "string":
		log.Infof("get string variation: %s", flagKey)
		var variation string
		variation, err = s.sdkClient.StringVariation(flagKey, t, "")
		log.Info(variation)
		out = fmt.Sprintf("%s", variation)
	case "int":
		log.Infof("get int variation: %s", flagKey)
		var variation int64
		variation, err = s.sdkClient.IntVariation(flagKey, t, -1)
		log.Info(variation)
		out = fmt.Sprintf("%d", variation)
	case "number":
		log.Infof("get number variation: %s", flagKey)
		var variation float64
		variation, err = s.sdkClient.NumberVariation(flagKey, t, -1)
		log.Info(variation)
		out = fmt.Sprintf("%g", variation)
	case "json":
		log.Infof("get json variation: %s", flagKey)
		var variation types.JSON
		variation, err = s.sdkClient.JSONVariation(flagKey, t, types.JSON{})
		log.Infof("%+v",variation)
		data, err := json.Marshal(variation)
		if err != nil {
			out = fmt.Sprintf("{}")
		} else {
			out = fmt.Sprintf("%s", string(data))
		}
	}

	if err != nil {
		log.Error(err)
		return out, err
	}
	return out, nil
}
