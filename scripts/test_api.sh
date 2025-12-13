#!/bin/bash
# API テストスクリプト
# 使用方法: ./scripts/test_api.sh

set -e

API_BASE="http://localhost:8081/api"

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
# 3. Ready状態トグル (ゲスト)
# =====================================
print_step "3. Ready状態トグル (POST /api/rooms/{roomId}/ready)"

RESPONSE=$(curl -s -X POST "${API_BASE}/rooms/${ROOM_ID}/ready" \
    -H "Content-Type: application/json" \
    -d "{\"playerId\": \"${GUEST_ID}\"}")

echo "$RESPONSE" | jq .
print_success "Ready状態をトグルしました"

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
# 6. 投票 (ゲスト)
# =====================================
print_step "6. 投票 - ゲスト (POST /api/rooms/{roomId}/vote)"

RESPONSE=$(curl -s -X POST "${API_BASE}/rooms/${ROOM_ID}/vote" \
    -H "Content-Type: application/json" \
    -d "{\"playerId\": \"${GUEST_ID}\", \"policyId\": \"${FIRST_POLICY}\"}")

echo "$RESPONSE" | jq .
print_success "ゲストが投票しました: policyId=$FIRST_POLICY"

# =====================================
# 7. 投票集計
# =====================================
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
