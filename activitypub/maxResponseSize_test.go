package activitypub

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/benpate/hannibal/streams"
	"github.com/benpate/hannibal/vocab"
	"github.com/stretchr/testify/require"
)

// TestLoad_AtMaxResponseSize confirms a document of exactly streams.MaxResponseSize bytes is read
func TestLoad_AtMaxResponseSize(t *testing.T) {

	server := servePaddedActor(t, streams.MaxResponseSize)
	document, err := New(WithAllowPrivateIPs(true)).Load(server.URL)

	require.NoError(t, err)
	require.Equal(t, server.URL+"/", document.ID())
}

// TestLoad_OverMaxResponseSize confirms a document one byte over streams.MaxResponseSize is refused
func TestLoad_OverMaxResponseSize(t *testing.T) {

	server := servePaddedActor(t, streams.MaxResponseSize+1)
	_, err := New(WithAllowPrivateIPs(true)).Load(server.URL)

	require.Error(t, err)
}

// TestLoad_WithMaxResponseSize confirms the option lowers the cap
func TestLoad_WithMaxResponseSize(t *testing.T) {

	server := servePaddedActor(t, 2048)
	_, err := New(WithAllowPrivateIPs(true), WithMaxResponseSize(1024)).Load(server.URL)

	require.Error(t, err)
}

// TestLoad_WithMaxResponseSize_ZeroUsesRemoteDefault confirms a zero or negative cap reads past
// streams.MaxResponseSize, under remote's 1 GB default
func TestLoad_WithMaxResponseSize_ZeroUsesRemoteDefault(t *testing.T) {

	server := servePaddedActor(t, streams.MaxResponseSize+1)

	for _, value := range []int64{0, -1} {
		document, err := New(WithAllowPrivateIPs(true), WithMaxResponseSize(value)).Load(server.URL)
		require.NoError(t, err, "WithMaxResponseSize(%d)", value)
		require.Equal(t, server.URL+"/", document.ID(), "WithMaxResponseSize(%d)", value)
	}
}

// servePaddedActor starts a server whose actor, identified by its own URL, encodes to the requested number of bytes
func servePaddedActor(t *testing.T, size int) *httptest.Server {

	t.Helper()

	var server *httptest.Server

	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		prefix := `{"id":"` + server.URL + `/","type":"Person","summary":"`
		suffix := `"}`
		w.Header().Set("Content-Type", vocab.ContentTypeActivityPub)
		_, _ = w.Write([]byte(prefix + strings.Repeat("a", size-len(prefix)-len(suffix)) + suffix))
	}))

	t.Cleanup(server.Close)
	return server
}
