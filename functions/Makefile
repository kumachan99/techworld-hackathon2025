.PHONY: build run test clean deploy

# 依存関係のダウンロード
deps:
	go mod tidy

# ビルド
build:
	go build -o bin/server ./cmd/main.go

# ローカル実行
run:
	OPENAI_API_KEY=$(OPENAI_API_KEY) go run ./cmd/main.go

# テスト
test:
	go test -v ./...

# クリーン
clean:
	rm -rf bin/

# Docker ビルド
docker-build:
	docker build -t game-api .

# Docker 実行
docker-run:
	docker run -p 8080:8080 \
		-e OPENAI_API_KEY=$(OPENAI_API_KEY) \
		-e GOOGLE_APPLICATION_CREDENTIALS=/app/credentials.json \
		-v $(PWD)/credentials.json:/app/credentials.json \
		game-api

# Cloud Run へデプロイ
deploy:
	gcloud run deploy game-api \
		--source . \
		--region asia-northeast1 \
		--allow-unauthenticated \
		--set-env-vars OPENAI_API_KEY=$(OPENAI_API_KEY)
