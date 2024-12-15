package data

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base32"
	"time"

	"github.com/V4N1LLA-1CE/movie-db-api/internal/validator"
)

// define constants for token scope
const (
	ScopeActivation     = "activation"
	ScopeAuthentication = "authentication"
)

type Token struct {
	Plaintext string    `json:"token"`
	Hash      []byte    `json:"-"`
	UserID    int64     `json:"-"`
	Expiry    time.Time `json:"expiry"`
	Scope     string    `json:"-"`
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

func ValidateTokenPlaintext(v *validator.Validator, tokenPlaintext string) {
	v.Check(tokenPlaintext != "", "token", "must be provided")
	v.Check(len(tokenPlaintext) == 26, "token", "must be 26 bytes long")
}

type TokenModel struct {
	DB *sql.DB
}

func (m *TokenModel) New(userID int64, ttl time.Duration, scope string) (*Token, error) {
	token, err := generateToken(userID, ttl, scope)
	if err != nil {
		return nil, err
	}

	err = m.Insert(token)
	return token, err
}

func (m *TokenModel) Insert(token *Token) error {
	stmt := `INSERT INTO tokens (hash, user_id, expiry, scope)
  VALUES ($1, $2, $3, $4)`

	args := []any{token.Hash, token.UserID, token.Expiry, token.Scope}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, stmt, args...)
	return err
}

func (m *TokenModel) DeleteAllForUser(scope string, userID int64) error {
	stmt := `DELETE FROM tokens
  WHERE scope = $1 AND user_id = $2`

	args := []any{scope, userID}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := m.DB.ExecContext(ctx, stmt, args...)
	return err
}
