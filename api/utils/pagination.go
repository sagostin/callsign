package utils

import (
	"github.com/kataras/iris/v12"
	"gorm.io/gorm"
)

// PaginationParams holds pagination query parameters
type PaginationParams struct {
	Page  int    `url:"page"`
	Limit int    `url:"limit"`
	Sort  string `url:"sort"`
	Order string `url:"order"`
}

// DefaultPagination returns default pagination values
func DefaultPagination() PaginationParams {
	return PaginationParams{
		Page:  1,
		Limit: 20,
		Sort:  "id",
		Order: "asc",
	}
}

// GetPagination extracts pagination parameters from the request
func GetPagination(ctx iris.Context) PaginationParams {
	p := DefaultPagination()

	if page := ctx.URLParamIntDefault("page", 1); page > 0 {
		p.Page = page
	}

	if limit := ctx.URLParamIntDefault("limit", 20); limit > 0 && limit <= 100 {
		p.Limit = limit
	}

	if sort := ctx.URLParam("sort"); sort != "" {
		p.Sort = sort
	}

	if order := ctx.URLParam("order"); order == "desc" || order == "asc" {
		p.Order = order
	}

	return p
}

// Offset calculates the offset for database queries
func (p *PaginationParams) Offset() int {
	return (p.Page - 1) * p.Limit
}

// Apply applies pagination to a GORM query
func (p *PaginationParams) Apply(db *gorm.DB) *gorm.DB {
	return db.
		Offset(p.Offset()).
		Limit(p.Limit).
		Order(p.Sort + " " + p.Order)
}

// ApplyWithTotal applies pagination and returns total count
func (p *PaginationParams) ApplyWithTotal(db *gorm.DB, model interface{}) (*gorm.DB, int64) {
	var total int64
	db.Model(model).Count(&total)

	return p.Apply(db), total
}
