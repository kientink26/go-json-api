package mock

import (
	"github.com/kientink26/go-json-api/internal/data/dto"
	"github.com/kientink26/go-json-api/internal/data/postgresql"
	"time"
)

var mockMovie = &dto.Movie{
	ID:        1,
	CreatedAt: time.Now(),
	Title:     "Black Panther",
	Year:      2018,
	Runtime:   134,
	Genres:    []string{"sci-fi", "action", "adventure"},
	Version:   1,
}

type MovieModel struct{}

func (m MovieModel) GetAll(title string, genres []string, filters dto.Filters) ([]*dto.Movie, dto.Metadata, error) {
	return []*dto.Movie{mockMovie}, dto.Metadata{}, nil
}

func (m MovieModel) Insert(movie *dto.Movie) error {
	return nil
}
func (m MovieModel) Get(id int64) (*dto.Movie, error) {
	switch id {
	case 1:
		return mockMovie, nil
	default:
		return nil, postgresql.ErrRecordNotFound
	}

}
func (m MovieModel) Update(movie *dto.Movie) error {
	return nil
}
func (m MovieModel) Delete(id int64) error {
	return nil
}
