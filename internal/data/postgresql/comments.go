package postgresql

import (
	"database/sql"
	"fmt"
	"github.com/kientink26/go-json-api/internal/data/dto"
	"strings"
)

type CommentModel struct {
	DB *sql.DB
}

func (m CommentModel) Insert(comment *dto.Comment, userID int64, movieID int64) error {
	query := `INSERT INTO comments (body, user_id, movie_id)
			VALUES ($1, $2, $3)
			RETURNING id, created_at`
	args := []interface{}{comment.Body, userID, movieID}
	err := m.DB.QueryRow(query, args...).Scan(&comment.ID, &comment.CreatedAt)
	if err != nil {
		switch {
		case strings.Contains(err.Error(), `violates foreign key constraint "comments_movie_id_fkey"`):
			return ErrRecordNotFound
		case strings.Contains(err.Error(), `violates foreign key constraint "comments_user_id_fkey"`):
			panic(err)
		default:
			return err
		}
	}
	return nil
}

func (m CommentModel) GetAllForMovie(movieID int64, filters dto.Filters) ([]*dto.CommentUser, dto.Metadata, error) {
	query := fmt.Sprintf(`SELECT count(*) OVER(), comments.id, comments.created_at, comments.body, 
       									users.id, users.name, users.email, users.activated, users.created_at
			FROM comments
			INNER JOIN movies ON comments.movie_id = movies.id
			INNER JOIN users ON comments.user_id = users.id
			WHERE movies.id = $1
			ORDER BY comments.%s %s, comments.id ASC`, filters.SortColumn(), filters.SortDirection())
	rows, err := m.DB.Query(query, movieID)
	if err != nil {
		return nil, dto.Metadata{}, err
	}
	defer rows.Close()
	totalRecords := 0
	comments := []*dto.CommentUser{}
	// Use rows.Next to iterate through the rows in the resultset.
	for rows.Next() {
		var comment dto.CommentUser
		err := rows.Scan(
			&totalRecords, // Scan the count from the window function into totalRecords.
			&comment.ID,
			&comment.CreatedAt,
			&comment.Body,
			&comment.User.ID,
			&comment.User.Name,
			&comment.User.Email,
			&comment.User.Activated,
			&comment.User.CreatedAt,
		)
		if err != nil {
			return nil, dto.Metadata{}, err
		}
		comments = append(comments, &comment)
	}
	// When the rows.Next() loop has finished, call rows.Err() to retrieve any error
	// that was encountered during the iteration.
	if err = rows.Err(); err != nil {
		return nil, dto.Metadata{}, err
	}
	return comments, dto.Metadata{TotalRecords: totalRecords}, nil
}
