package service

import (
	"context"
	"database/sql"
	"errors"

	"github.com/terenjit/Cafe-POS/internal/entity"
	"github.com/terenjit/Cafe-POS/internal/repository"
	"github.com/terenjit/Cafe-POS/pkg/txmanager"
)

type StockService struct {
	stockRepo   repository.StockRepository
	productRepo repository.ProductRepository
	txManager   *txmanager.TxManager
}

func NewStockService(
	stockRepo repository.StockRepository,
	productRepo repository.ProductRepository,
	txManager *txmanager.TxManager,
) *StockService {
	return &StockService{stockRepo: stockRepo, productRepo: productRepo, txManager: txManager}
}

func (s *StockService) GetStock(ctx context.Context, productID string) (*entity.Product, error) {
	product, err := s.productRepo.FindByID(ctx, productID)
	if err != nil {
		return nil, err
	}
	if product == nil {
		return nil, errors.New("product not found")
	}
	return product, nil
}

func (s *StockService) GetMovements(ctx context.Context, productID string, filter repository.StockFilter) ([]entity.StockMovement, int, error) {
	product, err := s.productRepo.FindByID(ctx, productID)
	if err != nil {
		return nil, 0, err
	}
	if product == nil {
		return nil, 0, errors.New("product not found")
	}

	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.Limit <= 0 {
		filter.Limit = 20
	}

	return s.stockRepo.FindByProductID(ctx, productID, filter)
}

func (s *StockService) Adjust(ctx context.Context, productID string, userID string, req entity.StockAdjustmentRequest) error {
	product, err := s.productRepo.FindByID(ctx, productID)
	if err != nil {
		return err
	}
	if product == nil {
		return errors.New("product not found")
	}

	qty := req.Quantity
	if req.Type == entity.MovementTypeOut {
		qty = qty * -1
	}

	var newStock int
	if req.Type == entity.MovementTypeOut {
		newStock = product.Stock - req.Quantity
	} else {
		newStock = product.Stock + req.Quantity
	}
	if newStock < 0 {
		return errors.New("insufficient stock")
	}

	return s.txManager.WithTx(ctx, func(tx *sql.Tx) error {
		txStockRepo := s.stockRepo.WithTx(tx)
		movement := &entity.StockMovement{
			ProductID:   productID,
			UserID:      userID,
			Type:        req.Type,
			Quantity:    qty,
			StockBefore: product.Stock,
			StockAfter:  newStock,
			Notes:       req.Notes,
		}
		if err := txStockRepo.Create(ctx, movement); err != nil {
			return err
		}

		txProductRepo := s.productRepo.WithTx(tx)
		return txProductRepo.UpdateStock(ctx, productID, newStock)
	})
}
