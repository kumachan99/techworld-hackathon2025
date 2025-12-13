# ローカル開発ガイド

## 前提条件

- Go 1.21以上
- Node.js 18以上
- Java（Firestoreエミュレータ用）

### Javaのインストール（未インストールの場合）

```bash
brew install openjdk
```

## ローカル環境の起動

### 1. Firestoreエミュレータの起動

```bash
# JAVA_HOMEを設定してエミュレータを起動
export JAVA_HOME=/opt/homebrew/opt/openjdk/libexec/openjdk.jdk/Contents/Home
npx firebase-tools emulators:start --project demo-project
```

起動後のURL:
| サービス | URL |
|---------|-----|
| Firestore | localhost:8080 |
| Emulator UI | http://127.0.0.1:4040 |
| Hosting | http://127.0.0.1:5050 |

### 2. Goサーバーの起動（別ターミナル）

```bash
cd functions
FIRESTORE_EMULATOR_HOST=127.0.0.1:8080 GOOGLE_CLOUD_PROJECT=demo-project go run ./cmd/
```

起動後のURL:
| サービス | URL |
|---------|-----|
| Go API | http://127.0.0.1:8081 |
| Health Check | http://127.0.0.1:8081/health |

## マスターデータの投入

### seedスクリプトを使用

```bash
cd scripts
FIRESTORE_EMULATOR_HOST=127.0.0.1:8080 GOOGLE_CLOUD_PROJECT=demo-project go run seed.go
```

### 手動投入（curl）

#### 思想マスター

```bash
curl -X POST "http://127.0.0.1:8080/v1/projects/demo-project/databases/(default)/documents/master_ideologies?documentId=ideology_environmentalist" \
  -H "Content-Type: application/json" \
  -d '{
    "fields": {
      "ideologyId": {"stringValue": "ideology_environmentalist"},
      "name": {"stringValue": "環境主義者"},
      "description": {"stringValue": "自然環境の保護を最優先とする思想"},
      "coefficients": {
        "mapValue": {
          "fields": {
            "economy": {"doubleValue": 0.5},
            "welfare": {"doubleValue": 1.0},
            "education": {"doubleValue": 1.0},
            "environment": {"doubleValue": 2.0},
            "security": {"doubleValue": 0.5},
            "humanRights": {"doubleValue": 1.0}
          }
        }
      }
    }
  }'
```

#### 政策マスター

```bash
curl -X POST "http://127.0.0.1:8080/v1/projects/demo-project/databases/(default)/documents/master_policies?documentId=policy_001" \
  -H "Content-Type: application/json" \
  -d '{
    "fields": {
      "policyId": {"stringValue": "policy_001"},
      "title": {"stringValue": "経済政策1"},
      "description": {"stringValue": "テスト政策の説明"},
      "newsFlash": {"stringValue": "【速報】政策が可決されました！"},
      "effects": {
        "mapValue": {
          "fields": {
            "economy": {"integerValue": "10"},
            "welfare": {"integerValue": "-5"},
            "education": {"integerValue": "0"},
            "environment": {"integerValue": "-3"},
            "security": {"integerValue": "5"},
            "humanRights": {"integerValue": "0"}
          }
        }
      }
    }
  }'
```

---

## APIテスト

### ヘルスチェック

```bash
curl http://127.0.0.1:8081/health
# => OK
```

### 部屋作成 API

```bash
# playerIdはバックエンドで自動生成される
curl -X POST "http://127.0.0.1:8081/api/rooms" \
  -H "Content-Type: application/json" \
  -d '{
    "displayName": "ホスト太郎"
  }'
```

レスポンス例:
```json
{
  "roomId": "abc123xyz",
  "status": "LOBBY",
  "playerId": "550e8400-e29b-41d4-a716-446655440000"
}
```

### 部屋参加 API

```bash
# playerIdはバックエンドで自動生成される
curl -X POST "http://127.0.0.1:8081/api/rooms/{roomId}/join" \
  -H "Content-Type: application/json" \
  -d '{
    "displayName": "プレイヤー花子"
  }'
```

レスポンス例:
```json
{
  "playerId": "550e8400-e29b-41d4-a716-446655440001"
}
```

### Ready状態トグル API

```bash
curl -X POST "http://127.0.0.1:8081/api/rooms/{roomId}/ready" \
  -H "Content-Type: application/json" \
  -d '{
    "playerId": "user456"
  }'
```

レスポンス例:
```json
{
  "isReady": true
}
```

### ゲーム開始 API

```bash
# ホストのみ実行可能
curl -X POST "http://127.0.0.1:8081/api/rooms/{roomId}/start" \
  -H "Content-Type: application/json" \
  -d '{
    "playerId": "user123"
  }'
```

レスポンス例:
```json
{
  "status": "VOTING",
  "turn": 1,
  "currentPolicyIds": ["policy_003", "policy_001", "policy_002"]
}
```

### 投票 API

```bash
curl -X POST "http://127.0.0.1:8081/api/rooms/{roomId}/vote" \
  -H "Content-Type: application/json" \
  -d '{
    "playerId": "user123",
    "policyId": "policy_001"
  }'
```

レスポンス例:
```json
{
  "success": true
}
```

### 投票集計 API

```bash
# フロントエンドから全員投票完了時に自動でトリガー
curl -X POST "http://127.0.0.1:8081/api/rooms/{roomId}/resolve" \
  -H "Content-Type: application/json"
```

レスポンス例:
```json
{
  "status": "RESULT",
  "isGameOver": false,
  "cityParams": {
    "economy": 60,
    "welfare": 45,
    "education": 50,
    "environment": 47,
    "security": 55,
    "humanRights": 50
  },
  "lastResult": {
    "passedPolicyId": "policy_001",
    "passedPolicyTitle": "経済政策1",
    "actualEffects": { "economy": 10, "welfare": -5, ... },
    "newsFlash": "【速報】政策が可決されました！",
    "voteDetails": { "user1": "policy_001" }
  }
}
```

### 次ターン API

```bash
# フロントエンドから結果確認後に自動でトリガー
curl -X POST "http://127.0.0.1:8081/api/rooms/{roomId}/next" \
  -H "Content-Type: application/json"
```

レスポンス例:
```json
{
  "status": "VOTING",
  "turn": 2
}
```

### AI陳情 API

```bash
curl -X POST "http://127.0.0.1:8081/api/rooms/{roomId}/petition" \
  -H "Content-Type: application/json" \
  -d '{
    "playerId": "user123",
    "text": "週休3日制を導入したい"
  }'
```

レスポンス例:
```json
{
  "approved": true,
  "policyId": "petition_abc123",
  "message": "政策が承認されました"
}
```

---

## プレイヤーIDについて

このシステムでは Firebase Authentication を使用せず、シンプルな UUID ベースの識別を採用しています。

- `playerId` は部屋作成・参加時にバックエンドで自動生成される
- フロントエンドはレスポンスで受け取った `playerId` を localStorage に保存
- 以降のAPIリクエスト（ready, vote, petition等）ではリクエストボディに `playerId` を含める

```typescript
// フロントエンドでのplayerId管理例
// 部屋作成時
const response = await fetch('/api/rooms', {
  method: 'POST',
  body: JSON.stringify({ displayName: 'プレイヤー名' })
});
const { roomId, playerId } = await response.json();
localStorage.setItem('playerId', playerId);

// 以降のリクエスト
const playerId = localStorage.getItem('playerId');
await fetch(`/api/rooms/${roomId}/vote`, {
  method: 'POST',
  body: JSON.stringify({ playerId, policyId: 'policy_001' })
});
```

---

## Emulator UIでのデータ確認

ブラウザで http://127.0.0.1:4040 を開くと、Firestoreエミュレータに保存されたデータを確認・編集できます。

---

## トラブルシューティング

### ポートが使用中の場合

```bash
# 使用中のポートを確認
lsof -i :8080 -i :4040 -i :5050

# プロセスを終了
kill <PID>
```

### Javaが見つからない場合

```bash
# JAVA_HOMEを設定
export JAVA_HOME=/opt/homebrew/opt/openjdk/libexec/openjdk.jdk/Contents/Home
export PATH="$JAVA_HOME/bin:$PATH"

# 確認
java -version
```

### Firestoreへの接続エラー

環境変数が正しく設定されているか確認:

```bash
echo $FIRESTORE_EMULATOR_HOST  # => 127.0.0.1:8080
echo $GOOGLE_CLOUD_PROJECT     # => demo-project
```

