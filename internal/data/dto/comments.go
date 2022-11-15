package dto

import (
	"github.com/kientink26/go-json-api/internal/validator"
	"strings"
	"time"
	"unicode/utf8"
)

type Comment struct {
	ID        int64     `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Body      string    `json:"body"`
}

func ValidateComment(v *validator.Validator, comment *Comment) {
	v.Check(strings.TrimSpace(comment.Body) != "", "body", "must be provided")
	v.Check(utf8.RuneCountInString(comment.Body) <= 500, "body", "must not be more than 500 characters long")
}

type CommentUser struct {
	Comment `json:"comment"`
	User    User `json:"user"`
}
