package data

import (
	"database/sql"
	"github.com/kientink26/go-json-api/internal/data/dto"
	"github.com/kientink26/go-json-api/internal/data/mock"
	"github.com/kientink26/go-json-api/internal/data/postgresql"
)

type Models struct {
	Movies interface {
		GetAll(title string, genres []string, filters dto.Filters) ([]*dto.Movie, dto.Metadata, error)
		Insert(movie *dto.Movie) error
		Get(id int64) (*dto.Movie, error)
		Update(movie *dto.Movie) error
		Delete(id int64) error
	}
	Users       postgresql.UserModel
	Tokens      postgresql.TokenModel
	Permissions postgresql.PermissionModel
	Comments    postgresql.CommentModel
}

func NewModels(db *sql.DB) Models {
	return Models{
		Movies:      postgresql.MovieModel{DB: db},
		Users:       postgresql.UserModel{DB: db},
		Tokens:      postgresql.TokenModel{DB: db},
		Permissions: postgresql.PermissionModel{DB: db},
		Comments:    postgresql.CommentModel{DB: db},
	}
}

func NewMockModels() Models {
	return Models{
		Movies: mock.MovieModel{},
	}
}
