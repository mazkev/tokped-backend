package repository

import (
	"context"
	"time"

	"tokped-backend/internal/model"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type ReviewRepository interface {
	Create(ctx context.Context, review *model.Review) error
	FindByProductID(ctx context.Context, productID string) ([]model.Review, error)
	GetAverageRating(ctx context.Context, productID string) (float64, error)
}

type reviewRepository struct {
	collection *mongo.Collection
}

func NewReviewRepository(db *mongo.Database) ReviewRepository {
	return &reviewRepository{
		collection: db.Collection("reviews"),
	}
}

func (r *reviewRepository) Create(ctx context.Context, review *model.Review) error {
	review.CreatedAt = time.Now()
	res, err := r.collection.InsertOne(ctx, review)
	if err != nil {
		return err
	}
	if oid, ok := res.InsertedID.(bson.ObjectID); ok {
		review.ID = oid
	}
	return nil
}

func (r *reviewRepository) FindByProductID(ctx context.Context, productID string) ([]model.Review, error) {
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}})
	cursor, err := r.collection.Find(ctx, bson.M{"product_id": productID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var reviews []model.Review
	if err := cursor.All(ctx, &reviews); err != nil {
		return nil, err
	}
	if reviews == nil {
		reviews = []model.Review{}
	}
	return reviews, nil
}

func (r *reviewRepository) GetAverageRating(ctx context.Context, productID string) (float64, error) {
	pipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: bson.M{"product_id": productID}}},
		bson.D{{Key: "$group", Value: bson.M{
			"_id":       "$product_id",
			"avgRating": bson.M{"$avg": "$rating"},
		}}},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return 0, err
	}
	defer cursor.Close(ctx)

	var results []struct {
		AvgRating float64 `bson:"avgRating"`
	}
	if err := cursor.All(ctx, &results); err != nil || len(results) == 0 {
		return 0, nil
	}

	return results[0].AvgRating, nil
}
