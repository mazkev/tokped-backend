package service

import (
	"context"
	"errors"
	"math"

	"tokped-backend/internal/model"
	"tokped-backend/internal/repository"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type ReviewService interface {
	CreateReview(ctx context.Context, userID, userName string, req model.CreateReviewRequest) (*model.Review, error)
	GetProductReviews(ctx context.Context, productID string) ([]model.Review, error)
}

type reviewService struct {
	reviewRepo  repository.ReviewRepository
	orderRepo   repository.OrderRepository
	productRepo repository.ProductRepository
}

func NewReviewService(reviewRepo repository.ReviewRepository, orderRepo repository.OrderRepository, productRepo repository.ProductRepository) ReviewService {
	return &reviewService{
		reviewRepo:  reviewRepo,
		orderRepo:   orderRepo,
		productRepo: productRepo,
	}
}

func (s *reviewService) CreateReview(ctx context.Context, userID, userName string, req model.CreateReviewRequest) (*model.Review, error) {
	// 1. Validasi rating 1 - 5
	if req.Rating < 1 || req.Rating > 5 {
		return nil, errors.New("rating harus bernilai antara 1 sampai 5")
	}

	// 2. Simpan ulasan
	review := &model.Review{
		OrderID:   req.OrderID,
		ProductID: req.ProductID,
		UserID:    userID,
		UserName:  userName,
		Rating:    req.Rating,
		Comment:   req.Comment,
	}

	if err := s.reviewRepo.Create(ctx, review); err != nil {
		return nil, err
	}

	// 3. Tandai pesanan sudah diulas
	if orderOID, err := bson.ObjectIDFromHex(req.OrderID); err == nil {
		_ = s.orderRepo.SetOrderReviewed(ctx, orderOID)
	}

	// 4. Hitung ulang rata-rata rating produk
	if productOID, err := bson.ObjectIDFromHex(req.ProductID); err == nil {
		avgRating, err := s.reviewRepo.GetAverageRating(ctx, req.ProductID)
		if err == nil && avgRating > 0 {
			// Bulatkan 1 angka di belakang koma (misal: 4.8)
			rounded := math.Round(avgRating*10) / 10
			_ = s.productRepo.Update(ctx, productOID, &model.Product{Rating: rounded})
		}
	}

	return review, nil
}

func (s *reviewService) GetProductReviews(ctx context.Context, productID string) ([]model.Review, error) {
	return s.reviewRepo.FindByProductID(ctx, productID)
}
