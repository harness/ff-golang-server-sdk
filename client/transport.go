package client

import "net/http"

// HeadersFn is a function type that provides headers dynamically.
type HeadersFn func() (map[string]string, error)

// customTransport wraps an http.RoundTripper and allows adding headers.
type customTransport struct {
	baseTransport http.RoundTripper // This will be the retryablehttp transport.
	getHeaders    HeadersFn
}

// RoundTrip executes a single HTTP transaction, adding the headers dynamically.
func (t *customTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Retrieve the headers using the provided function.
	headers, err := t.getHeaders()
	if err != nil {
		return nil, err // Handle the error as you see fit.
	}

	// Add the headers to the request.
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// Call the base transport's RoundTrip method.
	return t.baseTransport.RoundTrip(req)
}
