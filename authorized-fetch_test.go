package sherlock

import (
	"crypto/rand"
	"crypto/rsa"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/benpate/remote"
	"github.com/stretchr/testify/require"
)

func TestAuthorizedFetch_MissingPublicKey(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.Nil(t, err)

	// No publicKeyID -> empty option (no ModifyRequest hook)
	option := AuthorizedFetch("", privateKey)
	require.Nil(t, option.ModifyRequest)
}

func TestAuthorizedFetch_MissingPrivateKey(t *testing.T) {
	// No privateKey -> empty option (no ModifyRequest hook)
	option := AuthorizedFetch("https://example.com/actor#key", nil)
	require.Nil(t, option.ModifyRequest)
}

func TestAuthorizedFetch_Valid(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.Nil(t, err)

	option := AuthorizedFetch("https://example.com/actor#key", privateKey)
	require.NotNil(t, option.ModifyRequest)

	// The hook should sign the request and return nil (meaning "send it normally").
	request, err := http.NewRequest(http.MethodGet, "https://example.com/inbox", nil)
	require.Nil(t, err)

	txn := remote.Get("https://example.com/inbox")
	response, err := option.ModifyRequest(txn, request)
	require.Nil(t, err)
	require.Nil(t, response)

	// A Signature header should have been added to the outbound request.
	require.NotEmpty(t, request.Header.Get("Signature"))
}

// TestAuthorizedFetch_CannotSign verifies that a request is refused, and never sent, when its key cannot sign it
func TestAuthorizedFetch_CannotSign(t *testing.T) {

	received := 0

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, _ *http.Request) {
		received++
		writer.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	// A string is not a key that hannibal can sign with
	option := AuthorizedFetch("https://example.com/actor#key", "not a key")

	err := remote.Get(server.URL).AllowPrivateIPs(true).With(option).Send()

	require.Error(t, err)
	require.Zero(t, received)
}
