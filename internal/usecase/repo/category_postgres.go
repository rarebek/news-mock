package repo

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/k0kubun/pp"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"tarkib.uz/internal/entity"
	"tarkib.uz/pkg/postgres"
)

type CategoryRepo struct {
	*postgres.Postgres
	database *mongo.Database
}

func NewCategoryRepo(pg *postgres.Postgres, database *mongo.Database) *CategoryRepo {
	return &CategoryRepo{pg, database}
}

func (n *CategoryRepo) getCollection(collectionName string) *mongo.Collection {
	return n.database.Collection(collectionName)
}

func (n *CategoryRepo) AppendCategory(ctx context.Context, category *entity.Category) error {
	collection := n.getCollection("categories")
	_, err := collection.InsertOne(ctx, bson.M{
		"id":      category.ID,
		"name_uz": category.NameUz,
		"name_ru": category.NameRu,
	})
	return err
}

func (n *CategoryRepo) UpdateCategory(ctx context.Context, categoryID string, nameUz string, nameRu string) error {
	collection := n.getCollection("categories")
	filter := bson.M{"id": categoryID}
	update := bson.M{
		"$set": bson.M{
			"name_uz": nameUz,
			"name_ru": nameRu,
		},
	}
	_, err := collection.UpdateOne(ctx, filter, update)
	return err
}

func (n *CategoryRepo) DeleteCategory(ctx context.Context, categoryID string) error {
	collection := n.getCollection("categories")
	_, err := collection.DeleteOne(ctx, bson.M{"id": categoryID})
	return err
}

func (n *CategoryRepo) GetAllCategories(ctx context.Context) ([]entity.Category, error) {
	collection := n.getCollection("categories")
	var categories []entity.Category
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	if err = cursor.All(ctx, &categories); err != nil {
		return nil, err
	}
	return categories, nil
}

func (n *CategoryRepo) AppendSubCategory(ctx context.Context, subcategory *entity.SubCategory) error {
	collection := n.getCollection("subcategories")
	_, err := collection.InsertOne(ctx, bson.M{
		"id":          uuid.NewString(),
		"category_id": subcategory.CategoryID,
		"name_uz":     subcategory.NameUz,
		"name_ru":     subcategory.NameRu,
	})
	return err
}

func (n *CategoryRepo) UpdateSubCategory(ctx context.Context, subcategoryID string, nameUz string, nameRu string) error {
	collection := n.getCollection("subcategories")
	filter := bson.M{"id": subcategoryID}
	update := bson.M{
		"$set": bson.M{
			"name_uz": nameUz,
			"name_ru": nameRu,
		},
	}
	_, err := collection.UpdateOne(ctx, filter, update)
	return err
}

func (n *CategoryRepo) DeleteSubCategory(ctx context.Context, subcategoryID string) error {
	collection := n.getCollection("subcategories")
	_, err := collection.DeleteOne(ctx, bson.M{"id": subcategoryID})
	return err
}

func (n *CategoryRepo) GetAllSubCategories(ctx context.Context, categoryID string) ([]entity.SubCategory, error) {
	collection := n.getCollection("subcategories")
	var subcategories []entity.SubCategory
	filter := bson.M{"category_id": categoryID}
	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	if err = cursor.All(ctx, &subcategories); err != nil {
		return nil, err
	}
	return subcategories, nil
}

func (n *CategoryRepo) GetAllCategoriesWithSubCategories(ctx context.Context, language string) ([]entity.CategoryWithSubCategories, error) {
	// Get the collections for categories and subcategories
	collectionCategories := n.getCollection("categories")
	collectionSubCategories := n.getCollection("subcategories")

	// Fetch all categories
	var categories []entity.Category
	categoryCursor, err := collectionCategories.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	if err = categoryCursor.All(ctx, &categories); err != nil {
		return nil, err
	}

	// Fetch all subcategories
	var allSubcategories []entity.SubCategory
	subcategoryCursor, err := collectionSubCategories.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	if err = subcategoryCursor.All(ctx, &allSubcategories); err != nil {
		return nil, err
	}

	// Create a map to associate category_id with subcategories
	subcategoryMap := make(map[string][]entity.SubCategory)
	for _, subcategory := range allSubcategories {
		// Store each subcategory under its respective category_id in the map
		subcategoryMap[subcategory.CategoryID] = append(subcategoryMap[subcategory.CategoryID], subcategory)
	}

	// Build the result
	var result []entity.CategoryWithSubCategories
	for _, category := range categories {
		// Get the subcategories associated with this category's ID
		subcategories, found := subcategoryMap[category.ID]
		if !found {
			// If no subcategories found for this category, initialize an empty slice
			subcategories = []entity.SubCategory{}
		}

		// Use the requested language to fetch the correct field
		var categoryName string
		if language == "uz" {
			categoryName = category.NameUz
		} else if language == "ru" {
			categoryName = category.NameRu
		} else {
			return nil, fmt.Errorf("unsupported language: %s", language)
		}

		// For each subcategory, select the correct language field
		var subcategoryList []entity.SubCategory
		for _, subcategory := range subcategories {
			var subcategoryName string
			if language == "uz" {
				subcategoryName = subcategory.NameUz
			} else if language == "ru" {
				subcategoryName = subcategory.NameRu
			} else {
				return nil, fmt.Errorf("unsupported language: %s", language)
			}
			// Append the subcategory with the selected name to the list
			subcategoryList = append(subcategoryList, entity.SubCategory{
				ID:         subcategory.ID,
				CategoryID: subcategory.CategoryID,
				NameUz:     subcategory.NameUz, // Store both names for internal reference
				NameRu:     subcategory.NameRu, // Store both names for internal reference
				Name:       subcategoryName,    // Set the selected name
			})
		}

		// Append the category along with its associated subcategories to the result
		result = append(result, entity.CategoryWithSubCategories{
			ID:            category.ID,
			NameUz:        category.NameUz, // Store both names for internal reference
			NameRu:        category.NameRu, // Store both names for internal reference
			Name:          categoryName,    // Set the selected name
			SubCategories: subcategoryList,
		})
	}

	return result, nil
}

func (n *CategoryRepo) GetOneCategoryByID(ctx context.Context, id string) (*entity.CategoryWithSubCategories, error) {
	categoryCollection := n.getCollection("categories")

	var category entity.Category
	err := categoryCollection.FindOne(ctx, bson.M{"id": id}).Decode(&category)
	if err != nil {
		return nil, err
	}

	subcategoryCollection := n.getCollection("subcategories")
	var subcategories []entity.SubCategory
	cursor, err := subcategoryCollection.Find(ctx, bson.M{"category_id": id})
	if err != nil {
		return nil, err
	}

	if err = cursor.All(ctx, &subcategories); err != nil {
		return nil, err
	}

	pp.Println("SUBSSSSS", subcategories)

	result := &entity.CategoryWithSubCategories{
		ID:            category.ID,
		NameUz:        category.NameUz,
		NameRu:        category.NameRu,
		SubCategories: subcategories,
	}

	return result, nil
}

func (n *CategoryRepo) CreateSource(ctx context.Context, source *entity.Source) error {
	data := map[string]interface{}{
		"id":             uuid.NewString(),
		"site_name":      source.SiteName,
		"site_image_url": source.SiteImageURL,
		"site_url":       source.SiteURL,
	}

	sql, args, err := n.Builder.Insert("sources").
		SetMap(data).ToSql()
	if err != nil {
		return err
	}

	_, err = n.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return err
	}

	return nil
}

func (n *CategoryRepo) GetAllSources(ctx context.Context) ([]*entity.Source, error) {
	sql := "SELECT id, site_name, site_image_url, site_url FROM sources"

	rows, err := n.Pool.Query(ctx, sql)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sources []*entity.Source
	for rows.Next() {
		var source entity.Source
		if err := rows.Scan(&source.ID, &source.SiteName, &source.SiteImageURL, &source.SiteURL); err != nil {
			return nil, err
		}
		sources = append(sources, &source)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return sources, nil
}

func (n *CategoryRepo) DeleteSource(ctx context.Context, id string) error {
	sql := "DELETE FROM sources WHERE id = $1"

	_, err := n.Pool.Exec(ctx, sql, id)
	if err != nil {
		return err
	}

	return nil
}
