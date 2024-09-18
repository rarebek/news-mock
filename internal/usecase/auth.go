package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/go-redis/redis/v8"
	"tarkib.uz/config"
	"tarkib.uz/internal/entity"
	tokens "tarkib.uz/pkg/token"
)

type AuthUseCase struct {
	repo        AuthRepo
	cfg         *config.Config
	RedisClient *redis.Client
}

func NewAuthUseCase(r AuthRepo, cfg *config.Config) *AuthUseCase {
	return &AuthUseCase{
		repo: r,
		cfg:  cfg,
	}
}

func (uc *AuthUseCase) Login(ctx context.Context, request *entity.Admin) (*entity.AdminLoginResponse, error) {
	admin, err := uc.repo.GetAdminData(ctx, request.Username)
	if err != nil {
		return nil, err
	}

	if admin.Password != request.Password {
		return nil, errors.New("xato parol kiritdingiz")
	}

	expDuration := time.Now().Add(time.Duration(uc.cfg.Casbin.AccessTokenTimeOut) * time.Second)
	expTime := expDuration.Format(time.RFC3339)

	jwtHandler := tokens.JWTHandler{
		Sub:       admin.Id,
		Iss:       time.Now().String(),
		Exp:       expTime,
		Role:      "admin",
		SigninKey: uc.cfg.Casbin.SigningKey,
		Timeout:   uc.cfg.Casbin.AccessTokenTimeOut,
	}

	accessToken, _, err := jwtHandler.GenerateAuthJWT()
	if err != nil {
		return nil, err
	}

	return &entity.AdminLoginResponse{
		Id:          admin.Id,
		Username:    admin.Username,
		Password:    admin.Password,
		Avatar:      admin.Avatar,
		AccessToken: accessToken,
	}, nil
}

func (uc *AuthUseCase) SuperAdminLogin(ctx context.Context, request *entity.SuperAdmin) (*entity.SuperAdminLoginResponse, error) {
	admin, err := uc.repo.GetSuperAdminData(ctx, request.PhoneNumber)
	if err != nil {
		return nil, err
	}

	if admin.Password != request.Password {
		return nil, errors.New("xato parol kiritdingiz")
	}

	expDuration := time.Now().Add(time.Duration(uc.cfg.Casbin.AccessTokenTimeOut) * time.Second)
	expTime := expDuration.Format(time.RFC3339)

	jwtHandler := tokens.JWTHandler{
		Sub:       admin.Id,
		Iss:       time.Now().Format(time.RFC3339),
		Exp:       expTime,
		Role:      "super-admin",
		SigninKey: uc.cfg.Casbin.SigningKey,
		Timeout:   uc.cfg.Casbin.AccessTokenTimeOut,
	}

	accessToken, _, err := jwtHandler.GenerateAuthJWT()
	if err != nil {
		return nil, err
	}

	return &entity.SuperAdminLoginResponse{
		Id:          admin.Id,
		PhoneNumber: admin.PhoneNumber,
		Password:    admin.Password,
		Avatar:      admin.Avatar,
		AccessToken: accessToken,
		IsBlocked:   admin.IsBlocked,
	}, nil
}

func (uc *AuthUseCase) CreateAdmin(ctx context.Context, admin *entity.Admin) error {
	if err := uc.repo.CreateAdmin(ctx, admin); err != nil {
		return err
	}

	return nil
}

func (uc *AuthUseCase) DeleteAdmin(ctx context.Context, id string) error {
	if err := uc.repo.DeleteAdmin(ctx, id); err != nil {
		return err
	}

	return nil
}

func (uc *AuthUseCase) GetAllAdmins(ctx context.Context) ([]entity.Admin, error) {
	admins, err := uc.repo.GetAllAdmins(ctx)
	if err != nil {
		return nil, err
	}

	return admins, nil
}

func (uc *AuthUseCase) EditAdmin(ctx context.Context, admin *entity.Admin) error {
	if err := uc.repo.EditAdmin(ctx, admin); err != nil {
		return err
	}

	return nil
}

func (uc *AuthUseCase) GetAdminById(ctx context.Context, id string) (*entity.Admin, error) {
	admin, err := uc.repo.GetAdminById(ctx, id)
	if err != nil {
		return nil, err
	}

	return admin, nil
}

func (uc *AuthUseCase) ChangeSuperAdminData(ctx context.Context, superadmin *entity.SuperAdmin) error {
	return uc.repo.ChangeSuperAdminData(ctx, superadmin)
}

func (uc *AuthUseCase) BlockSuperAdmin(ctx context.Context) error {
	return uc.repo.BlockSuperAdmin(ctx)
}
