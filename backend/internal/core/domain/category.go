package domain

import (
	"fmt"
	"strings"
)

type Category struct {
	ID          int
	Name        string
	Description string
}

func NewCategory(name, description string) (*Category, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, fmt.Errorf("category name must not be empty")
	}
	return &Category{Name: name, Description: description}, nil
}
