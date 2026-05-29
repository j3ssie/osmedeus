package cli

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRetryableHTTPStatus(t *testing.T) {
	for _, code := range []int{408, 429, 500, 502, 503, 504} {
		assert.Truef(t, retryableHTTPStatus(code), "status %d should be retryable", code)
	}
	for _, code := range []int{200, 301, 400, 401, 403, 404, 422} {
		assert.Falsef(t, retryableHTTPStatus(code), "status %d should not be retryable", code)
	}
}

// fetchURLContent should retry transient (5xx) failures and eventually succeed.
func TestFetchURLContentRetriesTransientThenSucceeds(t *testing.T) {
	var attempts int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if atomic.AddInt32(&attempts, 1) < 2 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		_, _ = w.Write([]byte("name: test\nkind: module\n"))
	}))
	defer srv.Close()

	content, err := fetchURLContent(srv.URL, nil)
	require.NoError(t, err)
	assert.Contains(t, string(content), "name: test")
	assert.Equal(t, int32(2), atomic.LoadInt32(&attempts), "should retry once then succeed")
}

// fetchURLContent should NOT retry non-retryable statuses (e.g. 404) so the
// GitHub auth fallback in fetchWorkflowFromURL can kick in promptly.
func TestFetchURLContentDoesNotRetryNonRetryable(t *testing.T) {
	var attempts int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempts, 1)
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	_, err := fetchURLContent(srv.URL, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "HTTP 404")
	assert.Equal(t, int32(1), atomic.LoadInt32(&attempts), "404 should not be retried")
}
