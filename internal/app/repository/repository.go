package repository

import (
	"DIA_Backend/internal/app/dsn"
	"DIA_Backend/internal/app/redis"
	"context"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type JWTConfig struct {
	Secret        string
	ExpiresIn     time.Duration
	RefreshIn     time.Duration
	SigningMethod jwt.SigningMethod
}

type Repository struct {
	db          *gorm.DB
	Cost        *CostRepository
	CostRequest *CostRequestRepository
	User        *UserRepository
	redis       *redis.Client
}

func (r *Repository) GetJWTSecret() string {
	return os.Getenv("JWT_SECRET")
}

func (r *Repository) GetJWTConfig() *JWTConfig {
	secret := r.GetJWTSecret()
	if secret == "" {
		secret = "default-jwt-secret-key"
	}

	return &JWTConfig{
		Secret:        secret,
		ExpiresIn:     time.Hour * 24,
		RefreshIn:     time.Hour * 24 * 7,
		SigningMethod: jwt.SigningMethodHS256,
	}
}

func NewRepository() (*Repository, error) {
	db, err := gorm.Open(postgres.Open(dsn.FromEnv()), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	minioClient, err := InitMinioClient()
	if err != nil {
		return nil, err
	}

	return &Repository{
		db:          db,
		Cost:        NewCostRepository(db, minioClient),
		CostRequest: NewCostRequestRepository(db),
		User:        NewUserRepository(db),
	}, nil
}

func CloseDBConn(r *Repository) {
	dbInstance, _ := r.db.DB()
	_ = dbInstance.Close()
}

func (r *Repository) GetRedisClient() *redis.Client {
	return r.redis
}

func InitMinioClient() (*minio.Client, error) {
	endpoint := os.Getenv("MINIO_HOST") + ":" + os.Getenv("MINIO_SERVER_PORT")
	accessKeyID := os.Getenv("MINIO_ACCESS_KEY")
	secretAccessKey := os.Getenv("MINIO_SECRET_KEY")
	useSSL := false

	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO client: %v", err)
	}

	ctx := context.Background()

	exists, err := minioClient.BucketExists(ctx, costImagesBucket)
	if err != nil {
		return nil, fmt.Errorf("failed to check bucket existence: %v", err)
	}

	if !exists {
		err = minioClient.MakeBucket(ctx, costImagesBucket, minio.MakeBucketOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to create bucket: %v", err)
		}
		logrus.Printf("Bucket '%s' created successfully\n", costImagesBucket)
	}

	return minioClient, nil
}
