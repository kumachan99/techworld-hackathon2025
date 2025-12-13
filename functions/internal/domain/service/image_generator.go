package service

import (
	"context"

	"github.com/techworld-hackathon/functions/internal/domain/entity"
)

// ImageGenerateResult は画像生成結果
type ImageGenerateResult struct {
	Image string // Base64エンコードされた画像
	Seed  int    // 使用されたシード値
}

// ImageGenerator は街の画像を生成するインターフェース
type ImageGenerator interface {
	// GenerateCityImage は街のパラメータから街の風景画像を生成する
	GenerateCityImage(ctx context.Context, cityParams *entity.CityParams, passedPolicies []*entity.MasterPolicy) (*ImageGenerateResult, error)
}
