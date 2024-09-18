package usecase

import (
	"context"

	"github.com/go-redis/redis/v8"
	"tarkib.uz/config"
	"tarkib.uz/internal/entity"
)

type CategoryUseCase struct {
	repo        CategoryRepo
	cfg         *config.Config
	RedisClient *redis.Client
}

func NewCategoryUseCase(r CategoryRepo, cfg *config.Config) *CategoryUseCase {
	return &CategoryUseCase{
		repo: r,
		cfg:  cfg,
	}
}

func (n *CategoryUseCase) AppendCategory(ctx context.Context, category *entity.Category) error {
	return n.repo.AppendCategory(ctx, category)
}

func (n *CategoryUseCase) UpdateCategory(ctx context.Context, categoryID string, nameUz string, nameRu string) error {
	return n.repo.UpdateCategory(ctx, categoryID, nameUz, nameRu)
}

func (n *CategoryUseCase) DeleteCategory(ctx context.Context, categoryID string) error {
	return n.repo.DeleteCategory(ctx, categoryID)
}

func (n *CategoryUseCase) GetAllCategories(ctx context.Context) ([]entity.Category, error) {
	return n.repo.GetAllCategories(ctx)
}

func (n *CategoryUseCase) AppendSubCategory(ctx context.Context, subcategory *entity.SubCategory) error {
	return n.repo.AppendSubCategory(ctx, subcategory)
}

func (n *CategoryUseCase) UpdateSubCategory(ctx context.Context, subcategoryID string, nameUz string, nameRu string) error {
	return n.repo.UpdateSubCategory(ctx, subcategoryID, nameUz, nameRu)
}

func (n *CategoryUseCase) DeleteSubCategory(ctx context.Context, subcategoryID string) error {
	return n.repo.DeleteSubCategory(ctx, subcategoryID)
}

func (n *CategoryUseCase) GetAllSubCategories(ctx context.Context, categoryID string) ([]entity.SubCategory, error) {
	return n.repo.GetAllSubCategories(ctx, categoryID)
}

func (n *CategoryUseCase) CreateSource(ctx context.Context, s *entity.Source) error {
	return n.repo.CreateSource(ctx, s)
}

func (n *CategoryUseCase) DeleteSource(ctx context.Context, id string) error {
	return n.repo.DeleteSource(ctx, id)
}

func (n *CategoryUseCase) GetAllSources(ctx context.Context) ([]*entity.Source, error) {
	return n.repo.GetAllSources(ctx)
}

func (n *CategoryUseCase) GetAllCategoriesWithSubCategories(ctx context.Context, language string) ([]entity.CategoryWithSubCategories, error) {
	return n.repo.GetAllCategoriesWithSubCategories(ctx, language)
}

func (n *CategoryUseCase) GetOneCategoryByID(ctx context.Context, id string) (*entity.CategoryWithSubCategories, error) {
	return n.repo.GetOneCategoryByID(ctx, id)
}
