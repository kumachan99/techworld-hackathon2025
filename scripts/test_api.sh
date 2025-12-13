#!/bin/bash
# API テストスクリプト
# 使用方法: ./scripts/test_api.sh
# 本番環境: API_BASE=https://game-api-936357358706.asia-northeast1.run.app/api SAVE_IMAGE=true ./scripts/test_api.sh

# オプション
# 本番環境: API_BASE=https://game-api-xxxxx.run.app/api ./scripts/test_api.sh
# 画像保存: SAVE_IMAGE=true ./scripts/test_api.sh

set -e

API_BASE="${API_BASE:-http://localhost:8081/api}"
echo "API_BASE: $API_BASE"

# 色付き出力
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

print_step() {
    echo -e "\n${YELLOW}=== $1 ===${NC}"
}

print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

# jq が必要
if ! command -v jq &> /dev/null; then
    echo "jq が必要です。brew install jq でインストールしてください。"
    exit 1
fi

# =====================================
# 1. 部屋作成
# =====================================
print_step "1. 部屋作成 (POST /api/rooms)"

RESPONSE=$(curl -s -X POST "${API_BASE}/rooms" \
    -H "Content-Type: application/json" \
    -d '{"displayName": "Host Player"}')

echo "$RESPONSE" | jq .

ROOM_ID=$(echo "$RESPONSE" | jq -r '.roomId')
HOST_ID=$(echo "$RESPONSE" | jq -r '.playerId')

if [ "$ROOM_ID" != "null" ] && [ -n "$ROOM_ID" ]; then
    print_success "部屋作成成功: roomId=$ROOM_ID, playerId=$HOST_ID"
else
    print_error "部屋作成失敗"
    exit 1
fi

# =====================================
# 2. 部屋参加
# =====================================
print_step "2. 部屋参加 (POST /api/rooms/{roomId}/join)"

RESPONSE=$(curl -s -X POST "${API_BASE}/rooms/${ROOM_ID}/join" \
    -H "Content-Type: application/json" \
    -d '{"displayName": "Guest Player"}')

echo "$RESPONSE" | jq .

GUEST_ID=$(echo "$RESPONSE" | jq -r '.playerId')

if [ "$GUEST_ID" != "null" ] && [ -n "$GUEST_ID" ]; then
    print_success "部屋参加成功: playerId=$GUEST_ID"
else
    print_error "部屋参加失敗"
    exit 1
fi

# =====================================
# 3. Ready状態トグル (両プレイヤー)
# =====================================
print_step "3. Ready状態トグル (POST /api/rooms/{roomId}/ready)"

# ゲストをReadyに
RESPONSE=$(curl -s -X POST "${API_BASE}/rooms/${ROOM_ID}/ready" \
    -H "Content-Type: application/json" \
    -d "{\"playerId\": \"${GUEST_ID}\"}")

IS_READY=$(echo "$RESPONSE" | jq -r '.isReady')
if [ "$IS_READY" != "true" ]; then
    # もう一度トグル
    RESPONSE=$(curl -s -X POST "${API_BASE}/rooms/${ROOM_ID}/ready" \
        -H "Content-Type: application/json" \
        -d "{\"playerId\": \"${GUEST_ID}\"}")
fi
echo "$RESPONSE" | jq .
print_success "ゲストがReadyになりました"

# ホストをReadyに (ホストは自動的にReadyになっていない場合)
RESPONSE=$(curl -s -X POST "${API_BASE}/rooms/${ROOM_ID}/ready" \
    -H "Content-Type: application/json" \
    -d "{\"playerId\": \"${HOST_ID}\"}")

IS_READY=$(echo "$RESPONSE" | jq -r '.isReady')
if [ "$IS_READY" != "true" ]; then
    # もう一度トグル
    RESPONSE=$(curl -s -X POST "${API_BASE}/rooms/${ROOM_ID}/ready" \
        -H "Content-Type: application/json" \
        -d "{\"playerId\": \"${HOST_ID}\"}")
fi
echo "$RESPONSE" | jq .
print_success "ホストがReadyになりました"

# =====================================
# 4. ゲーム開始
# =====================================
print_step "4. ゲーム開始 (POST /api/rooms/{roomId}/start)"

RESPONSE=$(curl -s -X POST "${API_BASE}/rooms/${ROOM_ID}/start" \
    -H "Content-Type: application/json" \
    -d "{\"playerId\": \"${HOST_ID}\"}")

echo "$RESPONSE" | jq .

STATUS=$(echo "$RESPONSE" | jq -r '.status')

if [ "$STATUS" = "VOTING" ]; then
    print_success "ゲーム開始成功: status=$STATUS"
else
    print_error "ゲーム開始失敗: status=$STATUS"
    exit 1
fi

# 現在の政策カードを取得
POLICY_IDS=$(echo "$RESPONSE" | jq -r '.currentPolicyIds[]')
FIRST_POLICY=$(echo "$POLICY_IDS" | head -n1)

echo "現在の政策: $POLICY_IDS"
echo "投票先: $FIRST_POLICY"

# =====================================
# 5. 投票 (ホスト)
# =====================================
print_step "5. 投票 - ホスト (POST /api/rooms/{roomId}/vote)"

RESPONSE=$(curl -s -X POST "${API_BASE}/rooms/${ROOM_ID}/vote" \
    -H "Content-Type: application/json" \
    -d "{\"playerId\": \"${HOST_ID}\", \"policyId\": \"${FIRST_POLICY}\"}")

echo "$RESPONSE" | jq .
print_success "ホストが投票しました: policyId=$FIRST_POLICY"

# =====================================
# 6. 投票 (ゲスト) - 全員投票で自動resolveされ、画像生成される
# =====================================
print_step "6. 投票 - ゲスト (POST /api/rooms/{roomId}/vote)"

RESPONSE=$(curl -s -X POST "${API_BASE}/rooms/${ROOM_ID}/vote" \
    -H "Content-Type: application/json" \
    -d "{\"playerId\": \"${GUEST_ID}\", \"policyId\": \"${FIRST_POLICY}\"}")

# 画像以外を表示
echo "$RESPONSE" | jq 'del(.lastResult.cityImage)'
print_success "ゲストが投票しました: policyId=$FIRST_POLICY"

# 自動resolveされたか確認
IS_RESOLVED=$(echo "$RESPONSE" | jq -r '.isResolved')
if [ "$IS_RESOLVED" = "true" ]; then
    print_success "自動resolveされました"

    # 画像が生成されたか確認
    CITY_IMAGE=$(echo "$RESPONSE" | jq -r '.lastResult.cityImage // empty')
    if [ -n "$CITY_IMAGE" ]; then
        IMAGE_LENGTH=${#CITY_IMAGE}
        print_success "街の画像が生成されました (Base64: ${IMAGE_LENGTH} 文字)"

        # 画像をファイルに保存（オプション）
        if [ "$SAVE_IMAGE" = "true" ]; then
            echo "$CITY_IMAGE" | base64 -d > "city_image_turn1.png"
            print_success "画像を city_image_turn1.png に保存しました"
        fi
    else
        print_error "街の画像が生成されませんでした"
    fi
fi

# =====================================
# 7. 投票集計（自動resolveされていない場合のみ）
# =====================================
if [ "$IS_RESOLVED" != "true" ]; then
    print_step "7. 投票集計 (POST /api/rooms/{roomId}/resolve)"

    RESPONSE=$(curl -s -X POST "${API_BASE}/rooms/${ROOM_ID}/resolve" \
        -H "Content-Type: application/json" \
        -d '{}')

    echo "$RESPONSE" | jq .

    STATUS=$(echo "$RESPONSE" | jq -r '.status')
    PASSED_POLICY=$(echo "$RESPONSE" | jq -r '.lastResult.passedPolicyTitle')

    if [ "$STATUS" = "RESULT" ] || [ "$STATUS" = "FINISHED" ]; then
        print_success "投票集計成功: 可決=$PASSED_POLICY"
    else
        print_error "投票集計失敗"
        exit 1
    fi
else
    print_step "7. 投票集計 (スキップ: 自動resolveされました)"
    STATUS=$(echo "$RESPONSE" | jq -r '.status')
fi

# =====================================
# 8. 次ターンへ (ゲーム継続の場合)
# =====================================
if [ "$STATUS" = "RESULT" ]; then
    print_step "8. 次ターンへ (POST /api/rooms/{roomId}/next)"

    RESPONSE=$(curl -s -X POST "${API_BASE}/rooms/${ROOM_ID}/next" \
        -H "Content-Type: application/json" \
        -d '{}')

    echo "$RESPONSE" | jq .

    STATUS=$(echo "$RESPONSE" | jq -r '.status')
    TURN=$(echo "$RESPONSE" | jq -r '.turn')

    if [ "$STATUS" = "VOTING" ]; then
        print_success "次ターン開始: turn=$TURN"
    else
        print_error "次ターン開始失敗"
    fi
fi

# =====================================
# 完了
# =====================================
echo -e "\n${GREEN}=== テスト完了 ===${NC}"
echo "Room ID: $ROOM_ID"
echo "Host ID: $HOST_ID"
echo "Guest ID: $GUEST_ID"
