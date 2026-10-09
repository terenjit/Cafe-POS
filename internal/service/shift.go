package service

import (
	"context"
	"errors"

	"github.com/terenjit/Cafe-POS/internal/entity"
	"github.com/terenjit/Cafe-POS/internal/repository"
)

type ShiftService struct {
	shiftRepo repository.ShiftRepository
	orderRepo repository.OrderRepository
}

func NewShiftService(shiftRepo repository.ShiftRepository, orderRepo repository.OrderRepository) *ShiftService {
	return &ShiftService{shiftRepo: shiftRepo, orderRepo: orderRepo}
}

func (s *ShiftService) Open(ctx context.Context, cashierID string, req entity.OpenShiftRequest) (*entity.Shift, error) {
	existing, err := s.shiftRepo.FindOpenByCashierID(ctx, cashierID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("you already have an active shift")
	}

	shift := &entity.Shift{
		CashierID:   cashierID,
		OpeningCash: req.OpeningCash,
		Notes:       req.Notes,
		Status:      entity.ShiftStatusOpen,
	}

	if err := s.shiftRepo.Create(ctx, shift); err != nil {
		return nil, err
	}

	return s.shiftRepo.FindByID(ctx, shift.ID)
}

func (s *ShiftService) Close(ctx context.Context, cashierID string, req entity.CloseShiftRequest) (*entity.Shift, error) {
	shift, err := s.shiftRepo.FindOpenByCashierID(ctx, cashierID)
	if err != nil {
		return nil, err
	}
	if shift == nil {
		return nil, errors.New("no active shift found")
	}

	totalTransactions, err := s.orderRepo.CountPaidByShiftID(ctx, shift.ID)
	if err != nil {
		return nil, err
	}

	if err := s.shiftRepo.UpdateTotalTransactions(ctx, shift.ID, totalTransactions); err != nil {
		return nil, err
	}

	if err := s.shiftRepo.Close(ctx, shift.ID, req.ClosingCash, req.Notes); err != nil {
		return nil, err
	}

	return s.shiftRepo.FindByID(ctx, shift.ID)
}

func (s *ShiftService) GetCurrent(ctx context.Context, cashierID string) (*entity.Shift, error) {
	shift, err := s.shiftRepo.FindOpenByCashierID(ctx, cashierID)
	if err != nil {
		return nil, err
	}
	if shift == nil {
		return nil, errors.New("no active shift found")
	}

	return s.shiftRepo.FindByID(ctx, shift.ID)
}
