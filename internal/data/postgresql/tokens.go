package postgresql

import (
	"database/sql"
	"github.com/kientink26/go-json-api/internal/data/dto"
	"time"
)

// Define the TokenModel type.
type TokenModel struct {
	DB *sql.DB
}

// The New() method is a shortcut which creates a new Token struct and then inserts the
// data in the tokens table.
func (m TokenModel) New(userID int64, ttl time.Duration, scope string) (*dto.Token, error) {
	token, err := dto.GenerateToken(userID, ttl, scope)
	if err != nil {
		return nil, err
	}
	err = m.Insert(token)
	return token, err
}

// Insert() adds the data for a specific token to the tokens table.
func (m TokenModel) Insert(token *dto.Token) error {
	query := `
INSERT INTO tokens (hash, user_id, expiry, scope)
VALUES ($1, $2, $3, $4)`
	args := []interface{}{token.Hash, token.UserID, token.Expiry, token.Scope}
	_, err := m.DB.Exec(query, args...)
	return err
}

// DeleteAllForUser() deletes all tokens for a specific user and scope.
func (m TokenModel) DeleteAllForUser(scope string, userID int64) error {
	query := `
DELETE FROM tokens
WHERE scope = $1 AND user_id = $2`
	_, err := m.DB.Exec(query, scope, userID)
	return err
}
