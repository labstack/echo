package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/lestrrat-go/jwx/jws"
	"github.com/lestrrat-go/jwx/jwt"
)

type (
	// OIDCDiscoveryConfig defines the config for OIDC Discovery for the `ParseTokenFunc` for the `JWT` middleware
	OIDCDiscoveryConfig struct {
		// Issuer is the authority that issues the tokens
		Issuer string

		// DiscoveryUri is where the `jwks_uri` will be grabbed
		// Defaults to `fmt.Sprintf("%s/.well-known/openid-configuration", strings.TrimSuffix(issuer, "/"))`
		DiscoveryUri string

		// JwksUri is used to download the public key(s)
		// Defaults to the `jwks_uri` from the response of DiscoveryUri
		JwksUri string

		// RequiredTokenType is used if only specific tokens should be allowed.
		// Default is empty string `""` and means all token types are allowed.
		// Use case could be to configure this if the TokenType (set in the header of the JWT)
		// should be `JWT` or maybe even `JWT+AT` to diffirentiate between access tokens and
		// id tokens. Not all providers support or use this.
		RequiredTokenType string

		// RequiredAudience is used to require a specific Audience `aud` in the claims.
		// Default to empty string `""` and means all audiences are allowed.
		RequiredAudience string

		// JwksFetchTimeout sets the context timeout when downloading the jwks
		// Defaults to 5 seconds
		JwksFetchTimeout time.Duration

		// AllowedTokenDrift adds the duration to the token expiration to allow
		// for time drift between parties.
		// Defaults to 10 seconds
		AllowedTokenDrift time.Duration

		// oidcKeyHandler handles jwks
		oidcKeyHandler *oidcKeyHandler
	}
)

// OIDCDiscovery returns an OpenID Connect (OIDC) discovery `ParseTokenFunc` to be used
// with the `JWT` middleware.
// See: https://openid.net/connect/
func OIDCDiscovery(config OIDCDiscoveryConfig) func(auth string, c echo.Context) (interface{}, error) {
	if config.Issuer == "" {
		panic("echo: oidc middleware requires Issuer")
	}
	if config.DiscoveryUri == "" {
		config.DiscoveryUri = getDiscoveryUriFromIssuer(config.Issuer)
	}
	if config.JwksUri == "" {
		jwksUri, err := getJwksUriFromDiscoveryUri(config.DiscoveryUri, 5*time.Second)
		if err != nil {
			panic(fmt.Sprintf("echo: oidc middleware unable to fetch JwksUri from DiscoveryUri (%s): %v", config.DiscoveryUri, err))
		}
		config.JwksUri = jwksUri
	}
	if config.JwksFetchTimeout == 0 {
		config.JwksFetchTimeout = 5 * time.Second
	}
	if config.AllowedTokenDrift == 0 {
		config.AllowedTokenDrift = 10 * time.Second
	}

	oidcKeyHandler, err := newOIDCKeyHandler(config.JwksUri, config.JwksFetchTimeout)
	if err != nil {
		panic(fmt.Sprintf("echo: oidc middleware unable to initialize oidcKeyHandler: %v", err))
	}

	config.oidcKeyHandler = oidcKeyHandler

	return config.parseToken
}

func (config *OIDCDiscoveryConfig) parseToken(auth string, c echo.Context) (interface{}, error) {
	keyID, err := getKeyIDFromTokenString(auth)
	if err != nil {
		return nil, err
	}

	tokenTypeValid := isTokenTypeValid(config.RequiredTokenType, auth)
	if !tokenTypeValid {
		return nil, fmt.Errorf("token type %q required", config.RequiredTokenType)
	}

	key, err := config.oidcKeyHandler.getByKeyID(keyID, false)
	if err != nil {
		return nil, err
	}

	token, err := getAndValidateTokenFromString(auth, key)
	if err != nil {
		return nil, err
	}

	validExpiration := isTokenExpirationValid(token.Expiration(), config.AllowedTokenDrift)
	if !validExpiration {
		return nil, fmt.Errorf("token has expired: %s", token.Expiration())
	}

	validIssuer := isTokenIssuerValid(config.Issuer, token.Issuer())
	if !validIssuer {
		return nil, fmt.Errorf("required issuer %q was not found, received: %s", config.Issuer, token.Issuer())
	}

	validAudience := isTokenAudienceValid(config.RequiredAudience, token.Audience())
	if !validAudience {
		return nil, fmt.Errorf("required audience %q was not found, received: %v", config.RequiredAudience, token.Audience())
	}

	return token, nil
}

type oidcKeyHandler struct {
	sync.RWMutex
	jwksURI      string
	keySet       jwk.Set
	fetchTimeout time.Duration
}

func newOIDCKeyHandler(jwksUri string, fetchTimeout time.Duration) (*oidcKeyHandler, error) {
	h := &oidcKeyHandler{
		jwksURI:      jwksUri,
		fetchTimeout: fetchTimeout,
	}

	err := h.updateKeySet()
	if err != nil {
		return nil, err
	}

	return h, nil
}

func (h *oidcKeyHandler) updateKeySet() error {
	ctx, cancel := context.WithTimeout(context.Background(), h.fetchTimeout)
	defer cancel()
	keySet, err := jwk.Fetch(ctx, h.jwksURI)
	if err != nil {
		return fmt.Errorf("Unable to fetch keys from %q: %v", h.jwksURI, err)
	}

	h.Lock()
	h.keySet = keySet
	h.Unlock()

	return nil
}

func (h *oidcKeyHandler) getKeySet() jwk.Set {
	h.RLock()
	defer h.RUnlock()
	return h.keySet
}

func (h *oidcKeyHandler) getByKeyID(keyID string, retry bool) (jwk.Key, error) {
	keySet := h.getKeySet()
	key, found := keySet.LookupKeyID(keyID)

	if !found && !retry {
		err := h.updateKeySet()
		if err != nil {
			return nil, fmt.Errorf("unable to update key set for key %q: %v", keyID, err)
		}

		return h.getByKeyID(keyID, true)
	}

	if !found && retry {
		return nil, fmt.Errorf("unable to find key %q", keyID)
	}

	return key, nil
}

func getDiscoveryUriFromIssuer(issuer string) string {
	return fmt.Sprintf("%s/.well-known/openid-configuration", strings.TrimSuffix(issuer, "/"))
}

func getJwksUriFromDiscoveryUri(discoveryUri string, fetchTimeout time.Duration) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), fetchTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, discoveryUri, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("Accept", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	bodyBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	err = res.Body.Close()
	if err != nil {
		return "", err
	}

	var discoveryData struct {
		JwksUri string `json:"jwks_uri"`
	}

	err = json.Unmarshal(bodyBytes, &discoveryData)
	if err != nil {
		return "", err
	}

	if discoveryData.JwksUri == "" {
		return "", fmt.Errorf("JwksUri is empty")
	}

	return discoveryData.JwksUri, nil
}

func getKeyIDFromTokenString(tokenString string) (string, error) {
	headers, err := getHeadersFromTokenString(tokenString)
	if err != nil {
		return "", err
	}

	keyID := headers.KeyID()
	if keyID == "" {
		return "", fmt.Errorf("token header does not contain key id (kid)")
	}

	return keyID, nil
}

func getTokenTypeFromTokenString(tokenString string) (string, error) {
	headers, err := getHeadersFromTokenString(tokenString)
	if err != nil {
		return "", err
	}

	tokenType := headers.Type()
	if tokenType == "" {
		return "", fmt.Errorf("token header does not contain type (typ)")
	}

	return tokenType, nil
}

func getHeadersFromTokenString(tokenString string) (jws.Headers, error) {
	msg, err := jws.ParseString(tokenString)
	if err != nil {
		return nil, fmt.Errorf("unable to parse tokenString: %w", err)
	}

	signatures := msg.Signatures()
	if len(signatures) != 1 {
		return nil, fmt.Errorf("more than one signature in token")
	}

	headers := signatures[0].ProtectedHeaders()

	return headers, nil
}

func isTokenAudienceValid(requiredAudience string, audiences []string) bool {
	if requiredAudience == "" {
		return true
	}

	for _, audience := range audiences {
		if audience == requiredAudience {
			return true
		}
	}

	return false
}

func isTokenExpirationValid(expiration time.Time, allowedDrift time.Duration) bool {
	expirationWithAllowedDrift := expiration.Round(0).Add(allowedDrift)

	return expirationWithAllowedDrift.After(time.Now())
}

func isTokenIssuerValid(requiredIssuer string, tokenIssuer string) bool {
	if requiredIssuer == "" {
		return false
	}

	return tokenIssuer == requiredIssuer
}

func isTokenTypeValid(requiredTokenType string, tokenString string) bool {
	if requiredTokenType == "" {
		return true
	}

	tokenType, err := getTokenTypeFromTokenString(tokenString)
	if err != nil {
		return false
	}

	if tokenType != requiredTokenType {
		return false
	}

	return true
}

func getAndValidateTokenFromString(tokenString string, key jwk.Key) (jwt.Token, error) {
	keySet := getKeySetFromKey(key)

	token, err := jwt.ParseString(tokenString, jwt.WithKeySet(keySet))
	if err != nil {
		return nil, err
	}

	return token, nil
}

func getKeySetFromKey(key jwk.Key) jwk.Set {
	keySet := jwk.NewSet()
	keySet.Add(key)

	return keySet
}
