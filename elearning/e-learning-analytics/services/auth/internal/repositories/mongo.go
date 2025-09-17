// internal/repositories/mongo.go
package repositories

import (
	"context"
	"errors"
	"fmt"
	"time"

	"auth-service/internal/metrics"
	"auth-service/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"github.com/redis/go-redis/v9"
)

// NewMongoClient returns a connected *mongo.Client with sensible options.
func NewMongoClient(ctx context.Context, uri string) (*mongo.Client, error) {
	clientOpts := options.Client().ApplyURI(uri).
		SetMaxPoolSize(100).
		SetMinPoolSize(10).
		SetConnectTimeout(10 * time.Second)

	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return nil, err
	}
	// Ping to verify connection
	if err := client.Ping(ctx, nil); err != nil {
		_ = client.Disconnect(ctx)
		return nil, err
	}
	return client, nil
}

// UsersRepository handles simple user operations.
type UsersRepository struct {
	client     *mongo.Client
	collection *mongo.Collection
}

// NewUsersRepository returns a repository wired to "elearn.users" collection.
func NewUsersRepository(client *mongo.Client) *UsersRepository {
	coll := client.Database("elearn").Collection("users")
	// ensure index on email (unique)
	_ = ensureEmailIndex(coll)
	return &UsersRepository{client: client, collection: coll}
}

func ensureEmailIndex(coll *mongo.Collection) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	mod := mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}},
		Options: options.Index().SetUnique(true).SetBackground(true),
	}
	_, err := coll.Indexes().CreateOne(ctx, mod)
	return err
}

// Create inserts a new user document and returns its ID.
func (r *UsersRepository) Create(ctx context.Context, u models.User) (string, error) {
	res, err := r.collection.InsertOne(ctx, u)
	if err != nil {
		return "", err
	}
	id := fmt.Sprintf("%v", res.InsertedID)
	return id, nil
}

// ExistsByEmail checks if an email is already registered.
func (r *UsersRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	filter := bson.M{"email": email}
	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// FindByEmail retrieves a user by email.
func (r *UsersRepository) FindByEmail(ctx context.Context, email string) (models.User, error) {
	var u models.User
	err := r.collection.FindOne(ctx, bson.M{"email": email}).Decode(&u)
	if err != nil {
		metrics.MongoOperationsTotal.WithLabelValues("find", "users", "error").Inc()
	} else {
		metrics.MongoOperationsTotal.WithLabelValues("find", "users", "success").Inc()
	}
	return u, err
}

// FindByID retrieves a user by ID (string representation).
func (r *UsersRepository) FindByID(ctx context.Context, id string) (models.User, error) {
	var u models.User
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&u)
	return u, err
}

// FindUserIDByRefreshToken attempts to locate which user has the given refresh token.
// NOTE: This is a simple implementation for scaffolding. In production, store refresh token keyed by user ID
// and include the user ID in the refresh API call (or use a mapping).
func (r *UsersRepository) FindUserIDByRefreshToken(ctx context.Context, redisClient *redis.Client, refreshToken string) (string, error) {
	// Attempt simple scan: not scalable — placeholder for scaffold
	// In production, refresh token should be passed with user_id or token should be JWT itself.
	// We'll simulate by checking a common convention: "refresh_tokens:{userID}" for a small set of users.
	// For scaffold: attempt to list user documents and check redis key for each user.
	cursor, err := r.collection.Find(ctx, bson.M{}, options.Find().SetProjection(bson.M{"_id": 1}))
	if err != nil {
		return "", err
	}
	defer cursor.Close(ctx)
	for cursor.Next(ctx) {
		var doc struct {
			ID string `bson:"_id"`
		}
		if err := cursor.Decode(&doc); err != nil {
			continue
		}
		val, err := redisClient.Get(ctx, "refresh_tokens:"+doc.ID).Result()
		if err == nil && val == refreshToken {
			return doc.ID, nil
		}
		// ignore redis miss errors and continue
	}
	return "", errors.New("refresh token not found")
}
