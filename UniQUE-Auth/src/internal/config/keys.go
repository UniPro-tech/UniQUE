package config

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// ParseKeys walks the public and private key directories recursively, parses
// RSA keys and pairs them by normalized filename (e.g. "rsa_public.pem" and
// "rsa_private.pem" -> name "rsa"). Returns only pairs that have both
// public and private keys.
func ParseKeys(cfg KeyPathsConfig) ([]KeyPairConfig, error) {
	if cfg.KeyType != "RSA" {
		return nil, fmt.Errorf("unsupported key type: %s", cfg.KeyType)
	}

	pubMap := map[string]*rsa.PublicKey{}
	privMap := map[string]*rsa.PrivateKey{}

	isKeyFile := func(name string) bool {
		ext := strings.ToLower(filepath.Ext(name))
		switch ext {
		case ".pem", ".pub", ".crt", ".key", ".der":
			return true
		default:
			return false
		}
	}

	normalize := func(filename string) string {
		base := strings.ToLower(strings.TrimSuffix(filename, filepath.Ext(filename)))
		// strip common suffixes
		for _, s := range []string{"_public", "-public", "_pub", "-pub", "_private", "-private", "_priv", "-priv"} {
			if strings.HasSuffix(base, s) {
				return strings.TrimSuffix(base, s)
			}
		}
		return base
	}

	// walk public keys
	_ = filepath.WalkDir(cfg.PublicKeysPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		if !isKeyFile(d.Name()) {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		rest := data
		for {
			var block *pem.Block
			block, rest = pem.Decode(rest)
			if block == nil {
				break
			}
			// try certificate
			if block.Type == "CERTIFICATE" {
				cert, err := x509.ParseCertificate(block.Bytes)
				if err == nil {
					if pk, ok := cert.PublicKey.(*rsa.PublicKey); ok {
						name := normalize(d.Name())
						pubMap[name] = pk
						break
					}
				}
				continue
			}
			// try PKIX
			if pk, err := x509.ParsePKIXPublicKey(block.Bytes); err == nil {
				if rsaPub, ok := pk.(*rsa.PublicKey); ok {
					name := normalize(d.Name())
					pubMap[name] = rsaPub
					break
				}
			}
			// try PKCS1
			if rsaPub, err := x509.ParsePKCS1PublicKey(block.Bytes); err == nil {
				name := normalize(d.Name())
				pubMap[name] = rsaPub
				break
			}
		}
		return nil
	})

	// walk private keys
	_ = filepath.WalkDir(cfg.PrivateKeysPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		if !isKeyFile(d.Name()) {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		rest := data
		for {
			var block *pem.Block
			block, rest = pem.Decode(rest)
			if block == nil {
				break
			}
			// try PKCS8
			if block.Type == "PRIVATE KEY" || block.Type == "ENCRYPTED PRIVATE KEY" {
				if keyIface, err := x509.ParsePKCS8PrivateKey(block.Bytes); err == nil {
					if rsaPriv, ok := keyIface.(*rsa.PrivateKey); ok {
						name := normalize(d.Name())
						privMap[name] = rsaPriv
						break
					}
				}
			}
			// try PKCS1
			if rsaPriv, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
				name := normalize(d.Name())
				privMap[name] = rsaPriv
				break
			}
		}
		return nil
	})

	var results []KeyPairConfig
	for name, pub := range pubMap {
		if priv, ok := privMap[name]; ok {
			results = append(results, KeyPairConfig{
				PublicKey:  *pub,
				PrivateKey: *priv,
			})
		}
	}

	return results, nil
}
