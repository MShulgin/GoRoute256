package store

import (
	"github.com/jmoiron/sqlx"
	"gitlab.ozon.dev/MShulgin/homework-2/portfolio/internal/ex"
	"gitlab.ozon.dev/MShulgin/homework-2/portfolio/internal/logging"
	"gitlab.ozon.dev/MShulgin/homework-2/portfolio/internal/model"
)

type AssetRepo interface {
	SaveAsset(model.Asset) *ex.AppError
}

type assetRepoDb struct {
	db *sqlx.DB
}

func NewAssetRepo(db *sqlx.DB) AssetRepo {
	return assetRepoDb{db: db}
}

func (r assetRepoDb) SaveAsset(asset model.Asset) *ex.AppError {
	insertSql := "INSERT INTO asset (code) VALUES ($1) ON CONFLICT (code) DO NOTHING"
	_, err := r.db.Exec(insertSql, asset.Code)
	if err != nil {
		logging.Error("Failed to save new asset: " + err.Error())
		return ex.NewUnexpectedError("Unexpected database error")
	}
	return nil
}
