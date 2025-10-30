package repository

import (
	"context"

	"go-esb/internal/models"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type ConnectionRepository interface {
	GetConnectionSettings(ctx context.Context, systemID uuid.UUID) (*models.ConnectionSetting, error)
	GetConnectionAuth(ctx context.Context, authID uuid.UUID) (*models.ConnectionAuthentication, error)
	CreateConnectionSetting(ctx context.Context, setting *models.ConnectionSetting) error
	CreateConnectionAuth(ctx context.Context, auth *models.ConnectionAuthentication) error
}

type connectionRepository struct {
	db *sqlx.DB
}

func NewConnectionRepository(db *sqlx.DB) ConnectionRepository {
	return &connectionRepository{db: db}
}

func (r *connectionRepository) GetConnectionSettings(ctx context.Context, systemID uuid.UUID) (*models.ConnectionSetting, error) {
	var setting models.ConnectionSetting
	err := r.db.GetContext(ctx, &setting, `
        SELECT ref, name, system, path, port, auth 
        FROM connection_settings 
        WHERE system = $1
        LIMIT 1
    `, systemID)
	if err != nil {
		return nil, err
	}
	return &setting, nil
}

func (r *connectionRepository) GetConnectionAuth(ctx context.Context, authID uuid.UUID) (*models.ConnectionAuthentication, error) {
	var auth models.ConnectionAuthentication
	err := r.db.GetContext(ctx, &auth, `
        SELECT ref, name, system, type, username, password, token 
        FROM connection_authentications 
        WHERE ref = $1
    `, authID)
	if err != nil {
		return nil, err
	}
	return &auth, nil
}

func (r *connectionRepository) CreateConnectionSetting(ctx context.Context, setting *models.ConnectionSetting) error {
	setting.Ref = uuid.New()
	_, err := r.db.ExecContext(ctx, `
        INSERT INTO connection_settings (ref, name, system, path, port, auth)
        VALUES ($1, $2, $3, $4, $5, $6)
    `, setting.Ref, setting.Name, setting.System, setting.Path, setting.Port, setting.AuthRef)
	return err
}

func (r *connectionRepository) CreateConnectionAuth(ctx context.Context, auth *models.ConnectionAuthentication) error {
	auth.Ref = uuid.New()
	_, err := r.db.ExecContext(ctx, `
        INSERT INTO connection_authentications (ref, name, system, type, username, password, token)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
    `, auth.Ref, auth.Name, auth.System, auth.Type, auth.Username, auth.Password, auth.Token)
	return err
}

