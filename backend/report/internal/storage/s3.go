package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"report/internal/config"
	"report/internal/models"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type S3Client struct {
	client     *minio.Client
	bucketName string
}

func NewS3Client(cfg config.S3Config) *S3Client {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		panic(fmt.Sprintf("Failed to create S3 client: %v", err))
	}

	return &S3Client{
		client:     client,
		bucketName: cfg.BucketName,
	}
}

func (s *S3Client) ReportExists(ctx context.Context, userID int) (bool, error) {
	fileName := s.fileName(userID)

	_, err := s.client.StatObject(ctx, s.bucketName, fileName, minio.StatObjectOptions{})
	if err != nil {
		return false, nil
	}

	return true, nil
}

func (s *S3Client) GetReport(ctx context.Context, userID int) (*models.UserReport, error) {
	object, err := s.client.GetObject(ctx, s.bucketName, s.fileName(userID), minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	defer object.Close()

	data, err := io.ReadAll(object)
	if err != nil {
		return nil, err
	}

	var report models.UserReport
	return &report, json.Unmarshal(data, &report)
}

func (s *S3Client) StoreReport(ctx context.Context, userID int, report *models.UserReport) error {
	if err := s.ensureBucket(ctx); err != nil {
		return err
	}

	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}

	_, err = s.client.PutObject(ctx, s.bucketName, s.fileName(userID),
		bytes.NewReader(data), int64(len(data)), minio.PutObjectOptions{
			ContentType: "application/json",
	})

	return err
}

func (s *S3Client) GenerateCDNLink(userID int, cdnHost string) string {
	return fmt.Sprintf("http://%s/%s/%s", cdnHost, s.bucketName, s.fileName(userID))
}

func (s *S3Client) ensureBucket(ctx context.Context) error {
	exists, err := s.client.BucketExists(ctx, s.bucketName)
	if err != nil || exists {
		return err
	}
	return s.client.MakeBucket(ctx, s.bucketName, minio.MakeBucketOptions{})
}

func (s *S3Client) fileName(userID int) string {
	return fmt.Sprintf("report_%d.json", userID)
}
