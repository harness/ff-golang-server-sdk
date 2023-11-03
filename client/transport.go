package client

import "net/http"

// HeadersFn is a function type that provides headers dynamically.
type HeadersFn func() (map[string]string, error)

// customTransport wraps an http.RoundTripper and allows adding headers dynamically. This means we can still use
// the goretryable-http client's transport, which does the retrying for us.
type customTransport struct {
	baseTransport http.RoundTripper
	getHeaders    HeadersFn
}

func (t *customTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Retrieve the headers using the provided function.
	headers, err := t.getHeaders()
	if err != nil {
		return nil, err
	}

	// Add the headers to the request.
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// Call the base transport's RoundTrip method.
	return t.baseTransport.RoundTrip(req)
}
