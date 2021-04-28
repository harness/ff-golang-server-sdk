package wrapperconfig

// Config is used for the various env vars we need
type Config struct {
	WrapperHostname string
	WrapperPort     string
	SdkKey          string
	BaseURL         string
	EnableStreaming bool
}
