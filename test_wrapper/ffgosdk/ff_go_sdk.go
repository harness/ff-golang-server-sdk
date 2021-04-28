package ffgosdk

import (
	"encoding/json"
	"fmt"

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
		t.Attributes = make(map[string]interface{})
		for k, v := range *targetMap {
			t.Attributes[k] = v
		}
	}

	var err error
	switch kind {
	case "boolean":
		var variation bool
		variation, err = s.sdkClient.BoolVariation(flagKey, t, false)
		out = fmt.Sprintf("%t", variation)
	case "string":
		var variation string
		variation, err = s.sdkClient.StringVariation(flagKey, t, "")
		out = fmt.Sprintf("%s", variation)
	case "int":
		var variation int64
		variation, err = s.sdkClient.IntVariation(flagKey, t, -1)
		out = fmt.Sprintf("%d", variation)
	case "number":
		var variation float64
		variation, err = s.sdkClient.NumberVariation(flagKey, t, -1)
		out = fmt.Sprintf("%g", variation)
	case "json":
		var variation types.JSON
		variation, err = s.sdkClient.JSONVariation(flagKey, t, types.JSON{})
		data, err := json.Marshal(variation)
		if err != nil {
			out = fmt.Sprintf("{}")
		} else {
			out = fmt.Sprintf("%s", string(data))
		}
	}

	if err != nil {
		return out, err
	}
	return out, nil
}
