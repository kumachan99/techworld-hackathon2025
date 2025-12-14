package storage

import (
	"context"
	"fmt"
	"os"
	"time"

	"cloud.google.com/go/storage"

	"github.com/techworld-hackathon/functions/internal/domain/service"
)

const (
	defaultSignedURLExpiry = 7 * 24 * time.Hour // 7日間
)

// GCSClient は Google Cloud Storage クライアント
type GCSClient struct {
	client     *storage.Client
	bucketName string
}

// NewGCSClient は GCSClient を作成する
func NewGCSClient(ctx context.Context, bucketName string) (*GCSClient, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCS client: %w", err)
	}

	return &GCSClient{
		client:     client,
		bucketName: bucketName,
	}, nil
}

// NewGCSClientFromEnv は環境変数からバケット名を取得して GCSClient を作成する
func NewGCSClientFromEnv(ctx context.Context) (*GCSClient, error) {
	bucketName := os.Getenv("GCS_BUCKET_NAME")
	if bucketName == "" {
		return nil, fmt.Errorf("GCS_BUCKET_NAME environment variable is not set")
	}
	return NewGCSClient(ctx, bucketName)
}

// インターフェースの実装を保証
var _ service.ImageStorage = (*GCSClient)(nil)

// UploadCityImage は街の画像をGCSにアップロードし、signed URLを返す
func (c *GCSClient) UploadCityImage(ctx context.Context, roomID string, turn int, imageData []byte) (string, error) {
	// オブジェクトパスを生成: city_images/{roomID}/turn_{turn}.png
	objectPath := fmt.Sprintf("city_images/%s/turn_%d.png", roomID, turn)

	// バケットとオブジェクトへの参照を取得
	bucket := c.client.Bucket(c.bucketName)
	obj := bucket.Object(objectPath)

	// 画像をアップロード
	writer := obj.NewWriter(ctx)
	writer.ContentType = "image/png"
	writer.CacheControl = "public, max-age=604800" // 7日間キャッシュ

	if _, err := writer.Write(imageData); err != nil {
		writer.Close()
		return "", fmt.Errorf("failed to write image data: %w", err)
	}

	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("failed to close writer: %w", err)
	}

	// Signed URL を生成
	opts := &storage.SignedURLOptions{
		Scheme:  storage.SigningSchemeV4,
		Method:  "GET",
		Expires: time.Now().Add(defaultSignedURLExpiry),
	}

	signedURL, err := bucket.SignedURL(objectPath, opts)
	if err != nil {
		return "", fmt.Errorf("failed to generate signed URL: %w", err)
	}

	return signedURL, nil
}

// Close はGCSクライアントを閉じる
func (c *GCSClient) Close() error {
	return c.client.Close()
}
