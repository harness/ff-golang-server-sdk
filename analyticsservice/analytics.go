package analyticsservice

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/harness/ff-golang-server-sdk/sdk_codes"

	"github.com/harness/ff-golang-server-sdk/rest"

	"github.com/harness/ff-golang-server-sdk/evaluation"
	"github.com/harness/ff-golang-server-sdk/metricsclient"

	"github.com/harness/ff-golang-server-sdk/logger"
)

const (
	ffMetricType                 string = "FFMETRICS"
	featureIdentifierAttribute   string = "featureIdentifier"
	featureNameAttribute         string = "featureName"
	variationIdentifierAttribute string = "variationIdentifier"
	variationValueAttribute      string = "featureValue"
	targetAttribute              string = "target"
	sdkVersionAttribute          string = "SDK_VERSION"
	SdkVersion                   string = "0.1.26"
	sdkTypeAttribute             string = "SDK_TYPE"
	sdkType                      string = "server"
	sdkLanguageAttribute         string = "SDK_LANGUAGE"
	sdkLanguage                  string = "go"
	globalTarget                 string = "global"
	maxAnalyticsEntries          int    = 10000
	maxTargetEntries             int    = 100000
)

// SafeAnalyticsCache is a type that provides thread safe access to maps used by analytics
type SafeAnalyticsCache[K comparable, V any] interface {
	set(key K, value V)
	get(key K) (V, bool)
	delete(key K)
	size() int
	clear()
	iterate(func(K, V))
}

// SafeSeenTargetsCache extends SafeAnalyticsCache and adds behavior specific to seen targets
type SafeSeenTargetsCache[K comparable, V any] interface {
	SafeAnalyticsCache[K, V]
	isLimitExceeded() bool
}

type analyticsEvent struct {
	target        *evaluation.Target
	featureConfig *rest.FeatureConfig
	variation     *rest.Variation
	count         int
}

// AnalyticsService provides a way to cache and send analytics to the server
type AnalyticsService struct {
	analyticsChan               chan analyticsEvent
	evaluationAnalytics         SafeAnalyticsCache[string, analyticsEvent]
	targetAnalytics             SafeAnalyticsCache[string, evaluation.Target]
	seenTargets                 SafeSeenTargetsCache[string, bool]
	logEvaluationLimitReached   atomic.Bool
	logTargetLimitReached       atomic.Bool
	timeout                     time.Duration
	logger                      logger.Logger
	metricsClient               metricsclient.ClientWithResponsesInterface
	environmentID               string
	seenTargetsClearingInterval time.Duration
}

// NewAnalyticsService creates and starts a analytics service to send data to the client
func NewAnalyticsService(timeout time.Duration, logger logger.Logger, seenTargetsMaxSize int, seenTargetsClearingSchedule time.Duration) *AnalyticsService {
	serviceTimeout := timeout
	if timeout < 60*time.Second {
		serviceTimeout = 60 * time.Second
	} else if timeout > 1*time.Hour {
		serviceTimeout = 1 * time.Hour
	}
	as := AnalyticsService{
		analyticsChan:               make(chan analyticsEvent),
		evaluationAnalytics:         newSafeEvaluationAnalytics(),
		targetAnalytics:             newSafeTargetAnalytics(),
		seenTargets:                 newSafeSeenTargets(seenTargetsMaxSize),
		timeout:                     serviceTimeout,
		logger:                      logger,
		seenTargetsClearingInterval: seenTargetsClearingSchedule,
	}
	go as.listener()

	return &as
}

// Start starts the client and timer to send analytics
func (as *AnalyticsService) Start(ctx context.Context, client metricsclient.ClientWithResponsesInterface, environmentID string) {
	as.logger.Infof("%s Metrics started", sdk_codes.MetricsStarted)
	as.metricsClient = client
	as.environmentID = environmentID
	go as.startTimer(ctx)
	go as.startSeenTargetsClearingSchedule(ctx, as.seenTargetsClearingInterval)
}

func (as *AnalyticsService) startTimer(ctx context.Context) {
	for {
		select {
		case <-time.After(as.timeout):
			timeStamp := time.Now().UnixNano() / (int64(time.Millisecond) / int64(time.Nanosecond))
			as.sendDataAndResetCache(ctx, timeStamp)
		case <-ctx.Done():
			close(as.analyticsChan)
			as.logger.Infof("%s Metrics stopped", sdk_codes.MetricsStopped)
			return
		}
	}
}

// PushToQueue is used to queue analytics data to send to the server
func (as *AnalyticsService) PushToQueue(featureConfig *rest.FeatureConfig, target *evaluation.Target, variation *rest.Variation) {

	ad := analyticsEvent{
		target:        target,
		featureConfig: featureConfig,
		variation:     variation,
	}
	as.analyticsChan <- ad
}

func (as *AnalyticsService) listener() {
	as.logger.Info("Analytics cache successfully initialized")
	for ad := range as.analyticsChan {
		analyticsKey := getEvaluationAnalyticKey(ad)

		// Check if we've hit capacity for evaluations
		if as.evaluationAnalytics.size() < maxAnalyticsEntries {
			// Update evaluation metrics
			analytic, ok := as.evaluationAnalytics.get(analyticsKey)
			if !ok {
				ad.count = 1
				as.evaluationAnalytics.set(analyticsKey, ad)
			} else {
				ad.count = analytic.count + 1
				as.evaluationAnalytics.set(analyticsKey, ad)
			}
		} else {
			if !as.logEvaluationLimitReached.Load() {
				as.logger.Warnf("%s Evaluation analytic cache reached max size, remaining evaluation metrics for this analytics interval will not be sent", sdk_codes.EvaluationMetricsMaxSizeReached)
				as.logEvaluationLimitReached.Store(true)
			}
		}

		// Check if target is nil or anonymous
		if ad.target == nil || (ad.target.Anonymous != nil && *ad.target.Anonymous) {
			continue
		}

		// Check if target has been seen
		if _, seen := as.seenTargets.get(ad.target.Identifier); seen {
			continue
		}

		// Check if seen targets limit has been hit
		if as.seenTargets.isLimitExceeded() {
			continue
		}

		// Update seen targets
		as.seenTargets.set(ad.target.Identifier, true)

		// Update target metrics
		if as.targetAnalytics.size() < maxTargetEntries {
			as.targetAnalytics.set(ad.target.Identifier, *ad.target)
		} else {
			if !as.logTargetLimitReached.Load() {
				as.logger.Warnf("%s Target analytics cache reached max size, remaining target metrics for this analytics interval will not be sent", sdk_codes.TargetMetricsMaxSizeReached)
				as.logTargetLimitReached.Store(true)
			}
		}
	}
}

func convertInterfaceToString(i interface{}) string {
	if i == nil {
		return "nil"
	}

	switch v := i.(type) {
	case string:
		return v
	case int:
		return strconv.Itoa(v)
	case int64:
		return strconv.FormatInt(v, 10)
	case float64:
		return fmt.Sprintf("%v", v)
	case float32:
		return fmt.Sprintf("%v", v)
	case bool:
		val := "false"
		if v {
			val = "true"
		}
		return val
	default:
		// As a last resort
		return fmt.Sprintf("%v", v)
	}
}

func (as *AnalyticsService) sendDataAndResetCache(ctx context.Context, timeStamp int64) {

	// Clone and reset the evaluation analytics cache to minimise the duration
	// for which locks are held, so that metrics processing does not affect flag evaluations performance.
	// Although this might occasionally result in the loss of some metrics during periods of high load,
	// it is an acceptable tradeoff to prevent extended lock periods that could degrade user code.
	evaluationAnalyticsClone := as.evaluationAnalytics

	as.evaluationAnalytics = newSafeEvaluationAnalytics()

	// Clone and reset target analytics cache for same reason.
	targetAnalyticsClone := as.targetAnalytics
	as.targetAnalytics = newSafeTargetAnalytics()

	// Reset flags that keep track of cache limits being reached for logging purposes
	as.logEvaluationLimitReached.Store(false)
	as.logTargetLimitReached.Store(false)

	// Process evaluation metrics
	metricData := as.processEvaluationMetrics(evaluationAnalyticsClone, timeStamp)

	// Process target metrics
	targetData := as.processTargetMetrics(targetAnalyticsClone)

	analyticsPayload := metricsclient.PostMetricsJSONRequestBody{
		MetricsData: &metricData,
		TargetData:  &targetData,
	}

	if as.metricsClient != nil {
		emptyMetricsData := len(metricData) == 0
		emptyTargetData := len(targetData) == 0

		// if we have no metrics to send skip the post request
		if emptyMetricsData && emptyTargetData {
			as.logger.Debug("No metrics or target data to send")
			return
		}

		mClient := as.metricsClient

		jsonData, err := json.Marshal(analyticsPayload)
		if err != nil {
			as.logger.Errorf(err.Error())
		}
		as.logger.Debug(string(jsonData))

		resp, err := mClient.PostMetricsWithResponse(ctx, metricsclient.EnvironmentPathParam(as.environmentID), nil, analyticsPayload)
		if err != nil {
			as.logger.Warn(err)
			return
		}
		if resp == nil {
			as.logger.Warn("Empty response from metrics server")
			return
		}
		if resp.StatusCode() != 200 {
			as.logger.Warnf("%s Non 200 response from metrics server: %d", sdk_codes.MetricsSendFail, resp.StatusCode())
			return
		}

		as.logger.Debugf("%s Metrics sent to server", sdk_codes.MetricsSendSuccess)
	} else {
		as.logger.Warn("metrics client is not set")
	}
}

func (as *AnalyticsService) processEvaluationMetrics(evaluationAnalytics SafeAnalyticsCache[string, analyticsEvent], timeStamp int64) []metricsclient.MetricsData {
	metricData := make([]metricsclient.MetricsData, 0, evaluationAnalytics.size())
	evaluationAnalytics.iterate(func(key string, analytic analyticsEvent) {
		metricAttributes := []metricsclient.KeyValue{
			{Key: featureIdentifierAttribute, Value: analytic.featureConfig.Feature},
			{Key: featureNameAttribute, Value: analytic.featureConfig.Feature},
			{Key: variationIdentifierAttribute, Value: analytic.variation.Identifier},
			{Key: variationValueAttribute, Value: analytic.variation.Value},
			{Key: sdkTypeAttribute, Value: sdkType},
			{Key: sdkLanguageAttribute, Value: sdkLanguage},
			{Key: sdkVersionAttribute, Value: SdkVersion},
			{Key: targetAttribute, Value: globalTarget},
		}

		md := metricsclient.MetricsData{
			Timestamp:   timeStamp,
			Count:       analytic.count,
			MetricsType: metricsclient.MetricsDataMetricsType(ffMetricType),
			Attributes:  metricAttributes,
		}
		metricData = append(metricData, md)
	})

	return metricData
}

func (as *AnalyticsService) processTargetMetrics(targetAnalytics SafeAnalyticsCache[string, evaluation.Target]) []metricsclient.TargetData {
	targetData := make([]metricsclient.TargetData, 0, targetAnalytics.size())

	targetAnalytics.iterate(func(key string, target evaluation.Target) {
		targetAttributes := make([]metricsclient.KeyValue, 0)
		if target.Attributes != nil {
			targetAttributes = make([]metricsclient.KeyValue, 0, len(*target.Attributes))
			for k, v := range *target.Attributes {
				targetAttributes = append(targetAttributes, metricsclient.KeyValue{Key: k, Value: convertInterfaceToString(v)})
			}
		}

		td := metricsclient.TargetData{
			Identifier: target.Identifier,
			Name:       target.Name,
			Attributes: targetAttributes,
		}
		targetData = append(targetData, td)
	})

	return targetData
}

func (as *AnalyticsService) startSeenTargetsClearingSchedule(ctx context.Context, clearingInterval time.Duration) {
	ticker := time.NewTicker(clearingInterval)

	for {
		select {
		case <-ticker.C:
			as.logger.Debugf("Clearing seen targets")
			as.seenTargets.clear()

		case <-ctx.Done():
			ticker.Stop()
			return
		}
	}
}

func getEvaluationAnalyticKey(event analyticsEvent) string {
	return fmt.Sprintf("%s-%s-%s-%s", event.featureConfig.Feature, event.variation.Identifier, event.variation.Value, globalTarget)
}
