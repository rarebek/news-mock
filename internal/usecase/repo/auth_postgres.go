package repo

import (
	"context"
	"database/sql"

	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"tarkib.uz/internal/entity"
	"tarkib.uz/pkg/postgres"
)

type AuthRepo struct {
	*postgres.Postgres
}

func NewAuthRepo(pg *postgres.Postgres) *AuthRepo {
	return &AuthRepo{pg}
}

func (a *AuthRepo) GetAdminData(ctx context.Context, Username string) (*entity.Admin, error) {
	var adminPostgres entity.Admin
	var avatar sql.NullString
	sql, args, err := a.Builder.Select("id, username, password, avatar").
		From("admins").
		Where(squirrel.Eq{
			"username": Username,
		}).ToSql()
	if err != nil {
		return nil, err
	}

	row := a.Pool.QueryRow(ctx, sql, args...)

	if err = row.Scan(&adminPostgres.Id, &adminPostgres.Username, &adminPostgres.Password, &avatar); err != nil {
		return nil, err
	}

	if avatar.Valid {
		adminPostgres.Avatar = avatar.String
	}

	return &adminPostgres, nil
}

func (a *AuthRepo) GetAdminById(ctx context.Context, id string) (*entity.Admin, error) {
	var adminPostgres entity.Admin
	var avatar sql.NullString
	sql, args, err := a.Builder.Select("id, username, password, avatar").
		From("admins").
		Where(squirrel.Eq{
			"id": id,
		}).ToSql()
	if err != nil {
		return nil, err
	}

	row := a.Pool.QueryRow(ctx, sql, args...)

	if err = row.Scan(&adminPostgres.Id, &adminPostgres.Username, &adminPostgres.Password, &avatar); err != nil {
		return nil, err
	}

	if avatar.Valid {
		adminPostgres.Avatar = avatar.String
	}

	return &adminPostgres, nil
}

func (a *AuthRepo) GetSuperAdminData(ctx context.Context, PhoneNumber string) (*entity.SuperAdmin, error) {
	var adminPostgres entity.SuperAdmin
	var avatarNull sql.NullString
	sql, args, err := a.Builder.Select("id, phone_number, password, avatar, is_blocked").
		From("superadmins").
		Where(squirrel.Eq{
			"phone_number": PhoneNumber,
		}).ToSql()
	if err != nil {
		return nil, err
	}

	row := a.Pool.QueryRow(ctx, sql, args...)

	if err = row.Scan(&adminPostgres.Id, &adminPostgres.PhoneNumber, &adminPostgres.Password, &avatarNull, &adminPostgres.IsBlocked); err != nil {
		return nil, err
	}

	if avatarNull.Valid {
		adminPostgres.Avatar = avatarNull.String
	}

	return &adminPostgres, nil
}

func (a *AuthRepo) CreateAdmin(ctx context.Context, admin *entity.Admin) error {
	data := map[string]interface{}{
		"id":       uuid.NewString(),
		"username": admin.Username,
		"password": admin.Password,
		"avatar":   admin.Avatar,
	}

	sql, args, err := a.Builder.Insert("admins").
		SetMap(data).ToSql()
	if err != nil {
		return err
	}

	if _, err = a.Pool.Exec(ctx, sql, args...); err != nil {
		return err
	}

	return nil
}

func (a *AuthRepo) DeleteAdmin(ctx context.Context, id string) error {
	sql, args, err := a.Builder.Delete("admins").
		Where(squirrel.Eq{
			"id": id,
		}).ToSql()
	if err != nil {
		return err
	}

	if _, err = a.Pool.Exec(ctx, sql, args...); err != nil {
		return err
	}

	return nil
}

func (a *AuthRepo) GetAllAdmins(ctx context.Context) ([]entity.Admin, error) {
	var response []entity.Admin
	sql, args, err := a.Builder.Select("id, username, password, avatar").
		From("admins").ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := a.Pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var admin entity.Admin
		if err = rows.Scan(&admin.Id, &admin.Username, &admin.Password, &admin.Avatar); err != nil {
			return nil, err
		}

		response = append(response, admin)
	}

	return response, nil
}

func (a *AuthRepo) EditAdmin(ctx context.Context, admin *entity.Admin) error {
	data := map[string]interface{}{
		"username": admin.Username,
		"password": admin.Password,
		"avatar":   admin.Avatar,
	}

	sql, args, err := a.Builder.Update("admins").
		SetMap(data).
		Where(squirrel.Eq{
			"id": admin.Id,
		}).ToSql()
	if err != nil {
		return err
	}

	if _, err = a.Pool.Exec(ctx, sql, args...); err != nil {
		return err
	}

	return nil
}

func (a *AuthRepo) ChangeSuperAdminData(ctx context.Context, superAdmin *entity.SuperAdmin) error {
	data := map[string]interface{}{
		"phone_number": superAdmin.PhoneNumber,
		"password":     superAdmin.Password,
		"avatar":       superAdmin.Avatar,
	}
	sql, args, err := a.Builder.Update("superadmins").
		SetMap(data).
		Where(squirrel.Eq{
			"id": superAdmin.Id,
		}).ToSql()
	if err != nil {
		return err
	}

	if _, err := a.Pool.Exec(ctx, sql, args...); err != nil {
		return err
	}

	return nil
}

func (a *AuthRepo) BlockSuperAdmin(ctx context.Context) error {
	data := make(map[string]interface{})
	data["is_blocked"] = true
	sql, args, err := a.Builder.Update("superadmins").SetMap(data).ToSql()
	if err != nil {
		return err
	}

	if _, err = a.Pool.Exec(ctx, sql, args...); err != nil {
		return err
	}

	return nil
}
