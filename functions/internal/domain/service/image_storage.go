package service

import (
	"context"
)

// ImageStorage は画像をストレージに保存するインターフェース
type ImageStorage interface {
	// UploadCityImage は街の画像をアップロードし、signed URLを返す
	// roomID と turn を使ってユニークなパスを生成する
	UploadCityImage(ctx context.Context, roomID string, turn int, imageData []byte) (signedURL string, err error)
}
