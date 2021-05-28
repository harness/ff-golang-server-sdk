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
	log.Warn("flag body   ---------------")
	log.Warnf("%+v", *flagBody)

	variantion, err := s.sdk.GetVariant(flagBody)
	if err != nil {
		log.Error(err)
		return ctx.JSON(http.StatusInternalServerError, restapi.ErrorResponse{ErrorMessage: err.Error()})
	}

	flagValue := restapi.FlagCheckResponse{FlagKey: flagBody.FlagKey, FlagValue: variantion}
	return ctx.JSON(http.StatusOK, flagValue)
}
