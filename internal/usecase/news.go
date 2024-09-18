package usecase

import (
	"context"

	"github.com/go-redis/redis/v8"
	"tarkib.uz/config"
	"tarkib.uz/internal/entity"
)

type NewsUseCase struct {
	repo        NewsRepo
	cfg         *config.Config
	RedisClient *redis.Client
}

func NewNewsUseCase(r NewsRepo, cfg *config.Config) *NewsUseCase {
	return &NewsUseCase{
		repo: r,
		cfg:  cfg,
	}
}

func (n *NewsUseCase) CreateNews(ctx context.Context, newsUz, newsRu *entity.News) error {
	return n.repo.CreateNews(ctx, newsUz, newsRu)
}

func (n *NewsUseCase) GetAllNews(ctx context.Context, request *entity.GetAllNewsRequest, language string) ([]entity.News, error) {
	news, err := n.repo.GetAllNews(ctx, request, language)
	if err != nil {
		return nil, err
	}

	return news, nil
}

func (n *NewsUseCase) DeleteNews(ctx context.Context, id string) error {
	return n.repo.DeleteNews(ctx, id)
}

func (n *NewsUseCase) GetFilteredNews(ctx context.Context, request *entity.GetFilteredNewsRequest, language string) ([]entity.News, error) {
	news, err := n.repo.GetFilteredNews(ctx, request, language)
	if err != nil {
		return nil, err
	}

	return news, nil
}

func (n *NewsUseCase) UpdateNews(ctx context.Context, id string, request *entity.News) error {
	return n.repo.UpdateNews(ctx, id, request)
}

func (n *NewsUseCase) GetNewsByID(ctx context.Context, id string) (*entity.News, error) {
	news, err := n.repo.GetNewsByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return news, nil
}
