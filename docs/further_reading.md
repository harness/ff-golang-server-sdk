# Further Reading

Covers advanced topics (different config options and scenarios)

## Configuration Options
The following configuration options are available to control the behaviour of the SDK.
You can provide options by passing them in when the client is created e.g.

```golang
// Create Options
client, err := harness.NewCfClient(myApiKey, 
	harness.WithURL("https://config.ff.harness.io/api/1.0"), 
	harness.WithEventsURL("https://events.ff.harness.io/api/1.0"), 
	harness.WithPullInterval(1),
	harness.WithStreamEnabled(false))

```

| Name            | Config Option                                                  | Description                                                                                                                                      | default                              |
|-----------------|----------------------------------------------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------|--------------------------------------|
| baseUrl         | harness.WithURL("https://config.ff.harness.io/api/1.0")        | the URL used to fetch feature flag evaluations. You should change this when using the Feature Flag proxy to http://localhost:7000                | https://config.ff.harness.io/api/1.0 |
| eventsUrl       | harness.WithEventsURL("https://events.ff.harness.io/api/1.0"), | the URL used to post metrics data to the feature flag service. You should change this when using the Feature Flag proxy to http://localhost:7000 | https://events.ff.harness.io/api/1.0 |
| pollInterval    | harness.WithPullInterval(1))                                   | when running in stream mode, the interval in minutes that we poll for changes.                                                                   | 60                                   |
| enableStream    | harness.WithStreamEnabled(false),                              | Enable streaming mode.                                                                                                                           | true                                 |
| enableAnalytics | *Not Supported*                                                | Enable analytics.  Metrics data is posted every 60s                                                                                              | *Not Supported*                      |

## Logging Configuration
You can provide your own logger to the SDK, passing it in as a config option.
The following example creates an instance of the logrus logger and provides it as an option.


```golang
logger := logrus.New()

// Create a feature flag client
client, err := harness.NewCfClient(myApiKey, harness.WithLogger(logger))
```


## Recommended reading

[Feature Flag Concepts](https://ngdocs.harness.io/article/7n9433hkc0-cf-feature-flag-overview)

[Feature Flag SDK Concepts](https://ngdocs.harness.io/article/rvqprvbq8f-client-side-and-server-side-sdks)

## Setting up your Feature Flags

[Feature Flags Getting Started](https://ngdocs.harness.io/article/0a2u2ppp8s-getting-started-with-feature-flags)

## Other Variation Types

### String Variation
```golang
client.StringVariation(flagName, &target, "default_string")
```

### Number Variation
```golang
client.NumberVariation(flagName, &target, -1)
```

### JSON Variation
```golang
client.JSONVariation(flagName, &target, types.JSON{"darkmode": false})
```


## Cleanup
Call the close function on the client

```golang
client.Close()
```


## Change default URL

When using your Feature Flag SDKs with a [Harness Relay Proxy](https://ngdocs.harness.io/article/q0kvq8nd2o-relay-proxy) you need to change the default URL.
You can pass the URLs in when creating the client. i.e.

```golang
	client, err := harness.NewCfClient(apiKey,
		harness.WithURL("https://config.feature-flags.uat.harness.io/api/1.0"),
		harness.WithEventsURL("https://event.feature-flags.uat.harness.io/api/1.0"))
```