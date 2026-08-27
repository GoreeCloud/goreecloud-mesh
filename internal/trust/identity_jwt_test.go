package trust

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestIdentityJWTVerifierAcceptsValidServiceToken(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	server := jwksServer(t, &privateKey.PublicKey, "kid-1")
	defer server.Close()
	now := time.Date(2026, 8, 26, 20, 0, 0, 0, time.UTC)
	verifier := NewIdentityJWTVerifier(server.URL)
	verifier.Now = func() time.Time { return now }
	claims := map[string]any{
		"iss": "goreecloud-identity", "aud": "goreecloud-mesh",
		"sub": "service:wardveil-security", "service_id": "wardveil-security",
		"scope": "mesh.evidence.write mesh.evidence.read", "iat": now.Unix(),
		"nbf": now.Add(-time.Second).Unix(), "exp": now.Add(10 * time.Minute).Unix(), "jti": "token-001",
	}
	token := signToken(t, privateKey, "kid-1", claims)
	req := httptest.NewRequest(http.MethodPost, "/v1/evidence/envelopes", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	principal, err := verifier.Verify(req)
	if err != nil {
		t.Fatal(err)
	}
	if principal.ServiceID != "wardveil-security" || principal.Issuer != "goreecloud-identity" || !HasScope(principal, "mesh.evidence.write") {
		t.Fatalf("unexpected principal: %#v", principal)
	}
}

func TestIdentityJWTVerifierRejectsWrongAudienceAndLongLifetime(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	server := jwksServer(t, &privateKey.PublicKey, "kid-1")
	defer server.Close()
	now := time.Now().UTC().Truncate(time.Second)
	verifier := NewIdentityJWTVerifier(server.URL)
	verifier.Now = func() time.Time { return now }
	base := map[string]any{"iss": "goreecloud-identity", "aud": "wrong", "sub": "service:everkeep", "service_id": "everkeep", "scope": "mesh.evidence.write", "iat": now.Unix(), "exp": now.Add(10 * time.Minute).Unix(), "jti": "t1"}
	for name, mutate := range map[string]func(map[string]any){
		"audience": func(c map[string]any) {},
		"lifetime": func(c map[string]any) { c["aud"] = "goreecloud-mesh"; c["exp"] = now.Add(20 * time.Minute).Unix() },
	} {
		t.Run(name, func(t *testing.T) {
			claims := map[string]any{}
			for k, v := range base {
				claims[k] = v
			}
			mutate(claims)
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set("Authorization", "Bearer "+signToken(t, privateKey, "kid-1", claims))
			if _, err := verifier.Verify(req); err == nil {
				t.Fatal("expected verification failure")
			}
		})
	}
}

func TestIdentityJWTVerifierRejectsServiceSubjectMismatch(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	server := jwksServer(t, &privateKey.PublicKey, "kid-1")
	defer server.Close()
	now := time.Now().UTC().Truncate(time.Second)
	verifier := NewIdentityJWTVerifier(server.URL)
	verifier.Now = func() time.Time { return now }
	claims := map[string]any{"iss": "goreecloud-identity", "aud": "goreecloud-mesh", "sub": "service:privacy-shield", "service_id": "wardveil-security", "scope": "mesh.evidence.write", "iat": now.Unix(), "exp": now.Add(5 * time.Minute).Unix(), "jti": "t2"}
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+signToken(t, privateKey, "kid-1", claims))
	if _, err := verifier.Verify(req); err == nil {
		t.Fatal("expected subject/service mismatch rejection")
	}
}

func jwksServer(t *testing.T, publicKey *rsa.PublicKey, kid string) *httptest.Server {
	t.Helper()
	n := base64.RawURLEncoding.EncodeToString(publicKey.N.Bytes())
	e := big.NewInt(int64(publicKey.E)).Bytes()
	eEncoded := base64.RawURLEncoding.EncodeToString(e)
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"keys": []map[string]any{{"kty": "RSA", "use": "sig", "alg": "RS256", "kid": kid, "n": n, "e": eEncoded}}})
	}))
}

func signToken(t *testing.T, key *rsa.PrivateKey, kid string, claims map[string]any) string {
	t.Helper()
	header, _ := json.Marshal(map[string]any{"alg": "RS256", "typ": "JWT", "kid": kid})
	payload, _ := json.Marshal(claims)
	h := base64.RawURLEncoding.EncodeToString(header)
	p := base64.RawURLEncoding.EncodeToString(payload)
	digest := sha256.Sum256([]byte(h + "." + p))
	sig, err := rsa.SignPKCS1v15(rand.Reader, key, crypto.SHA256, digest[:])
	if err != nil {
		t.Fatal(err)
	}
	return h + "." + p + "." + base64.RawURLEncoding.EncodeToString(sig)
}
