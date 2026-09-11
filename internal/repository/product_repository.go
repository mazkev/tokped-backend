package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"tokped-backend/internal/model"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type ProductRepository interface {
	FindAll(ctx context.Context, filter model.ProductFilter) ([]model.Product, error)
	FindByID(ctx context.Context, id bson.ObjectID) (*model.Product, error)
	Create(ctx context.Context, product *model.Product) error
	Update(ctx context.Context, id bson.ObjectID, product *model.Product) error
	Delete(ctx context.Context, id bson.ObjectID) error
	SeedInitialProducts(ctx context.Context) error
}

type productRepository struct {
	collection *mongo.Collection
}

func NewProductRepository(db *mongo.Database) ProductRepository {
	repo := &productRepository{
		collection: db.Collection("products"),
	}

	// Otomatis seed data produk contoh jika collection masih kosong
	_ = repo.SeedInitialProducts(context.Background())

	return repo
}

func (r *productRepository) FindAll(ctx context.Context, filter model.ProductFilter) ([]model.Product, error) {
	query := bson.M{}

	if filter.Category != "" {
		query["category"] = bson.M{"$regex": filter.Category, "$options": "i"}
	}
	if filter.Search != "" {
		query["name"] = bson.M{"$regex": filter.Search, "$options": "i"}
	}
	if filter.Condition != "" {
		query["condition"] = filter.Condition
	}
	if filter.Location != "" {
		query["location"] = filter.Location
	}
	if filter.MinPrice > 0 || filter.MaxPrice > 0 {
		priceQuery := bson.M{}
		if filter.MinPrice > 0 {
			priceQuery["$gte"] = filter.MinPrice
		}
		if filter.MaxPrice > 0 {
			priceQuery["$lte"] = filter.MaxPrice
		}
		query["price"] = priceQuery
	}

	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	cursor, err := r.collection.Find(ctx, query, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var products []model.Product
	if err := cursor.All(ctx, &products); err != nil {
		return nil, err
	}

	if products == nil {
		products = []model.Product{}
	}
	return products, nil
}

func (r *productRepository) FindByID(ctx context.Context, id bson.ObjectID) (*model.Product, error) {
	var product model.Product
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&product)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return &product, nil
}

func (r *productRepository) Create(ctx context.Context, product *model.Product) error {
	product.CreatedAt = time.Now()
	product.UpdatedAt = time.Now()
	res, err := r.collection.InsertOne(ctx, product)
	if err != nil {
		return err
	}
	if oid, ok := res.InsertedID.(bson.ObjectID); ok {
		product.ID = oid
	}
	return nil
}

func (r *productRepository) Update(ctx context.Context, id bson.ObjectID, product *model.Product) error {
	product.UpdatedAt = time.Now()
	update := bson.M{
		"$set": bson.M{
			"name":           product.Name,
			"price":          product.Price,
			"original_price": product.OriginalPrice,
			"discount":       product.Discount,
			"image":          product.Image,
			"rating":         product.Rating,
			"shop":           product.Shop,
			"location":       product.Location,
			"badge":          product.Badge,
			"condition":      product.Condition,
			"category":       product.Category,
			"stock":          product.Stock,
			"updated_at":     product.UpdatedAt,
		},
	}
	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, update)
	return err
}

func (r *productRepository) Delete(ctx context.Context, id bson.ObjectID) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func (r *productRepository) SeedInitialProducts(ctx context.Context) error {
	count, err := r.collection.CountDocuments(ctx, bson.M{})
	if err != nil || count > 0 {
		return nil // Sudah ada data
	}

	initialProducts := []interface{}{
		model.Product{
			Name: "Apple iPhone 15 Pro Max 256GB Natural Titanium", Price: 21999000, OriginalPrice: 24999000, Discount: 12,
			Image: "https://images.tokopedia.net/img/cache/700/VqbcmM/2023/10/20/d193d5ff-c24c-474c-ba76-2f089d71ce04.jpg",
			Rating: 4.9, Sold: 450, Shop: "iBox Official", Location: "Jakarta Pusat", Badge: "official", Condition: "Baru", Category: "Elektronik", Stock: 50, CreatedAt: time.Now(), UpdatedAt: time.Now(),
		},
		model.Product{
			Name: "Sony WH-1000XM5 Wireless Noise Canceling Headphones", Price: 4999000, OriginalPrice: 5999000, Discount: 16,
			Image: "https://images.tokopedia.net/img/cache/700/VqbcmM/2022/6/15/4508ecbb-ca80-4cf8-a907-7ff763806be4.jpg",
			Rating: 4.8, Sold: 1200, Shop: "Sony Audio Official", Location: "Jakarta Selatan", Badge: "official", Condition: "Baru", Category: "Elektronik", Stock: 80, CreatedAt: time.Now(), UpdatedAt: time.Now(),
		},
		model.Product{
			Name: "Kaos Polos Pria Heavyweight Cotton Combed 24s", Price: 65000, OriginalPrice: 85000, Discount: 23,
			Image: "https://images.tokopedia.net/img/cache/700/VqbcmM/2021/11/4/cb9515ee-63f5-4424-95eb-a83ebba1a884.jpg",
			Rating: 4.7, Sold: 15400, Shop: "BasicWear ID", Location: "Bandung", Badge: "power-merchant", Condition: "Baru", Category: "Fashion Pria", Stock: 500, CreatedAt: time.Now(), UpdatedAt: time.Now(),
		},
		model.Product{
			Name: "Jam Tangan Pria Automatic Skeleton Luxury Stainless", Price: 350000, OriginalPrice: 700000, Discount: 50,
			Image: "https://images.tokopedia.net/img/cache/700/VqbcmM/2022/1/10/242d99d1-0f49-411a-85d9-482a4ebf0c45.jpg",
			Rating: 4.6, Sold: 890, Shop: "TimeMaster Store", Location: "Jakarta Barat", Badge: "power-merchant", Condition: "Baru", Category: "Perhiasan", Stock: 100, CreatedAt: time.Now(), UpdatedAt: time.Now(),
		},
		model.Product{
			Name: "Sepatu Sneaker Pria Casual Sporty Slip-on Breathable", Price: 189000, OriginalPrice: 299000, Discount: 36,
			Image: "https://images.tokopedia.net/img/cache/700/VqbcmM/2022/9/19/2d88fa7b-9c98-4ae7-a417-0b1a0e1c07ad.jpg",
			Rating: 4.8, Sold: 3200, Shop: "SneakerZone Official", Location: "Surabaya", Badge: "official", Condition: "Baru", Category: "Fashion Pria", Stock: 200, CreatedAt: time.Now(), UpdatedAt: time.Now(),
		},
	}

	_, err = r.collection.InsertMany(ctx, initialProducts)
	if err == nil {
		fmt.Println("📦 Data awal produk berhasil di-seed ke MongoDB!")
	}
	return err
}
