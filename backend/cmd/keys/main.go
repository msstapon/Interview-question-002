// Command keys generates an RS256 keypair into ./secrets for signing JWTs.
// Cross-platform replacement for `openssl genrsa` (no external tooling needed).
package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

func main() {
	dir := "secrets"
	if err := os.MkdirAll(dir, 0o755); err != nil {
		log.Fatalf("mkdir secrets: %v", err)
	}

	privPath := filepath.Join(dir, "jwt_private.pem")
	pubPath := filepath.Join(dir, "jwt_public.pem")
	// Idempotent: keep existing keys so re-runs (e.g. every deploy) don't rotate them
	// and invalidate live tokens. Pass FORCE=1 to regenerate.
	if os.Getenv("FORCE") == "" && fileExists(privPath) && fileExists(pubPath) {
		fmt.Println("RS256 keypair already exists in ./secrets — skipping (set FORCE=1 to regenerate)")
		return
	}

	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Fatalf("generate key: %v", err)
	}

	privPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(priv),
	})
	pubBytes, err := x509.MarshalPKIXPublicKey(&priv.PublicKey)
	if err != nil {
		log.Fatalf("marshal public key: %v", err)
	}
	pubPEM := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: pubBytes})

	if err := os.WriteFile(privPath, privPEM, 0o600); err != nil {
		log.Fatalf("write private: %v", err)
	}
	if err := os.WriteFile(pubPath, pubPEM, 0o644); err != nil {
		log.Fatalf("write public: %v", err)
	}
	fmt.Println("RS256 keypair written to ./secrets/{jwt_private.pem,jwt_public.pem}")
}

func fileExists(p string) bool {
	info, err := os.Stat(p)
	return err == nil && !info.IsDir()
}
