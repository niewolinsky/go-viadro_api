package data

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	ScopeAuthentication = "authentication"
	ScopeActivation     = "activation"
)

type Token struct {
	Plaintext string    `json:"token"`
	Hash      []byte    `json:"-"`
	UserID    int64     `json:"-"`
	Expiry    time.Time `json:"expiry"`
	Scope     string    `json:"-"`
}

func generateToken(userID int64, ttl time.Duration, scope string) (*Token, error) {

	token := &Token{
		UserID: userID,
		Expiry: time.Now().Add(ttl),
		Scope:  scope,
	}

	randomBytes := make([]byte, 16)

	_, err := rand.Read(randomBytes)
	if err != nil {
		return nil, err
	}

	token.Plaintext = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(randomBytes)

	hash := sha256.Sum256([]byte(token.Plaintext))
	token.Hash = hash[:]
	return token, nil
}

type TokenLayer struct {
	DB *pgxpool.Pool
}

func (t TokenLayer) New(userID int64, ttl time.Duration, scope string) (*Token, error) {
	token, err := generateToken(userID, ttl, scope)
	if err != nil {
		return nil, err
	}

	t.Insert(token)
	return token, nil
}

func (t TokenLayer) Insert(token *Token) {
	query := `
	INSERT INTO tokens (hash, user_id, expiry, scope)
	VALUES ($1, $2, $3, $4)`
	args := []interface{}{token.Hash, token.UserID, token.Expiry, token.Scope}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_ = t.DB.QueryRow(ctx, query, args...)
}

func (t TokenLayer) DeleteAllForUser(scope string, userID int64) {
	query := `
	DELETE FROM tokens
	WHERE scope = $1 AND user_id = $2`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	_ = t.DB.QueryRow(ctx, query, scope, userID)
}
