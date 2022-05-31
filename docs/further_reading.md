# Further Reading

Covers advanced topics (different config options and scenarios)

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