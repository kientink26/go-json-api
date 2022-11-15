package postgresql

import (
	"crypto/sha256"
	"database/sql"
	"errors"
	"fmt"
	"github.com/kientink26/go-json-api/internal/data/dto"
	"time"
)

var (
	ErrDuplicateEmail = errors.New("duplicate email")
)

type UserModel struct {
	DB *sql.DB
}

func (m UserModel) GetAll(name string, email string, filters dto.Filters) ([]*dto.User, dto.Metadata, error) {
	// Construct the SQL query to retrieve all movie records.
	query := fmt.Sprintf(`
SELECT count(*) OVER(), id, created_at, name, email, password_hash, activated, version
FROM users
WHERE (to_tsvector('simple', name) @@ plainto_tsquery('simple', $1) OR $1 = '')
AND (STRPOS(LOWER(email), LOWER($2)) > 0 OR $2 = '')
ORDER BY %s %s, id ASC
LIMIT $3 OFFSET $4`, filters.SortColumn(), filters.SortDirection())

	rows, err := m.DB.Query(query, name, email, filters.Limit(), filters.Offset())
	if err != nil {
		return nil, dto.Metadata{}, err
	}
	// defer a call to rows.Close() to ensure that the resultset is closed
	// before GetAll() returns.
	defer rows.Close()

	totalRecords := 0
	users := []*dto.User{}
	for rows.Next() {
		var user dto.User
		err := rows.Scan(
			&totalRecords,
			&user.ID,
			&user.CreatedAt,
			&user.Name,
			&user.Email,
			&user.Password.Hash,
			&user.Activated,
			&user.Version,
		)
		if err != nil {
			return nil, dto.Metadata{}, err
		}
		users = append(users, &user)
	}
	// When the rows.Next() loop has finished, call rows.Err() to retrieve any error
	// that was encountered during the iteration.
	if err = rows.Err(); err != nil {
		return nil, dto.Metadata{}, err
	}
	metadata := dto.CalculateMetadata(totalRecords, filters.Page, filters.PageSize)
	return users, metadata, nil
}

func (m UserModel) Insert(user *dto.User) error {
	query := `
INSERT INTO users (name, email, password_hash, activated)
VALUES ($1, $2, $3, $4)
RETURNING id, created_at, version`
	args := []interface{}{user.Name, user.Email, user.Password.Hash, user.Activated}
	// If the table already contains a record with this email address, then when we try
	// to perform the insert there will be a violation of the UNIQUE "users_email_key"
	// constraint
	err := m.DB.QueryRow(query, args...).Scan(&user.ID, &user.CreatedAt, &user.Version)
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			return ErrDuplicateEmail
		default:
			return err
		}
	}
	return nil
}

func (m UserModel) GetByEmail(email string) (*dto.User, error) {
	query := `
SELECT id, created_at, name, email, password_hash, activated, version
FROM users
WHERE email = $1`
	var user dto.User
	err := m.DB.QueryRow(query, email).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.Name,
		&user.Email,
		&user.Password.Hash,
		&user.Activated,
		&user.Version,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	return &user, nil
}

func (m UserModel) Update(user *dto.User) error {
	query := `
UPDATE users
SET name = $1, email = $2, password_hash = $3, activated = $4, version = version + 1
WHERE id = $5 AND version = $6
RETURNING version`
	args := []interface{}{
		user.Name,
		user.Email,
		user.Password.Hash,
		user.Activated,
		user.ID,
		user.Version,
	}
	err := m.DB.QueryRow(query, args...).Scan(&user.Version)
	if err != nil {
		switch {
		case err.Error() == `pq: duplicate key value violates unique constraint "users_email_key"`:
			return ErrDuplicateEmail
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}
	return nil
}

func (m UserModel) GetForToken(tokenScope, tokenPlaintext string) (*dto.User, error) {
	// Calculate the SHA-256 hash of the plaintext token provided by the client.
	tokenHash := sha256.Sum256([]byte(tokenPlaintext))
	// Set up the SQL query.
	query := `
SELECT users.id, users.created_at, users.name, users.email, users.password_hash, users.activated, users.version
FROM users
INNER JOIN tokens
ON users.id = tokens.user_id
WHERE tokens.hash = $1
AND tokens.scope = $2
AND tokens.expiry > $3`
	args := []interface{}{tokenHash[:], tokenScope, time.Now()}
	var user dto.User
	err := m.DB.QueryRow(query, args...).Scan(
		&user.ID,
		&user.CreatedAt,
		&user.Name,
		&user.Email,
		&user.Password.Hash,
		&user.Activated,
		&user.Version,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}
	// Return the matching user.
	return &user, nil
}
