## Flag Cleanup (beta)
The go sdk supports automated code cleanup using the Flag Cleanup drone plugin. See [here](https://github.com/harness/flag_cleanup) for detailed docs on usage of the plugin.

### Run example
1. View the [example file](/examples/code_cleanup/example.go) and observe the if block using our feature flag ```harnessappdemodarkmode```
2. Run the flag cleanup plugin. This can be done by running the make command

```make flag_cleanup_demo```

or by running the docker command directly from the root folder of this repository:

```docker run -v ${PWD}:/go-sdk -e PLUGIN_DEBUG=true -e PLUGIN_PATH_TO_CODEBASE="/go-sdk/examples/code_cleanup" -e PLUGIN_PATH_TO_CONFIGURATIONS="/go-sdk/examples/code_cleanup/config" -e PLUGIN_LANGUAGE="go" -e PLUGIN_SUBSTITUTIONS="stale_flag_name=harnessappdemodarkmode,treated=true" harness/flag_cleanup:latest```

3. Observe that the `if else` block has been removed from the code and the flag is now treated as globally true.
