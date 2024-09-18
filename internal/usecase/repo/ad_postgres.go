package repo

import (
	"context"
	ssq "database/sql"
	"fmt"

	"github.com/Masterminds/squirrel"
	"tarkib.uz/internal/entity"
	"tarkib.uz/pkg/postgres"
)

type AdRepo struct {
	*postgres.Postgres
}

func NewAdRepo(pg *postgres.Postgres) *AdRepo {
	return &AdRepo{pg}
}

func (a *AdRepo) CreateAd(ctx context.Context, request *entity.Ad) error {
	data := map[string]interface{}{
		"id":         request.ID,
		"link":       request.Link,
		"image_url":  request.ImageURL,
		"created_at": request.CreatedAt,
	}
	sql, args, err := a.Builder.Insert("ads").
		SetMap(data).ToSql()
	if err != nil {
		return err
	}

	if _, err = a.Pool.Exec(ctx, sql, args...); err != nil {
		return err
	}

	return nil
}

func (a *AdRepo) DeleteAd(ctx context.Context, id string) error {
	query := "DELETE FROM ads WHERE id = $1"

	if _, err := a.Pool.Exec(ctx, query, id); err != nil {
		return err
	}

	return nil
}

func (a *AdRepo) UpdateAd(ctx context.Context, request *entity.Ad) error {
	data := map[string]interface{}{
		"link":      request.Link,
		"image_url": request.ImageURL,
	}
	sql, args, err := a.Builder.Update("ads").
		SetMap(data).Where(squirrel.Eq{
		"id": request.ID,
	}).ToSql()
	if err != nil {
		return err
	}

	if _, err = a.Pool.Exec(ctx, sql, args...); err != nil {
		return err
	}

	return nil
}

func (a *AdRepo) GetAd(ctx context.Context, request *entity.GetAdRequest) (*entity.Ad, error) {
	var ad entity.Ad

	if request.IsAdmin {
		var viewCount ssq.NullInt64
		query := a.Builder.Select("id, link, image_url, view_count").From("ads").Where(squirrel.Eq{
			"id": request.ID,
		})
		sql, args, err := query.ToSql()
		if err != nil {
			return nil, fmt.Errorf("failed to build SQL query: %w", err)
		}

		row := a.Pool.QueryRow(ctx, sql, args...)
		if err := row.Scan(&ad.ID, &ad.Link, &ad.ImageURL, &viewCount); err != nil {
			return nil, fmt.Errorf("failed to scan ad for admin: %w", err)
		}

		if viewCount.Valid {
			ad.ViewCount = int(viewCount.Int64)
		} else {
			ad.ViewCount = 0
		}

		return &ad, nil
	} else {
		selectQuery := a.Builder.Select("id, link, image_url, view_count").From("ads").Where(squirrel.Eq{
			"id": request.ID,
		})
		sql, args, err := selectQuery.ToSql()
		if err != nil {
			return nil, fmt.Errorf("failed to build SQL query: %w", err)
		}

		row := a.Pool.QueryRow(ctx, sql, args...)
		var viewCount ssq.NullInt64
		if err := row.Scan(&ad.ID, &ad.Link, &ad.ImageURL, &viewCount); err != nil {
			return nil, fmt.Errorf("failed to scan ad for non-admin: %w", err)
		}

		if viewCount.Valid {
			ad.ViewCount = int(viewCount.Int64)
		} else {
			ad.ViewCount = 0
		}

		ad.ViewCount += 1

		updateQuery := "UPDATE ads SET view_count = $1 WHERE id = $2"
		_, err = a.Pool.Exec(ctx, updateQuery, ad.ViewCount, ad.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to execute update query: %w", err)
		}

		ad.ViewCount = 0

		return &ad, nil
	}

}

func (a *AdRepo) GetAllAds(ctx context.Context) ([]*entity.Ad, error) {
	var ads []*entity.Ad

	query := a.Builder.Select("id, link, image_url, view_count").From("ads").OrderBy("created_at DESC")
	sql, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build SQL query: %w", err)
	}

	rows, err := a.Pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var ad entity.Ad
		var viewCount ssq.NullInt64
		if err := rows.Scan(&ad.ID, &ad.Link, &ad.ImageURL, &viewCount); err != nil {
			return nil, fmt.Errorf("failed to scan ad: %w", err)
		}

		if viewCount.Valid {
			ad.ViewCount = int(viewCount.Int64)
		} else {
			ad.ViewCount = 0
		}

		ads = append(ads, &ad)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error occurred during rows iteration: %w", err)
	}

	return ads, nil
}
