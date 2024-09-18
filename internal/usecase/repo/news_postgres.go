package repo

import (
	"context"
	sql2 "database/sql"
	"encoding/json"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"tarkib.uz/internal/entity"
	"tarkib.uz/pkg/postgres"
)

type NewsRepo struct {
	*postgres.Postgres
	*mongo.Client
}

func NewNewsRepo(pg *postgres.Postgres, client *mongo.Client) *NewsRepo {
	return &NewsRepo{pg, client}
}
func (n *NewsRepo) CreateNews(ctx context.Context, newsUz *entity.News, newsRu *entity.News) error {
	// Create the first news entry for Uzbek (UZ)
	newsIDUz := uuid.NewString()
	linksJSONUz, err := json.Marshal(newsUz.Links)
	if err != nil {
		return err
	}
	specialId := uuid.NewString()

	dataUz := map[string]interface{}{
		"id":              newsIDUz,
		"name":            newsUz.Name,
		"description":     newsUz.Description,
		"image_url":       newsUz.ImageURL,
		"video_url":       newsUz.VideoURL, // Added video_url field
		"text":            newsUz.Text,     // Added text field
		"links":           linksJSONUz,
		"language":        "uz",
		"site_image_link": newsUz.SiteImageLink,
		"voice_url":       newsUz.VoiceURL,
		"created_at":      time.Now(),
		"special_id":      specialId,
	}


	for _, v := range newsUz.SubCategoryIDs {
		dataUz = map[string]interface{}{
			"subcategory_id": v,
			"news_id":        newsIDUz,
		}

		sqlUz, argsUz, err = n.Builder.Insert("subcategory_news").
			SetMap(dataUz).ToSql()
		if err != nil {
			return err
		}

		if _, err = n.Pool.Exec(ctx, sqlUz, argsUz...); err != nil {
			return err
		}
	}

	// Create the second news entry for Russian (RU)
	newsIDRu := uuid.NewString()
	linksJSONRu, err := json.Marshal(newsRu.Links)
	if err != nil {
		return err
	}

	dataRu := map[string]interface{}{
		"id":              newsIDRu,
		"name":            newsRu.Name,
		"description":     newsRu.Description,
		"image_url":       newsRu.ImageURL,
		"video_url":       newsRu.VideoURL, // Added video_url field
		"text":            newsRu.Text,     // Added text field
		"links":           linksJSONRu,
		"language":        "ru",
		"site_image_link": newsRu.SiteImageLink,
		"voice_url":       newsRu.VoiceURL,
		"created_at":      time.Now(),
		"special_id":      specialId,
	}
	sqlRu, argsRu, err := n.Builder.Insert("news").
		SetMap(dataRu).ToSql()
	if err != nil {
		return err
	}

	if _, err = n.Pool.Exec(ctx, sqlRu, argsRu...); err != nil {
		return err
	}

	for _, v := range newsRu.SubCategoryIDs {
		dataRu = map[string]interface{}{
			"subcategory_id": v,
			"news_id":        newsIDRu,
		}

		sqlRu, argsRu, err = n.Builder.Insert("subcategory_news").
			SetMap(dataRu).ToSql()
		if err != nil {
			return err
		}

		if _, err = n.Pool.Exec(ctx, sqlRu, argsRu...); err != nil {
			return err
		}
	}

	return nil
}

func (n *NewsRepo) DeleteNews(ctx context.Context, id string) error {
	deleteSubcategoryNewsSQL, args, err := n.Builder.Delete("subcategory_news").
		Where(squirrel.Eq{
			"news_id": id,
		}).ToSql()
	if err != nil {
		return err
	}
	if _, err = n.Pool.Exec(ctx, deleteSubcategoryNewsSQL, args...); err != nil {
		return err
	}

	deleteNewsUzSQL, args, err := n.Builder.Delete("news").
		Where(squirrel.Eq{
			"id": id,
		}).ToSql()
	if err != nil {
		return err
	}
	if _, err = n.Pool.Exec(ctx, deleteNewsUzSQL, args...); err != nil {
		return err
	}

	return nil
}

func (n *NewsRepo) GetAllNews(ctx context.Context, request *entity.GetAllNewsRequest, language string) ([]entity.News, error) {
	var (
		newsList []entity.News
		ids      []string
	)
	offset := (request.Page - 1) * request.Limit

	sql, args, err := n.Builder.Select("*").
		From("news").
		Where(squirrel.Eq{"language": language}).
		OrderBy("created_at DESC").
		Limit(uint64(request.Limit)).
		Offset(uint64(offset)).
		ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := n.Pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var news entity.News
		var linksJSON []byte
		var video sql2.NullString

		if err := rows.Scan(&news.ID, &news.Name, &news.Description, &news.ImageURL, &news.SiteImageLink, &news.VoiceURL, &news.CreatedAt, &linksJSON, &news.Language, &video, &news.Text, &news.SpecialID); err != nil {
			return nil, err
		}

		if video.Valid {
			news.VideoURL = video.String
		}

		var links []entity.Link
		if err := json.Unmarshal(linksJSON, &links); err != nil {
			return nil, err
		}
		news.Links = links

		subCategoryIDsSQL, subCategoryIDsArgs, err := n.Builder.Select("subcategory_id").
			From("subcategory_news").
			Where(squirrel.Eq{"news_id": news.ID}).
			ToSql()
		if err != nil {
			return nil, err
		}

		subCategoryRows, err := n.Pool.Query(ctx, subCategoryIDsSQL, subCategoryIDsArgs...)
		if err != nil {
			return nil, err
		}

		for subCategoryRows.Next() {
			var id string
			if err = subCategoryRows.Scan(&id); err != nil {
				return nil, err
			}

			ids = append(ids, id)
		}
		subCategoryRows.Close()

		news.SubCategoryIDs = ids
		newsList = append(newsList, news)
		ids = nil
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return newsList, nil
}

func (n *NewsRepo) GetFilteredNews(ctx context.Context, request *entity.GetFilteredNewsRequest, language string) ([]entity.News, error) {
	var newsList []entity.News

	// Step 1: Fetch Subcategory IDs from MongoDB subcategories collection using CategoryID
	if request.CategoryID != "" && len(request.SubCategoryIDs) == 0 {
		// MongoDB query to fetch subcategories for the given category ID
		filter := bson.M{"category_id": request.CategoryID} // Use category_id field in subcategories collection

		// Fetch from MongoDB subcategories collection
		cursor, err := n.Client.Database("news").Collection("subcategories").Find(ctx, filter)
		if err != nil {
			return nil, err
		}
		defer cursor.Close(ctx)

		for cursor.Next(ctx) {
			var subCategory struct {
				ID string `bson:"id"`
			}
			if err := cursor.Decode(&subCategory); err != nil {
				return nil, err
			}
			request.SubCategoryIDs = append(request.SubCategoryIDs, subCategory.ID)
		}

		if err := cursor.Err(); err != nil {
			return nil, err
		}

		// If no subcategory IDs are found, return empty news list (strict category filtering)
		if len(request.SubCategoryIDs) == 0 {
			return []entity.News{}, nil
		}
	}

	// Step 2: Build the PostgreSQL query
	queryBuilder := n.Builder.Select("DISTINCT n.id, n.name, n.description, n.image_url, n.created_at, n.links, n.site_image_link, n.voice_url, n.video_url, n.text, n.special_id").
		From("news n").
		Where(squirrel.Eq{"n.language": language})

	// Step 3: Filter by SubCategoryIDs if they are present
	if len(request.SubCategoryIDs) > 0 {
		queryBuilder = queryBuilder.
			Join("subcategory_news sn ON n.id = sn.news_id").
			Where(squirrel.Eq{"sn.subcategory_id": request.SubCategoryIDs})
	}

	// Step 4: Handle SearchTerm if provided
	if request.SearchTerm != "" {
		searchTerm := "%" + request.SearchTerm + "%"
		queryBuilder = queryBuilder.
			Where(squirrel.Or{
				squirrel.ILike{"n.name": searchTerm},
				squirrel.ILike{"n.description": searchTerm},
			})
	}

	// Step 5: Pagination
	offset := (request.Page - 1) * request.Limit
	queryBuilder = queryBuilder.
		OrderBy("n.created_at DESC").
		Limit(uint64(request.Limit)).
		Offset(uint64(offset))

	// Execute SQL query
	sql, args, err := queryBuilder.ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := n.Pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Step 6: Scan the results
	for rows.Next() {
		var news entity.News
		var linksJSON []byte

		if err := rows.Scan(&news.ID, &news.Name, &news.Description, &news.ImageURL, &news.CreatedAt, &linksJSON, &news.SiteImageLink, &news.VoiceURL, &news.VideoURL, &news.Text, &news.SpecialID); err != nil {
			return nil, err
		}

		var links []entity.Link
		if err := json.Unmarshal(linksJSON, &links); err != nil {
			return nil, err
		}
		news.Links = links

		// Fetch subcategory IDs for each news item from news_subcategory table
		subCategoryIDsSQL, subCategoryIDsArgs, err := n.Builder.Select("subcategory_id").
			From("subcategory_news").
			Where(squirrel.Eq{"news_id": news.ID}).
			ToSql()
		if err != nil {
			return nil, err
		}

		subCategoryRows, err := n.Pool.Query(ctx, subCategoryIDsSQL, subCategoryIDsArgs...)
		if err != nil {
			return nil, err
		}

		var subCategoryIDs []string
		for subCategoryRows.Next() {
			var subCategoryID string
			if err := subCategoryRows.Scan(&subCategoryID); err != nil {
				return nil, err
			}
			subCategoryIDs = append(subCategoryIDs, subCategoryID)
		}
		subCategoryRows.Close()

		// Assign the fetched subcategory IDs to the news
		news.SubCategoryIDs = subCategoryIDs

		// Add to result list
		newsList = append(newsList, news)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return newsList, nil
}

func (n *NewsRepo) UpdateNews(ctx context.Context, id string, newsUz *entity.News) error {
	linksJSONUz, err := json.Marshal(newsUz.Links)
	if err != nil {
		return err
	}

	dataUz := map[string]interface{}{
		"name":            newsUz.Name,
		"description":     newsUz.Description,
		"image_url":       newsUz.ImageURL,
		"links":           linksJSONUz,
		"site_image_link": newsUz.SiteImageLink,
		"voice_url":       newsUz.VoiceURL,
		"video_url":       newsUz.VideoURL,
		"text":            newsUz.Text,
	}

	sqlUz, argsUz, err := n.Builder.Update("news").
		SetMap(dataUz).
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return err
	}

	if _, err = n.Pool.Exec(ctx, sqlUz, argsUz...); err != nil {
		return err
	}

	deleteSubcategorySQL, deleteArgs, err := n.Builder.Delete("subcategory_news").
		Where(squirrel.Eq{"news_id": id}).ToSql()
	if err != nil {
		return err
	}

	if _, err = n.Pool.Exec(ctx, deleteSubcategorySQL, deleteArgs...); err != nil {
		return err
	}

	for _, v := range newsUz.SubCategoryIDs {
		dataUz := map[string]interface{}{
			"subcategory_id": v,
			"news_id":        id,
		}

		insertSubcategorySQL, insertArgs, err := n.Builder.Insert("subcategory_news").
			SetMap(dataUz).ToSql()
		if err != nil {
			return err
		}

		if _, err = n.Pool.Exec(ctx, insertSubcategorySQL, insertArgs...); err != nil {
			return err
		}
	}

	return nil
}

func (n *NewsRepo) GetNewsByID(ctx context.Context, id string) (*entity.News, error) {
	var news entity.News
	var linksJSON []byte

	sql, args, err := n.Builder.Select("*").
		From("news").
		Where(squirrel.Eq{"id": id}).
		ToSql()
	if err != nil {
		return nil, err
	}

	row := n.Pool.QueryRow(ctx, sql, args...)
	if err := row.Scan(&news.ID, &news.Name, &news.Description, &news.ImageURL, &news.SiteImageLink, &news.VoiceURL, &news.CreatedAt, &linksJSON, &news.Language, &news.VideoURL, &news.Text, &news.SpecialID); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(linksJSON, &news.Links); err != nil {
		return nil, err
	}

	subCategoryIDsSQL, subCategoryIDsArgs, err := n.Builder.Select("subcategory_id").
		From("subcategory_news").
		Where(squirrel.Eq{"news_id": news.ID}).
		ToSql()
	if err != nil {
		return nil, err
	}

	subCategoryRows, err := n.Pool.Query(ctx, subCategoryIDsSQL, subCategoryIDsArgs...)
	if err != nil {
		return nil, err
	}
	defer subCategoryRows.Close()

	for subCategoryRows.Next() {
		var id string
		if err = subCategoryRows.Scan(&id); err != nil {
			return nil, err
		}
		news.SubCategoryIDs = append(news.SubCategoryIDs, id)
	}

	return &news, nil
}
