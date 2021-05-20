package handlers

import (
	"net/http"

	"github.com/drone/ff-golang-server-sdk/test_wrapper/ffgosdk"
	restapi "github.com/drone/ff-golang-server-sdk/test_wrapper/restapi"
	"github.com/drone/ff-golang-server-sdk/test_wrapper/wrapperconfig"
	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
)

// ServerImpl .
type ServerImpl struct {
	sdk ffgosdk.SDK
}

// NewServer .
func NewServer(config wrapperconfig.Config) ServerImpl {
	sdkClient, err := ffgosdk.NewSDK(config)
	if err != nil {
		panic(err)
	}
	return ServerImpl{
		sdk: sdkClient,
	}
}

// Ping .
func (s *ServerImpl) Ping(ctx echo.Context) error {
	return ctx.JSON(http.StatusOK, restapi.PongResponse{Pong: true})
}

// GetFlagValue .
func (s *ServerImpl) GetFlagValue(ctx echo.Context) error {
	defer log.Info("--------------")
	log.Info("GetFlagValue")
	flagBody := new(restapi.FlagCheckBody)
	err := ctx.Bind(flagBody)
	if err != nil {
		log.Error(err)
		return ctx.JSON(http.StatusInternalServerError, restapi.ErrorResponse{ErrorMessage: err.Error()})
	}
	log.Infof("%+v", *flagBody)
	var targetMap map[string]string
	if flagBody.Target != nil {
		if flagBody.Target.Name != nil {
			targetMap["name"] = *flagBody.Target.Name
		}
		if flagBody.Target.Email != nil {
			targetMap["email"] = *flagBody.Target.Email
		}
		if flagBody.Target.Region != nil {
			targetMap["region"] = *flagBody.Target.Region
		}
	}

	variantion, err := s.sdk.GetVariant(flagBody.FlagKind, flagBody.FlagKey, &targetMap)
	if err != nil {
		log.Error(err)
		return ctx.JSON(http.StatusInternalServerError, restapi.ErrorResponse{ErrorMessage: err.Error()})
	}

	flagValue := restapi.FlagCheckResponse{FlagKey: flagBody.FlagKey, FlagValue: variantion}
	return ctx.JSON(http.StatusOK, flagValue)
}
