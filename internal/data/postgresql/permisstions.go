package postgresql

import (
	"database/sql"
	"errors"
	"github.com/kientink26/go-json-api/internal/data/dto"
	"github.com/lib/pq"
	"strings"
)

var (
	ErrDuplicatePermission = errors.New("permission already have")
)

type PermissionModel struct {
	DB *sql.DB
}

// The GetAllForUser() method returns all permission codes for a specific user in a
// Permissions slice
func (m PermissionModel) GetAllForUser(userID int64) (dto.Permissions, error) {
	query := `
SELECT permissions.code
FROM permissions
INNER JOIN users_permissions ON users_permissions.permission_id = permissions.id
INNER JOIN users ON users_permissions.user_id = users.id
WHERE users.id = $1`
	rows, err := m.DB.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	permissions := dto.Permissions{}
	for rows.Next() {
		var permission string
		err := rows.Scan(&permission)
		if err != nil {
			return nil, err
		}
		permissions = append(permissions, permission)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return permissions, nil
}

func (m PermissionModel) AddForUser(userID int64, codes ...string) error {
	query := `
INSERT INTO users_permissions
SELECT $1, permissions.id FROM permissions WHERE permissions.code = ANY($2)`
	_, err := m.DB.Exec(query, userID, pq.Array(codes))
	if err != nil {
		switch {
		case strings.Contains(err.Error(), `violates foreign key constraint "users_permissions_user_id_fkey"`):
			return ErrRecordNotFound
		case strings.Contains(err.Error(), `violates unique constraint "users_permissions_pkey"`):
			return ErrDuplicatePermission
		default:
			return err
		}
	}
	return nil
}

func (m PermissionModel) DeleteForUser(userID int64, codes ...string) error {
	// Calling the Begin() method on the connection pool creates a new sql.Tx
	// object, which represents the in-progress database transaction.
	tx, err := m.DB.Begin()
	if err != nil {
		return err
	}
	query := `
DELETE FROM users_permissions
WHERE user_id = $1
AND permission_id = ANY(SELECT permissions.id FROM permissions WHERE permissions.code = ANY($2))`

	//Call Exec() on the transaction
	result, err := tx.Exec(query, userID, pq.Array(codes))
	if err != nil {
		// If there is any error, we call the tx.Rollback() method on the
		// transaction. This will abort the transaction and no changes will be
		// made to the database.
		tx.Rollback()
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		tx.Rollback()
		return err
	}
	// If less rows were affected, we know that the users_permissions table didn't contain a record
	// we tried to delete it.
	if int(rowsAffected) < len(codes) {
		tx.Rollback()
		return ErrRecordNotFound
	}
	// If there are no errors, the statements in the transaction can be committed
	// to the database with the tx.Commit() method
	err = tx.Commit()
	return err
}
