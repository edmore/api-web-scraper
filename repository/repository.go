package repository

import "github.com/edmore/api-web-scraper/model"

type Repository interface {
	GetPageContents() (*model.Page, error)
}
