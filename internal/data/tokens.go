package data

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"time"
)

// define constants for token scope
const (
	ScopeActivation = "activation"
)

type Token struct {
	Plaintext string
	Hash      []byte
	UserID    int64
	Expiry    time.Time
	Scope     string
}

func generateToken(userID int64, ttl time.Duration, scope string) (*Token, error) {
	// create token that contains userid, expiry and scope
	// token expiry = now + ttl (time-to-live)
	token := &Token{
		UserID: userID,
		Expiry: time.Now().Add(ttl),
		Scope:  scope,
	}

	// initialise empty byte slice to store hash
	randomBytes := make([]byte, 16)

	// fill byte slice with random bytes from operating system's CSPRNG
	// this random byte will be encoded into plaintext and hashed into a hash
	_, err := rand.Read(randomBytes)
	if err != nil {
		return nil, err
	}

	// encode byte slice to base32-encoded string and assign to token plaintext
	// whilst remove the padding character '=' to make token cleaner and cause less issues along the line (i.e. URLs)
	token.Plaintext = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randomBytes)

	// generate sha256 hash of plaintext
	// Sum256() returns fixed-sized array [32]byte
	// hash[:] converts the hash array into slice []byte to match token hash type
	hash := sha256.Sum256([]byte(token.Plaintext))
	token.Hash = hash[:]

	return token, nil
}
