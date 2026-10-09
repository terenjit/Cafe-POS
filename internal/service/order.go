package service

import (
	"context"
	"database/sql"
	"errors"

	"github.com/terenjit/Cafe-POS/internal/entity"
	"github.com/terenjit/Cafe-POS/internal/repository"
	"github.com/terenjit/Cafe-POS/pkg/txmanager"
)

type OrderService struct {
	orderRepo   repository.OrderRepository
	shiftRepo   repository.ShiftRepository
	tableRepo   repository.TableRepository
	productRepo repository.ProductRepository
	promoRepo   repository.PromoRepository
	paymentRepo repository.PaymentRepository
	txMgr       *txmanager.TxManager
}

func NewOrderService(
	orderRepo repository.OrderRepository,
	shiftRepo repository.ShiftRepository,
	tableRepo repository.TableRepository,
	productRepo repository.ProductRepository,
	promoRepo repository.PromoRepository,
	paymentRepo repository.PaymentRepository,
	txMgr *txmanager.TxManager,
) *OrderService {
	return &OrderService{
		orderRepo:   orderRepo,
		shiftRepo:   shiftRepo,
		tableRepo:   tableRepo,
		productRepo: productRepo,
		promoRepo:   promoRepo,
		paymentRepo: paymentRepo,
		txMgr:       txMgr,
	}
}

func (s *OrderService) UpdateItem(ctx context.Context, cashierID string, orderID string, itemID string, req entity.UpdateOrderItemRequest) (*entity.Order, error) {
	order, err := s.orderRepo.FindByID(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, errors.New("order not found")
	}
	if order.CashierID != cashierID {
		return nil, errors.New("you are not authorized to modify this order")
	}
	if order.Status != entity.OrderStatusDraft {
		return nil, errors.New("order cannot be modified, status is not draft")
	}

	item, err := s.orderRepo.FindItemByID(ctx, itemID)
	if err != nil {
		return nil, err
	}
	if item == nil || item.OrderID != orderID {
		return nil, errors.New("item not found")
	}

	item.Quantity = req.Quantity
	item.Notes = req.Notes
	item.Subtotal = item.Price * int64(req.Quantity)

	if err := s.txMgr.WithTx(ctx, func(tx *sql.Tx) error {
		txRepo := s.orderRepo.WithTx(tx)
		if err := txRepo.UpdateItem(ctx, item); err != nil {
			return err
		}
		return txRepo.RecalculateTotal(ctx, orderID)
	}); err != nil {
		return nil, err
	}

	return s.orderRepo.FindByID(ctx, orderID)
}

func (s *OrderService) RemoveItem(ctx context.Context, cashierID string, orderID string, itemID string) (*entity.Order, error) {
	order, err := s.orderRepo.FindByID(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, errors.New("order not found")
	}
	if order.CashierID != cashierID {
		return nil, errors.New("you are not authorized to modify this order")
	}
	if order.Status != entity.OrderStatusDraft {
		return nil, errors.New("order cannot be modified, status is not draft")
	}

	item, err := s.orderRepo.FindItemByID(ctx, itemID)
	if err != nil {
		return nil, err
	}
	if item == nil || item.OrderID != orderID {
		return nil, errors.New("item not found")
	}

	if err := s.txMgr.WithTx(ctx, func(tx *sql.Tx) error {
		txRepo := s.orderRepo.WithTx(tx)
		if err := txRepo.DeleteItem(ctx, itemID); err != nil {
			return err
		}
		return txRepo.RecalculateTotal(ctx, orderID)
	}); err != nil {
		return nil, err
	}

	return s.orderRepo.FindByID(ctx, orderID)
}

func (s *OrderService) ApplyPromo(ctx context.Context, cashierID string, orderID string, req entity.ApplyPromoRequest) (*entity.Order, error) {
	order, err := s.orderRepo.FindByID(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, errors.New("order not found")
	}
	if order.CashierID != cashierID {
		return nil, errors.New("you are not authorized to modify this order")
	}
	if order.Status != entity.OrderStatusDraft {
		return nil, errors.New("order cannot be modified, status is not draft")
	}

	promo, err := s.promoRepo.FindByCode(ctx, req.PromoCode)
	if err != nil {
		return nil, err
	}
	if promo == nil {
		return nil, errors.New("invalid promo code")
	}
	if !promo.IsValid() {
		return nil, errors.New("promo is not active or has expired")
	}
	if order.Subtotal < promo.MinOrder {
		return nil, errors.New("order subtotal does not meet the promo minimum purchase requirement")
	}

	var discount int64
	if promo.Type == entity.PromoTypePercentage {
		discount = order.Subtotal * promo.Value / 100
		if promo.MaxDiscount != nil && discount > *promo.MaxDiscount {
			discount = *promo.MaxDiscount
		}
	} else {
		discount = promo.Value
		if discount > order.Subtotal {
			discount = order.Subtotal
		}
	}

	promoID := promo.ID
	order.PromoID = &promoID
	order.DiscountAmount = discount
	order.Total = order.Subtotal - discount

	if err := s.txMgr.WithTx(ctx, func(tx *sql.Tx) error {
		if err := s.orderRepo.WithTx(tx).Update(ctx, order); err != nil {
			return err
		}
		return s.promoRepo.WithTx(tx).IncrementUsedCount(ctx, promo.ID)
	}); err != nil {
		return nil, err
	}

	return s.orderRepo.FindByID(ctx, orderID)
}

func (s *OrderService) RemovePromo(ctx context.Context, cashierID string, orderID string) (*entity.Order, error) {
	order, err := s.orderRepo.FindByID(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, errors.New("order not found")
	}
	if order.CashierID != cashierID {
		return nil, errors.New("you are not authorized to modify this order")
	}
	if order.Status != entity.OrderStatusDraft {
		return nil, errors.New("order cannot be modified, status is not draft")
	}
	if order.PromoID == nil {
		return nil, errors.New("no promo applied to this order")
	}

	order.PromoID = nil
	order.DiscountAmount = 0
	order.Total = order.Subtotal

	if err := s.orderRepo.Update(ctx, order); err != nil {
		return nil, err
	}

	return s.orderRepo.FindByID(ctx, orderID)
}

func (s *OrderService) Checkout(ctx context.Context, cashierID string, orderID string, req entity.CheckoutRequest) (*entity.CheckoutResponse, error) {
	order, err := s.orderRepo.FindByID(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, errors.New("order not found")
	}
	if order.CashierID != cashierID {
		return nil, errors.New("you are not authorized to modify this order")
	}
	if order.Status != entity.OrderStatusDraft {
		return nil, errors.New("order cannot be checked out, status is not draft")
	}
	if len(order.Items) == 0 {
		return nil, errors.New("order cannot be checked out without items")
	}

	existing, err := s.paymentRepo.FindByOrderID(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("this order is already being processed for payment")
	}

	payment := &entity.Payment{
		OrderID: orderID,
		Method:  req.Method,
		Amount:  order.Total,
		Status:  entity.PaymentStatusPending,
	}

	order.Status = entity.OrderStatusPendingPayment

	if err := s.txMgr.WithTx(ctx, func(tx *sql.Tx) error {
		if err := s.paymentRepo.WithTx(tx).Create(ctx, payment); err != nil {
			return err
		}
		return s.orderRepo.WithTx(tx).Update(ctx, order)
	}); err != nil {
		return nil, err
	}

	return &entity.CheckoutResponse{Payment: *payment}, nil
}

func (s *OrderService) FindOrders(ctx context.Context, cashierID string, filter repository.OrderFilter) ([]entity.Order, int, error) {
	shift, err := s.shiftRepo.FindOpenByCashierID(ctx, cashierID)
	if err != nil {
		return nil, 0, err
	}
	if shift == nil {
		return nil, 0, errors.New("no active shift found")
	}

	filter.ShiftID = shift.ID
	filter.CashierID = cashierID

	return s.orderRepo.FindAll(ctx, filter)
}

func (s *OrderService) FindByID(ctx context.Context, orderID string) (*entity.Order, error) {
	order, err := s.orderRepo.FindByID(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, errors.New("order not found")
	}
	return order, nil
}

func (s *OrderService) AddItem(ctx context.Context, cashierID string, orderID string, req entity.AddOrderItemRequest) (*entity.Order, error) {
	order, err := s.orderRepo.FindByID(ctx, orderID)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, errors.New("order not found")
	}

	if order.CashierID != cashierID {
		return nil, errors.New("you are not authorized to modify this order")
	}

	if order.Status != entity.OrderStatusDraft {
		return nil, errors.New("order cannot be modified, status is not draft")
	}

	product, err := s.productRepo.FindByID(ctx, req.ProductID)
	if err != nil {
		return nil, err
	}
	if product == nil {
		return nil, errors.New("product not found")
	}
	if !product.IsActive {
		return nil, errors.New("product is not available")
	}
	if product.Stock < req.Quantity {
		return nil, errors.New("insufficient stock")
	}

	item := &entity.OrderItem{
		OrderID:   orderID,
		ProductID: req.ProductID,
		Price:     product.Price,
		Quantity:  req.Quantity,
		Notes:     req.Notes,
	}

	if err := s.txMgr.WithTx(ctx, func(tx *sql.Tx) error {
		txRepo := s.orderRepo.WithTx(tx)
		if err := txRepo.AddItem(ctx, item); err != nil {
			return err
		}
		return txRepo.RecalculateTotal(ctx, orderID)
	}); err != nil {
		return nil, err
	}

	return s.orderRepo.FindByID(ctx, orderID)
}

func (s *OrderService) Create(ctx context.Context, cashierID string, req entity.CreateOrderRequest) (*entity.Order, error) {
	shift, err := s.shiftRepo.FindOpenByCashierID(ctx, cashierID)
	if err != nil {
		return nil, err
	}
	if shift == nil {
		return nil, errors.New("you have not opened a shift")
	}

	order := &entity.Order{
		ShiftID:   shift.ID,
		CashierID: cashierID,
		Status:    entity.OrderStatusDraft,
		Notes:     req.Notes,
	}

	if req.TableID != "" {
		table, err := s.tableRepo.FindByID(ctx, req.TableID)
		if err != nil {
			return nil, err
		}
		if table == nil {
			return nil, errors.New("table not found")
		}
		if table.Status != entity.TableStatusAvailable {
			return nil, errors.New("table is not available")
		}

		tableIDStr := req.TableID
		order.TableID = &tableIDStr

		if err := s.txMgr.WithTx(ctx, func(tx *sql.Tx) error {
			if err := s.orderRepo.WithTx(tx).Create(ctx, order); err != nil {
				return err
			}
			return s.tableRepo.WithTx(tx).UpdateStatus(ctx, req.TableID, entity.TableStatusOccupied)
		}); err != nil {
			return nil, err
		}
	} else {
		if err := s.orderRepo.Create(ctx, order); err != nil {
			return nil, err
		}
	}

	return s.orderRepo.FindByID(ctx, order.ID)
}
