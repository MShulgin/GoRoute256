package offer

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/jmoiron/sqlx"
	"gitlab.ozon.dev/MShulgin/homework-3/common/pkg/cache"
	"gitlab.ozon.dev/MShulgin/homework-3/common/pkg/ex"
	"gitlab.ozon.dev/MShulgin/homework-3/common/pkg/logger"
)

const ID_SEQ = "offer_id"

type Storage interface {
	GetOffer(offerId string) (*Offer, *ex.AppError)
	SaveOffer(offer Offer) *ex.AppError
	UpdateOffer(offerId string, updateFn func(Offer) Offer) (*Offer, *ex.AppError)
	RemoveReserved(offers map[string]int64) *ex.AppError
	NextId() (int64, *ex.AppError)
}

type PgStorage struct {
	Db *sqlx.DB
}

func getOfferById(tx *sqlx.Tx, offerId string) (*Offer, *ex.AppError) {
	var o Offer
	if err := tx.Get(&o, "SELECT * FROM offer WHERE id = $1", offerId); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ex.NewNotFoundError(fmt.Sprintf("Not found offer: offerId='%s'", offerId))
		}
		logger.Error("failed to get offer: " + err.Error())
		return nil, ex.NewUnexpectedError("Unexpected database error")
	}
	return &o, nil
}

func (s *PgStorage) GetOffer(offerId string) (*Offer, *ex.AppError) {
	tx, err := s.Db.Beginx()
	if err != nil {
		logger.Error("failed to open transaction: " + err.Error())
		return nil, ex.NewUnexpectedError("Unexpected database error")
	}
	defer func(tx *sqlx.Tx) {
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			logger.Error("error during transaction rollback: " + err.Error())
		}
	}(tx)

	o, appErr := getOfferById(tx, offerId)
	if appErr != nil {
		return nil, appErr
	}

	err = tx.Commit()
	if err != nil {
		logger.Error("failed to commit getting offer: " + err.Error())
		return nil, ex.NewUnexpectedError("Unexpected database error")
	}
	return o, nil
}

func (s *PgStorage) SaveOffer(o Offer) *ex.AppError {
	_, err := s.Db.Exec("INSERT INTO offer VALUES ($1, $2, $3, $4, $5, $6)",
		o.Id, o.SellerId, o.ProductId, o.Stock, o.Reserved, o.Price)
	if err != nil {
		logger.Error("failed to save offer: " + err.Error())
		return ex.NewUnexpectedError("Unexpected database error")
	}
	return nil
}

func (s *PgStorage) UpdateOffer(offerId string, updateFn func(Offer) Offer) (*Offer, *ex.AppError) {
	tx, err := s.Db.Beginx()
	if err != nil {
		logger.Error("failed to open transaction: " + err.Error())
		return nil, ex.NewUnexpectedError("Unexpected database error")
	}
	defer func(tx *sqlx.Tx) {
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			logger.Error("error during transaction rollback: " + err.Error())
		}
	}(tx)

	off, appErr := getOfferById(tx, offerId)
	if appErr != nil {
		return nil, appErr
	}

	updated := updateFn(*off)

	exec, err := tx.Exec("UPDATE offer SET stock = $2, reserved = $3, price = $4 WHERE id = $1",
		offerId, updated.Stock, updated.Reserved, updated.Price)
	if err != nil {
		if rows, err := exec.RowsAffected(); err != nil {
			logger.Error("failed to get affected rows: " + err.Error())
			return nil, ex.NewUnexpectedError("Unexpected database error")
		} else {
			if rows == 0 {
				return nil, ex.NewNotFoundError(fmt.Sprintf("not found offer: offerId=%s", offerId))
			}
		}
		logger.Error("failed to update offer: " + err.Error())
		return nil, ex.NewUnexpectedError("Unexpected database error")
	}
	err = tx.Commit()
	if err != nil {
		logger.Error("failed to commit offer update: " + err.Error())
		return nil, ex.NewUnexpectedError("Unexpected database error")
	}

	return &updated, nil
}

func (s *PgStorage) RemoveReserved(offers map[string]int64) *ex.AppError {
	tx, err := s.Db.Beginx()
	if err != nil {
		logger.Error("failed to open transaction: " + err.Error())
		return ex.NewUnexpectedError("Unexpected database error")
	}
	defer func(tx *sqlx.Tx) {
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			logger.Error("error during transaction rollback: " + err.Error())
		}
	}(tx)

	for offerId, count := range offers {
		result, err := tx.Exec("UPDATE offer SET reserved = reserved - $1 WHERE id = $2", count, offerId)
		if err != nil {
			logger.Error("failed to update offer: " + err.Error())
			return ex.NewUnexpectedError("Unexpected database error")
		}
		if aff, err := result.RowsAffected(); err == nil {
			if aff == 0 {
				return ex.NewNotFoundError(fmt.Sprintf("not found offer: offerId=%s", offerId))
			}
		} else {
			logger.Error("failed to get affected rows: " + err.Error())
			return ex.NewUnexpectedError("Unexpected database error")
		}
	}

	err = tx.Commit()
	if err != nil {
		logger.Error("failed to commit offer update: " + err.Error())
		return ex.NewUnexpectedError("Unexpected database error")
	}

	return nil
}

func (s *PgStorage) NextId() (int64, *ex.AppError) {
	var nextId int64
	if err := s.Db.Get(&nextId, fmt.Sprintf("SELECT nextval('%s')", ID_SEQ)); err != nil {
		logger.Error("failed to get offerId from sequence: " + err.Error())
		return 0, ex.NewUnexpectedError("Unexpected database error")
	}
	return nextId, nil
}

type PgCachedStorage struct {
	dbRepo PgStorage
	cache  cache.Cache[Offer]
}

func NewPgCachedStorage(cache cache.Cache[Offer], db *sqlx.DB) *PgCachedStorage {
	return &PgCachedStorage{dbRepo: PgStorage{db}, cache: cache}
}

func (p PgCachedStorage) GetOffer(offerId string) (*Offer, *ex.AppError) {
	off, err := p.cache.Get(offerId)
	if err != nil {
		if errors.Is(err, cache.CacheMissError{}) {
			off, appErr := p.dbRepo.GetOffer(offerId)
			if appErr != nil {
				return nil, appErr
			}
			if e := p.cache.Set(offerId, off); e != nil {
				logger.Error("failed to put offer in cache: " + e.Error())
				return nil, ex.NewUnexpectedError("Unexpected error from cache: " + e.Error())
			}
			return off, nil
		}
		logger.Error("failed to get offer from cache: " + err.Error())
		return nil, ex.NewUnexpectedError("Unexpected error from cache: " + err.Error())
	}
	return off, nil
}

func (p PgCachedStorage) SaveOffer(offer Offer) *ex.AppError {
	err := p.cache.Invalidate(offer.Id)
	if err != nil {
		logger.Error("failed to invalidate cache: " + err.Error())
		return ex.NewUnexpectedError("Unexpected error from cache: " + err.Error())
	}
	return p.dbRepo.SaveOffer(offer)
}

func (p PgCachedStorage) UpdateOffer(offerId string, updateFn func(Offer) Offer) (*Offer, *ex.AppError) {
	err := p.cache.Invalidate(offerId)
	if err != nil {
		logger.Error("failed to invalidate cache: " + err.Error())
		return nil, ex.NewUnexpectedError("Unexpected error from cache: " + err.Error())
	}
	return p.dbRepo.UpdateOffer(offerId, updateFn)
}

func (p PgCachedStorage) RemoveReserved(offers map[string]int64) *ex.AppError {
	for offerId := range offers {
		err := p.cache.Invalidate(offerId)
		if err != nil {
			logger.Error("failed to invalidate cache: " + err.Error())
			return ex.NewUnexpectedError("Unexpected error from cache: " + err.Error())
		}
	}
	return p.dbRepo.RemoveReserved(offers)
}

func (p PgCachedStorage) NextId() (int64, *ex.AppError) {
	return p.dbRepo.NextId()
}
