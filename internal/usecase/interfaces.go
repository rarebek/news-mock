// Package usecase implements application business logic. Each logic group in own file.
package usecase

import (
	"context"

	"tarkib.uz/internal/entity"
)

// //go:generate mockgen -source=interfaces.go -destination=./mocks_test.go -package=usecase_test
type (
	Auth interface {
		Login(ctx context.Context, admin *entity.Admin) (*entity.AdminLoginResponse, error)
		SuperAdminLogin(ctx context.Context, admin *entity.SuperAdmin) (*entity.SuperAdminLoginResponse, error)
		CreateAdmin(ctx context.Context, admin *entity.Admin) error
		DeleteAdmin(ctx context.Context, id string) error
		GetAllAdmins(ctx context.Context) ([]entity.Admin, error)
		EditAdmin(ctx context.Context, admin *entity.Admin) error
		GetAdminById(ctx context.Context, id string) (*entity.Admin, error)
		ChangeSuperAdminData(ctx context.Context, superAdmin *entity.SuperAdmin) error
		BlockSuperAdmin(ctx context.Context) error
	}

	AuthRepo interface {
		GetAdminData(ctx context.Context, Username string) (*entity.Admin, error)
		GetSuperAdminData(ctx context.Context, PhoneNumber string) (*entity.SuperAdmin, error)
		CreateAdmin(ctx context.Context, admin *entity.Admin) error
		DeleteAdmin(ctx context.Context, id string) error
		GetAllAdmins(ctx context.Context) ([]entity.Admin, error)
		EditAdmin(ctx context.Context, admin *entity.Admin) error
		GetAdminById(ctx context.Context, id string) (*entity.Admin, error)
		ChangeSuperAdminData(ctx context.Context, superAdmin *entity.SuperAdmin) error
		BlockSuperAdmin(ctx context.Context) error
	}

	Ad interface {
		CreateAd(ctx context.Context, ad *entity.Ad) error
		DeleteAd(ctx context.Context, id string) error
		UpdateAd(ctx context.Context, ad *entity.Ad) error
		GetAd(ctx context.Context, request *entity.GetAdRequest) (*entity.Ad, error)
		GetAllAds(ctx context.Context) ([]*entity.Ad, error)
	}

	AdRepo interface {
		CreateAd(ctx context.Context, ad *entity.Ad) error
		DeleteAd(ctx context.Context, id string) error
		UpdateAd(ctx context.Context, ad *entity.Ad) error
		GetAd(ctx context.Context, request *entity.GetAdRequest) (*entity.Ad, error)
		GetAllAds(ctx context.Context) ([]*entity.Ad, error)
	}

	News interface {
		CreateNews(ctx context.Context, newsUz, newsRu *entity.News) error
		GetAllNews(ctx context.Context, request *entity.GetAllNewsRequest, language string) ([]entity.News, error)
		DeleteNews(ctx context.Context, id string) error
		GetFilteredNews(ctx context.Context, request *entity.GetFilteredNewsRequest, language string) ([]entity.News, error)
		UpdateNews(ctx context.Context, id string, request *entity.News) error
		GetNewsByID(crx context.Context, id string) (*entity.News, error)
	}

	NewsRepo interface {
		CreateNews(ctx context.Context, newsUz, newsRu *entity.News) error
		GetAllNews(ctx context.Context, request *entity.GetAllNewsRequest, language string) ([]entity.News, error)
		DeleteNews(ctx context.Context, id string) error
		GetFilteredNews(ctx context.Context, request *entity.GetFilteredNewsRequest, language string) ([]entity.News, error)
		UpdateNews(ctx context.Context, id string, request *entity.News) error
		GetNewsByID(crx context.Context, id string) (*entity.News, error)
	}

	Category interface {
		AppendCategory(ctx context.Context, category *entity.Category) error
		UpdateCategory(ctx context.Context, categoryID string, nameUz string, nameRu string) error
		DeleteCategory(ctx context.Context, categoryID string) error
		GetAllCategories(ctx context.Context) ([]entity.Category, error)
		AppendSubCategory(ctx context.Context, subcategory *entity.SubCategory) error
		UpdateSubCategory(ctx context.Context, subcategoryID string, nameUz string, nameRu string) error
		DeleteSubCategory(ctx context.Context, subcategoryID string) error
		GetAllSubCategories(ctx context.Context, categoryID string) ([]entity.SubCategory, error)
		CreateSource(ctx context.Context, source *entity.Source) error
		GetAllSources(ctx context.Context) ([]*entity.Source, error)
		DeleteSource(ctx context.Context, id string) error
		GetAllCategoriesWithSubCategories(ctx context.Context, language string) ([]entity.CategoryWithSubCategories, error)
		GetOneCategoryByID(ctx context.Context, id string) (*entity.CategoryWithSubCategories, error)
	}

	CategoryRepo interface {
		AppendCategory(ctx context.Context, category *entity.Category) error
		UpdateCategory(ctx context.Context, categoryID string, nameUz string, nameRu string) error
		DeleteCategory(ctx context.Context, categoryID string) error
		GetAllCategories(ctx context.Context) ([]entity.Category, error)
		AppendSubCategory(ctx context.Context, subcategory *entity.SubCategory) error
		UpdateSubCategory(ctx context.Context, subcategoryID string, nameUz string, nameRu string) error
		DeleteSubCategory(ctx context.Context, subcategoryID string) error
		GetAllSubCategories(ctx context.Context, categoryID string) ([]entity.SubCategory, error)
		CreateSource(ctx context.Context, source *entity.Source) error
		GetAllSources(ctx context.Context) ([]*entity.Source, error)
		DeleteSource(ctx context.Context, id string) error
		GetAllCategoriesWithSubCategories(ctx context.Context, language string) ([]entity.CategoryWithSubCategories, error)
		GetOneCategoryByID(ctx context.Context, id string) (*entity.CategoryWithSubCategories, error)
	}
)
