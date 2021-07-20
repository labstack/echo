package middleware

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/lestrrat-go/jwx/jws"
	"github.com/lestrrat-go/jwx/jwt"
	"github.com/stretchr/testify/require"
	"github.com/xenitab/dispans/server"
	"golang.org/x/oauth2"
)

func TestOIDCDiscoveryHandler(t *testing.T) {
	op := server.NewTesting(t)
	defer op.Close(t)

	handler := func(c echo.Context) error {
		token, ok := c.Get("user").(jwt.Token)
		if !ok {
			return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
		}

		claims, err := token.AsMap(c.Request().Context())
		if err != nil {
			return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
		}

		return c.JSON(http.StatusOK, claims)
	}

	e := echo.New()
	h := JWTWithConfig(JWTConfig{
		ParseTokenFunc: OIDCDiscovery(OIDCDiscoveryConfig{
			Issuer:            op.GetURL(t),
			RequiredAudience:  "test-client",
			RequiredTokenType: "JWT+AT",
		}),
	})(handler)

	// Test without authentication
	reqNoAuth := httptest.NewRequest(http.MethodGet, "/", nil)
	recNoAuth := httptest.NewRecorder()
	cNoAuth := e.NewContext(reqNoAuth, recNoAuth)

	err := h(cNoAuth)
	require.Error(t, err)

	// Test with authentication
	token := op.GetToken(t)
	testOIDCDiscoveryHandlerWithAuthentication(t, token, h, e)
	testOIDCDiscoveryHandlerWithIDTokenFailure(t, token, h, e)

	// Test with rotated key
	op.RotateKeys(t)
	tokenWithRotatedKey := op.GetToken(t)
	testOIDCDiscoveryHandlerWithAuthentication(t, tokenWithRotatedKey, h, e)

}

func testOIDCDiscoveryHandlerWithAuthentication(t *testing.T, token *oauth2.Token, restrictedHandler echo.HandlerFunc, e *echo.Echo) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	token.Valid()
	token.SetAuthHeader(req)

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := restrictedHandler(c)
	require.NoError(t, err)

	res := rec.Result()

	require.Equal(t, http.StatusOK, res.StatusCode)
}

func testOIDCDiscoveryHandlerWithIDTokenFailure(t *testing.T, token *oauth2.Token, restrictedHandler echo.HandlerFunc, e *echo.Echo) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	idToken, ok := token.Extra("id_token").(string)
	require.True(t, ok)

	token.AccessToken = idToken

	token.SetAuthHeader(req)

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := restrictedHandler(c)
	require.Error(t, err)
	require.Contains(t, err.Error(), "type \"JWT+AT\" required")
}

func TestOIDCDiscovery(t *testing.T) {
	op := server.NewTesting(t)
	defer op.Close(t)

	cases := []struct {
		testDescription string
		config          OIDCDiscoveryConfig
		expectPanic     bool
	}{
		{
			testDescription: "valid issuer doesn't panic",
			config: OIDCDiscoveryConfig{
				Issuer: op.GetURL(t),
			},
			expectPanic: false,
		},
		{
			testDescription: "valid issuer, invalid DiscoveryUri panics",
			config: OIDCDiscoveryConfig{
				Issuer:       op.GetURL(t),
				DiscoveryUri: "http://foo.bar/baz",
			},
			expectPanic: true,
		},
		{
			testDescription: "valid issuer, invalid JwksUri panics",
			config: OIDCDiscoveryConfig{
				Issuer:  op.GetURL(t),
				JwksUri: "http://foo.bar/baz",
			},
			expectPanic: true,
		},
		{
			testDescription: "empty config panics",
			config:          OIDCDiscoveryConfig{},
			expectPanic:     true,
		},
		{
			testDescription: "fake issuer panics",
			config: OIDCDiscoveryConfig{
				Issuer: "http://foo.bar/baz",
			},
			expectPanic: true,
		},
	}

	for i, c := range cases {
		t.Logf("Test iteration %d: %s", i, c.testDescription)
		if c.expectPanic {
			require.Panics(t, func() { OIDCDiscovery(c.config) })
		} else {
			require.NotPanics(t, func() { OIDCDiscovery(c.config) })
		}
	}
}

func TestNewOIDCKeyHandler(t *testing.T) {
	op := server.NewTesting(t)
	issuer := op.GetURL(t)
	discoveryUri := getDiscoveryUriFromIssuer(issuer)
	jwksUri, err := getJwksUriFromDiscoveryUri(discoveryUri, 10*time.Millisecond)
	require.NoError(t, err)

	oidcKeyHandler, err := newOIDCKeyHandler(jwksUri, 10*time.Millisecond)
	require.NoError(t, err)

	keySet1 := oidcKeyHandler.getKeySet()
	require.Equal(t, 1, keySet1.Len())

	expectedKey1, ok := keySet1.Get(0)
	require.True(t, ok)

	token1 := op.GetToken(t)
	keyID1, err := getKeyIDFromTokenString(token1.AccessToken)
	require.NoError(t, err)

	// Test valid key id
	key1, err := oidcKeyHandler.getByKeyID(keyID1, false)
	require.NoError(t, err)
	require.Equal(t, expectedKey1, key1)

	// Test invalid key id
	_, err = oidcKeyHandler.getByKeyID("foo", false)
	require.Error(t, err)

	// Test with rotated keys
	op.RotateKeys(t)

	token2 := op.GetToken(t)
	keyID2, err := getKeyIDFromTokenString(token2.AccessToken)
	require.NoError(t, err)

	key2, err := oidcKeyHandler.getByKeyID(keyID2, false)
	require.NoError(t, err)

	keySet2 := oidcKeyHandler.getKeySet()
	require.Equal(t, 1, keySet2.Len())

	expectedKey2, ok := keySet2.Get(0)
	require.True(t, ok)

	require.Equal(t, expectedKey2, key2)

	// Test that old key doesn't match new key
	require.NotEqual(t, key1, key2)

	// Validate that error is returned when using fake jwks uri
	_, err = newOIDCKeyHandler("http://foo.bar/baz", 10*time.Millisecond)
	require.Error(t, err)

	// Validate that error is returned when keys are rotated,
	// new token with new key and jwks uri isn't accessible
	op.RotateKeys(t)
	token3 := op.GetToken(t)
	keyID3, err := getKeyIDFromTokenString(token3.AccessToken)
	require.NoError(t, err)
	op.Close(t)
	_, err = oidcKeyHandler.getByKeyID(keyID3, false)
	require.Error(t, err)
}

func TestGetHeadersFromTokenString(t *testing.T) {
	key, _ := testNewOIDCKey(t)

	// Test with KeyID and Type
	token1 := jwt.New()
	token1.Set("foo", "bar")

	headers1 := jws.NewHeaders()
	headers1.Set(jws.TypeKey, "JWT")

	signedTokenBytes1, err := jwt.Sign(token1, jwa.ES384, key, jwt.WithHeaders(headers1))
	require.NoError(t, err)

	signedToken1 := string(signedTokenBytes1)
	parsedHeaders1, err := getHeadersFromTokenString(signedToken1)
	require.NoError(t, err)

	require.Equal(t, key.KeyID(), parsedHeaders1.KeyID())
	require.Equal(t, headers1.Type(), parsedHeaders1.Type())

	// Test with empty headers
	payload1 := `{"foo":"bar"}`

	headers2 := jws.NewHeaders()

	signedTokenBytes2, err := jws.Sign([]byte(payload1), jwa.ES384, key, jws.WithHeaders(headers2))
	require.NoError(t, err)

	signedToken2 := string(signedTokenBytes2)
	parsedHeaders2, err := getHeadersFromTokenString(signedToken2)
	require.NoError(t, err)

	require.Empty(t, parsedHeaders2.Type())

	// Test with multiple signatures
	payload2 := `{"foo":"bar"}`

	signer1, err := jws.NewSigner(jwa.ES384)
	require.NoError(t, err)
	signer2, err := jws.NewSigner(jwa.ES384)
	require.NoError(t, err)

	signedTokenBytes3, err := jws.SignMulti([]byte(payload2), jws.WithSigner(signer1, key, nil, nil), jws.WithSigner(signer2, key, nil, nil))
	require.NoError(t, err)

	signedToken3 := string(signedTokenBytes3)

	_, err = getHeadersFromTokenString(signedToken3)
	require.Error(t, err)
	require.Equal(t, "more than one signature in token", err.Error())

	// Test with non-token string
	_, err = getHeadersFromTokenString("foo")
	require.Error(t, err)
	require.Contains(t, err.Error(), "unable to parse tokenString")
}

func TestGetKeyIDFromTokenString(t *testing.T) {
	key, _ := testNewOIDCKey(t)

	// Test with KeyID
	token1 := jwt.New()
	token1.Set("foo", "bar")

	headers1 := jws.NewHeaders()

	signedTokenBytes1, err := jwt.Sign(token1, jwa.ES384, key, jwt.WithHeaders(headers1))
	require.NoError(t, err)

	signedToken1 := string(signedTokenBytes1)
	keyID, err := getKeyIDFromTokenString(signedToken1)
	require.NoError(t, err)

	require.Equal(t, key.KeyID(), keyID)

	// Test without KeyID
	keyWithoutKeyID := key
	err = keyWithoutKeyID.Remove(jwk.KeyIDKey)
	require.NoError(t, err)

	token2 := jwt.New()
	token2.Set("foo", "bar")

	headers2 := jws.NewHeaders()

	signedTokenBytes2, err := jwt.Sign(token2, jwa.ES384, keyWithoutKeyID, jwt.WithHeaders(headers2))
	require.NoError(t, err)

	signedToken2 := string(signedTokenBytes2)
	_, err = getKeyIDFromTokenString(signedToken2)
	require.Error(t, err)
	require.Equal(t, "token header does not contain key id (kid)", err.Error())

	// Test with non-token string
	_, err = getKeyIDFromTokenString("foo")
	require.Error(t, err)
	require.Contains(t, err.Error(), "unable to parse tokenString")
}

func TestGetTokenTypeFromTokenString(t *testing.T) {
	key, _ := testNewOIDCKey(t)

	// Test with Type
	token1 := jwt.New()
	token1.Set("foo", "bar")

	headers1 := jws.NewHeaders()
	headers1.Set(jws.TypeKey, "foo")

	signedTokenBytes1, err := jwt.Sign(token1, jwa.ES384, key, jwt.WithHeaders(headers1))
	require.NoError(t, err)

	signedToken1 := string(signedTokenBytes1)
	tokenType, err := getTokenTypeFromTokenString(signedToken1)
	require.NoError(t, err)

	require.Equal(t, headers1.Type(), tokenType)

	// Test without KeyID
	payload1 := `{"foo":"bar"}`

	signer1, err := jws.NewSigner(jwa.ES384)
	require.NoError(t, err)

	signedTokenBytes2, err := jws.SignMulti([]byte(payload1), jws.WithSigner(signer1, key, nil, nil))
	require.NoError(t, err)

	signedToken2 := string(signedTokenBytes2)
	_, err = getTokenTypeFromTokenString(signedToken2)
	require.Error(t, err)
	require.Equal(t, "token header does not contain type (typ)", err.Error())

	// Test with non-token string
	_, err = getTokenTypeFromTokenString("foo")
	require.Error(t, err)
	require.Contains(t, err.Error(), "unable to parse tokenString")
}

func TestIsTokenAudienceValid(t *testing.T) {
	cases := []struct {
		testDescription  string
		requiredAudience string
		tokenAudiences   []string
		expectedResult   bool
	}{
		{
			testDescription:  "empty requiredAudience, empty tokenAudiences",
			requiredAudience: "",
			tokenAudiences:   []string{},
			expectedResult:   true,
		},
		{
			testDescription:  "empty requiredAudience, one tokenAudiences",
			requiredAudience: "",
			tokenAudiences:   []string{"foo"},
			expectedResult:   true,
		},
		{
			testDescription:  "empty requiredAudience, two tokenAudiences",
			requiredAudience: "",
			tokenAudiences:   []string{"foo", "bar"},
			expectedResult:   true,
		},
		{
			testDescription:  "empty requiredAudience, three tokenAudiences",
			requiredAudience: "",
			tokenAudiences:   []string{"foo", "bar", "baz"},
			expectedResult:   true,
		},
		{
			testDescription:  "one tokenAudiences, same as requiredAudience",
			requiredAudience: "foo",
			tokenAudiences:   []string{"foo"},
			expectedResult:   true,
		},
		{
			testDescription:  "two tokenAudiences, first same as requiredAudience",
			requiredAudience: "foo",
			tokenAudiences:   []string{"foo", "bar"},
			expectedResult:   true,
		},
		{
			testDescription:  "two tokenAudiences, second same as requiredAudience",
			requiredAudience: "bar",
			tokenAudiences:   []string{"foo", "bar"},
			expectedResult:   true,
		},
		{
			testDescription:  "three tokenAudiences, third same as requiredAudience",
			requiredAudience: "baz",
			tokenAudiences:   []string{"foo", "bar", "baz"},
			expectedResult:   true,
		},
		{
			testDescription:  "set requiredAudience, empty tokenAudiences",
			requiredAudience: "foo",
			tokenAudiences:   []string{},
			expectedResult:   false,
		},
		{
			testDescription:  "one tokenAudience, not same as requiredAudience",
			requiredAudience: "foo",
			tokenAudiences:   []string{"bar"},
			expectedResult:   false,
		},
		{
			testDescription:  "two tokenAudience, none same as requiredAudience",
			requiredAudience: "foo",
			tokenAudiences:   []string{"bar", "baz"},
			expectedResult:   false,
		},
		{
			testDescription:  "three tokenAudience, none same as requiredAudience",
			requiredAudience: "foo",
			tokenAudiences:   []string{"bar", "baz", "foobar"},
			expectedResult:   false,
		},
	}

	for i, c := range cases {
		t.Logf("Test iteration %d: %s", i, c.testDescription)
		result := isTokenAudienceValid(c.requiredAudience, c.tokenAudiences)
		require.Equal(t, c.expectedResult, result)
	}
}

func TestTokenExpirationValid(t *testing.T) {
	cases := []struct {
		testDescription string
		expiration      time.Time
		allowedDrift    time.Duration
		expectedResult  bool
	}{
		{
			testDescription: "expires now, 50 millisecond drift allowed",
			expiration:      time.Now(),
			allowedDrift:    50 * time.Millisecond,
			expectedResult:  true,
		},
		{
			testDescription: "expires now, 10 second drift allowed",
			expiration:      time.Now(),
			allowedDrift:    10 * time.Second,
			expectedResult:  true,
		},
		{
			testDescription: "expires in one hour, 10 second drift allowed",
			expiration:      time.Now().Add(1 * time.Hour),
			allowedDrift:    10 * time.Second,
			expectedResult:  true,
		},
		{
			testDescription: "expired 5 seconds ago, 10 second drift allowed",
			expiration:      time.Now().Add(-5 * time.Second),
			allowedDrift:    10 * time.Second,
			expectedResult:  true,
		},
		{
			testDescription: "expired 11 seconds ago, 10 second drift allowed",
			expiration:      time.Now().Add(-11 * time.Second),
			allowedDrift:    10 * time.Second,
			expectedResult:  false,
		},
		{
			testDescription: "expires now, no drift",
			expiration:      time.Now(),
			allowedDrift:    0,
			expectedResult:  false,
		},
		{
			testDescription: "expired an hour ago, no drift",
			expiration:      time.Now().Add(-1 * time.Hour),
			allowedDrift:    0,
			expectedResult:  false,
		},
		{
			testDescription: "expired an hour ago, 10 second drift",
			expiration:      time.Now().Add(-1 * time.Hour),
			allowedDrift:    10 * time.Second,
			expectedResult:  false,
		},
	}

	for i, c := range cases {
		t.Logf("Test iteration %d: %s", i, c.testDescription)
		result := isTokenExpirationValid(c.expiration, c.allowedDrift)
		require.Equal(t, c.expectedResult, result)
	}
}

func TestIsTokenIssuerValid(t *testing.T) {
	cases := []struct {
		testDescription string
		requiredIssuer  string
		tokenIssuer     string
		expectedResult  bool
	}{
		{
			testDescription: "both requiredIssuer and tokenIssuer are the same",
			requiredIssuer:  "foo",
			tokenIssuer:     "foo",
			expectedResult:  true,
		},
		{
			testDescription: "requiredIssuer and tokenIssuer are not the same",
			requiredIssuer:  "foo",
			tokenIssuer:     "bar",
			expectedResult:  false,
		},
		{
			testDescription: "both requiredIssuer and tokenIssuer are empty",
			requiredIssuer:  "",
			tokenIssuer:     "",
			expectedResult:  false,
		},
		{
			testDescription: "requiredIssuer is empty and tokenIssuer is set",
			requiredIssuer:  "",
			tokenIssuer:     "foo",
			expectedResult:  false,
		},
		{
			testDescription: "requiredIssuer is set and tokenIssuer is empty",
			requiredIssuer:  "foo",
			tokenIssuer:     "",
			expectedResult:  false,
		},
	}

	for i, c := range cases {
		t.Logf("Test iteration %d: %s", i, c.testDescription)
		result := isTokenIssuerValid(c.requiredIssuer, c.tokenIssuer)
		require.Equal(t, c.expectedResult, result)
	}
}

func TestIsTokenTypeValid(t *testing.T) {
	cases := []struct {
		testDescription   string
		requiredTokenType string
		tokenType         string
		expectedResult    bool
	}{
		{
			testDescription:   "both requiredTokenType and tokenType are empty",
			requiredTokenType: "",
			tokenType:         "",
			expectedResult:    true,
		},
		{
			testDescription:   "requiredTokenType is empty and tokenType is set",
			requiredTokenType: "",
			tokenType:         "foo",
			expectedResult:    true,
		},
		{
			testDescription:   "both requiredTokenType and tokenType are set to the same",
			requiredTokenType: "foo",
			tokenType:         "foo",
			expectedResult:    true,
		},
		{
			testDescription:   "requiredTokenType and tokenType are set to different",
			requiredTokenType: "foo",
			tokenType:         "bar",
			expectedResult:    false,
		},
		{
			testDescription:   "requiredTokenType and tokenType are set to different but tokenType contains requiredTokenType",
			requiredTokenType: "foo",
			tokenType:         "foobar",
			expectedResult:    false,
		},
	}

	for i, c := range cases {
		t.Logf("Test iteration %d: %s", i, c.testDescription)

		key, _ := testNewOIDCKey(t)
		payload := `{"foo":"bar"}`

		signer, err := jws.NewSigner(jwa.ES384)
		require.NoError(t, err)

		var signedTokenBytes []byte
		if c.tokenType == "" {
			signedTokenBytes, err = jws.SignMulti([]byte(payload), jws.WithSigner(signer, key, nil, nil))
			require.NoError(t, err)
		} else {
			headers := jws.NewHeaders()
			headers.Set(jws.TypeKey, c.tokenType)

			signedTokenBytes, err = jws.SignMulti([]byte(payload), jws.WithSigner(signer, key, nil, headers))
			require.NoError(t, err)
		}

		token := string(signedTokenBytes)

		result := isTokenTypeValid(c.requiredTokenType, token)
		require.Equal(t, c.expectedResult, result)
	}
}

func TestGetTokenFromString(t *testing.T) {
	op := server.NewTesting(t)
	defer op.Close(t)

	issuer := op.GetURL(t)
	discoveryUri := getDiscoveryUriFromIssuer(issuer)
	jwksUri, err := getJwksUriFromDiscoveryUri(discoveryUri, 10*time.Millisecond)
	require.NoError(t, err)

	oidcKeyHandler, err := newOIDCKeyHandler(jwksUri, 10*time.Millisecond)
	require.NoError(t, err)

	validKey, ok := oidcKeyHandler.getKeySet().Get(0)
	require.True(t, ok)

	validAccessToken := op.GetToken(t).AccessToken
	require.NotEmpty(t, validAccessToken)

	validIDToken, ok := op.GetToken(t).Extra("id_token").(string)
	require.True(t, ok)
	require.NotEmpty(t, validIDToken)

	invalidKey, invalidPubKey := testNewOIDCKey(t)

	invalidToken := jwt.New()
	invalidToken.Set("foo", "bar")

	invalidHeaders := jws.NewHeaders()
	invalidHeaders.Set(jws.TypeKey, "JWT")

	invalidTokenBytes, err := jwt.Sign(invalidToken, jwa.ES384, invalidKey, jwt.WithHeaders(invalidHeaders))
	require.NoError(t, err)

	invalidSignedToken := string(invalidTokenBytes)

	cases := []struct {
		testDescription string
		tokenString     string
		key             jwk.Key
		expectedError   bool
	}{
		{
			testDescription: "valid access token, valid key",
			tokenString:     validAccessToken,
			key:             validKey,
			expectedError:   false,
		},
		{
			testDescription: "valid id token, valid key",
			tokenString:     validIDToken,
			key:             validKey,
			expectedError:   false,
		},
		{
			testDescription: "empty string, valid key",
			tokenString:     "",
			key:             validKey,
			expectedError:   true,
		},
		{
			testDescription: "random string, valid key",
			tokenString:     "foobar",
			key:             validKey,
			expectedError:   true,
		},
		{
			testDescription: "invalid token, valid key",
			tokenString:     invalidSignedToken,
			key:             validKey,
			expectedError:   true,
		},
		{
			testDescription: "invalid token, invalid key",
			tokenString:     invalidSignedToken,
			key:             invalidPubKey,
			expectedError:   false,
		},
	}

	for i, c := range cases {
		t.Logf("Test iteration %d: %s", i, c.testDescription)

		token, err := getTokenFromString(c.tokenString, c.key)
		if c.expectedError {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
			require.NotEmpty(t, token)
		}
	}
}

func testNewOIDCKey(t *testing.T) (jwk.Key, jwk.Key) {
	ecdsaKey, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	require.NoError(t, err)

	key, err := jwk.New(ecdsaKey)
	require.NoError(t, err)

	_, ok := key.(jwk.ECDSAPrivateKey)
	require.True(t, ok)

	thumbprint, err := key.Thumbprint(crypto.SHA256)
	require.NoError(t, err)

	keyID := fmt.Sprintf("%x", thumbprint)
	key.Set(jwk.KeyIDKey, keyID)

	pubKey, err := jwk.New(ecdsaKey.PublicKey)
	require.NoError(t, err)

	_, ok = pubKey.(jwk.ECDSAPublicKey)
	require.True(t, ok)

	pubKey.Set(jwk.KeyIDKey, keyID)
	pubKey.Set(jwk.AlgorithmKey, jwa.ES384)

	return key, pubKey
}
