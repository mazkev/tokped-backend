package service

import (
	"context"
	"errors"

	"tokped-backend/internal/model"
	"tokped-backend/internal/repository"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type ProductService interface {
	GetAll(ctx context.Context, filter model.ProductFilter) ([]model.Product, error)
	GetByID(ctx context.Context, idStr string) (*model.Product, error)
	Create(ctx context.Context, req model.CreateProductRequest) (*model.Product, error)
	Update(ctx context.Context, idStr string, req model.UpdateProductRequest) (*model.Product, error)
	Delete(ctx context.Context, idStr string) error
}

type productService struct {
	productRepo repository.ProductRepository
}

func NewProductService(productRepo repository.ProductRepository) ProductService {
	return &productService{productRepo: productRepo}
}

func (s *productService) GetAll(ctx context.Context, filter model.ProductFilter) ([]model.Product, error) {
	return s.productRepo.FindAll(ctx, filter)
}

func (s *productService) GetByID(ctx context.Context, idStr string) (*model.Product, error) {
	oid, err := bson.ObjectIDFromHex(idStr)
	if err != nil {
		return nil, errors.New("ID produk tidak valid")
	}

	product, err := s.productRepo.FindByID(ctx, oid)
	if err != nil || product == nil {
		return nil, errors.New("produk tidak ditemukan")
	}

	return product, nil
}

func (s *productService) Create(ctx context.Context, req model.CreateProductRequest) (*model.Product, error) {
	// Jika originalPrice tidak diisi, gunakan harga normal
	origPrice := req.OriginalPrice
	if origPrice == 0 {
		origPrice = req.Price
	}

	product := &model.Product{
		Name:          req.Name,
		Price:         req.Price,
		OriginalPrice: origPrice,
		Discount:      req.Discount,
		Image:         req.Image,
		Rating:        req.Rating,
		Shop:          req.Shop,
		Location:      req.Location,
		Badge:         req.Badge,
		Condition:     req.Condition,
		Category:      req.Category,
		Stock:         req.Stock,
	}

	if err := s.productRepo.Create(ctx, product); err != nil {
		return nil, err
	}

	return product, nil
}

func (s *productService) Update(ctx context.Context, idStr string, req model.UpdateProductRequest) (*model.Product, error) {
	oid, err := bson.ObjectIDFromHex(idStr)
	if err != nil {
		return nil, errors.New("ID produk tidak valid")
	}

	existing, err := s.productRepo.FindByID(ctx, oid)
	if err != nil || existing == nil {
		return nil, errors.New("produk tidak ditemukan")
	}

	// Update field yang diisi
	if req.Name != "" { existing.Name = req.Name }
	if req.Price > 0 { existing.Price = req.Price }
	if req.OriginalPrice > 0 { existing.OriginalPrice = req.OriginalPrice }
	if req.Discount > 0 { existing.Discount = req.Discount }
	if req.Image != "" { existing.Image = req.Image }
	if req.Rating > 0 { existing.Rating = req.Rating }
	if req.Shop != "" { existing.Shop = req.Shop }
	if req.Location != "" { existing.Location = req.Location }
	if req.Badge != "" { existing.Badge = req.Badge }
	if req.Condition != "" { existing.Condition = req.Condition }
	if req.Category != "" { existing.Category = req.Category }
	if req.Stock >= 0 { existing.Stock = req.Stock }

	if err := s.productRepo.Update(ctx, oid, existing); err != nil {
		return nil, err
	}

	return existing, nil
}

func (s *productService) Delete(ctx context.Context, idStr string) error {
	oid, err := bson.ObjectIDFromHex(idStr)
	if err != nil {
		return errors.New("ID produk tidak valid")
	}

	return s.productRepo.Delete(ctx, oid)
}
