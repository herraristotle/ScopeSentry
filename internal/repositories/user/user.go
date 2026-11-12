package user

import (
	"context"
	"time"

	"github.com/Autumn-27/ScopeSentry/internal/database/mongodb"
	"github.com/Autumn-27/ScopeSentry/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	mongodriver "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Repository interface {
	FindByUsername(ctx context.Context, username string) (*models.User, error)
	Create(ctx context.Context, user *models.User) error
	FindAll(ctx context.Context) ([]*models.User, error)
	FindByID(ctx context.Context, id string) (*models.User, error)
	Update(ctx context.Context, id string, user *models.User) error
	UpdateFields(ctx context.Context, id string, fields bson.M) error
	Delete(ctx context.Context, id string) error
}

type repository struct {
	collection *mongodriver.Collection
}

func NewRepository() Repository {
	return &repository{
		collection: mongodb.DB.Collection("user"),
	}
}

// Create creates a user
func (r *repository) Create(ctx context.Context, user *models.User) error {
	user.ID = primitive.NewObjectID()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	_, err := r.collection.InsertOne(ctx, user)
	return err
}

// FindByID gets a user by ID
func (r *repository) FindByID(ctx context.Context, id string) (*models.User, error) {
	var user models.User
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByUsername gets a user by username
func (r *repository) FindByUsername(ctx context.Context, username string) (*models.User, error) {
	var user models.User
	err := r.collection.FindOne(ctx, bson.M{"username": username}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindAll returns the list of users
func (r *repository) FindAll(ctx context.Context) ([]*models.User, error) {
	// Set pagination options
	opts := options.Find().
		SetSort(bson.D{{"createdAt", -1}})

	// Query the user list
	cursor, err := r.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []*models.User
	if err = cursor.All(ctx, &users); err != nil {
		return nil, err
	}

	return users, nil
}

// Update updates a user
func (r *repository) Update(ctx context.Context, id string, user *models.User) error {
	updateDoc := bson.M{
		"$set": user,
	}
	objectId, _ := primitive.ObjectIDFromHex(id)
	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": objectId}, updateDoc)
	return err
}

// UpdateFields updates by field (updates only the provided fields to avoid zero-value overwrite)
func (r *repository) UpdateFields(ctx context.Context, id string, fields bson.M) error {
	updateDoc := bson.M{
		"$set": fields,
	}
	objectId, _ := primitive.ObjectIDFromHex(id)
	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": objectId}, updateDoc)
	return err
}

// Delete deletes a user
func (r *repository) Delete(ctx context.Context, id string) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}
