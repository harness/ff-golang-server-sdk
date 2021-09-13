package ffgosdk

import (
	"encoding/json"
	"fmt"

	log "github.com/sirupsen/logrus"

	harness "github.com/drone/ff-golang-server-sdk/client"
	"github.com/drone/ff-golang-server-sdk/evaluation"
	"github.com/drone/ff-golang-server-sdk/test_wrapper/restapi"
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
func (s SDK) GetVariant(flag *restapi.FlagCheckBody) (string, error) {
	out := "failed to get flag"
	var err error
	if flag != nil {
		var t *evaluation.Target = nil
		if flag.Target != nil {
			t = &evaluation.Target{}
			if flag.Target.TargetIdentifier != nil {
				t.Identifier = *flag.Target.TargetIdentifier
			}
			if flag.Target.TargetName != nil {
				t.Name = *flag.Target.TargetName
			}
			attributes := map[string]interface{}{}
			if flag.Target.Attributes != nil {
				if flag.Target.Attributes.Email != nil {
					attributes["email"] = *flag.Target.Attributes.Email
				}
				if flag.Target.Attributes.Username != nil {
					attributes["username"] = *flag.Target.Attributes.Username
				}
				if flag.Target.Attributes.Region != nil {
					attributes["region"] = *flag.Target.Attributes.Region
				}
			}
			t.Attributes = &attributes
		}

		switch flag.FlagKind {
		case "boolean", "bool":
			log.Infof("get boolean variation: %s with target: %+v", flag.FlagKey, t)
			var variation bool
			variation, err = s.sdkClient.BoolVariation(flag.FlagKey, t, false)
			log.Info(variation)
			out = fmt.Sprintf("%t", variation)
		case "string":
			log.Infof("get string variation: %s with target: %+v", flag.FlagKey, t)
			var variation string
			variation, err = s.sdkClient.StringVariation(flag.FlagKey, t, "")
			log.Info(variation)
			out = fmt.Sprintf("%s", variation)
		case "int":
			log.Infof("get int variation: %s with target: %+v", flag.FlagKey, t)
			var variation int64
			variation, err = s.sdkClient.IntVariation(flag.FlagKey, t, -1)
			log.Info(variation)
			out = fmt.Sprintf("%d", variation)
		case "number":
			log.Infof("get number variation: %s with target: %+v", flag.FlagKey, t)
			var variation float64
			variation, err = s.sdkClient.NumberVariation(flag.FlagKey, t, -1)
			log.Info(variation)
			out = fmt.Sprintf("%g", variation)
		case "json":
			log.Infof("get json variation: %s with target: %+v", flag.FlagKey, t)
			var variation types.JSON
			variation, err = s.sdkClient.JSONVariation(flag.FlagKey, t, types.JSON{})
			log.Infof("%+v", variation)
			data, err := json.Marshal(variation)
			if err != nil {
				out = fmt.Sprintf("{}")
			} else {
				out = fmt.Sprintf("%s", string(data))
			}
		}
	}

	if err != nil {
		log.Error(err)
		return out, err
	}
	return out, nil
}
