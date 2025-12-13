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

## APIテスト

### ヘルスチェック

```bash
curl http://127.0.0.1:8081/health
# => OK
```

### テストデータの投入

#### 1. 思想マスター

```bash
curl -X POST "http://127.0.0.1:8080/v1/projects/demo-project/databases/(default)/documents/master_ideologies?documentId=ideology_environmentalist" \
  -H "Content-Type: application/json" \
  -d '{
    "fields": {
      "id": {"stringValue": "ideology_environmentalist"},
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

#### 2. 政策マスター

```bash
curl -X POST "http://127.0.0.1:8080/v1/projects/demo-project/databases/(default)/documents/master_policies?documentId=policy_001" \
  -H "Content-Type: application/json" \
  -d '{
    "fields": {
      "id": {"stringValue": "policy_001"},
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

#### 3. ルーム作成

```bash
curl -X POST "http://127.0.0.1:8080/v1/projects/demo-project/databases/(default)/documents/rooms?documentId=test-room-001" \
  -H "Content-Type: application/json" \
  -d '{
    "fields": {
      "hostId": {"stringValue": "user_host"},
      "status": {"stringValue": "LOBBY"},
      "turn": {"integerValue": "1"},
      "maxTurns": {"integerValue": "5"},
      "cityParams": {
        "mapValue": {
          "fields": {
            "economy": {"integerValue": "50"},
            "welfare": {"integerValue": "50"},
            "education": {"integerValue": "50"},
            "environment": {"integerValue": "50"},
            "security": {"integerValue": "50"},
            "humanRights": {"integerValue": "50"}
          }
        }
      },
      "isCollapsed": {"booleanValue": false},
      "currentPolicyIds": {"arrayValue": {"values": []}},
      "deckIds": {
        "arrayValue": {
          "values": [
            {"stringValue": "policy_001"},
            {"stringValue": "policy_002"},
            {"stringValue": "policy_003"}
          ]
        }
      },
      "votes": {"mapValue": {"fields": {}}}
    }
  }'
```

#### 4. プレイヤー作成

```bash
curl -X POST "http://127.0.0.1:8080/v1/projects/demo-project/databases/(default)/documents/rooms/test-room-001/players?documentId=user_host" \
  -H "Content-Type: application/json" \
  -d '{
    "fields": {
      "displayName": {"stringValue": "ホスト太郎"},
      "isHost": {"booleanValue": true},
      "isReady": {"booleanValue": true},
      "isPetitionUsed": {"booleanValue": false},
      "currentVote": {"stringValue": ""}
    }
  }'
```

### APIエンドポイントのテスト

#### ゲーム開始 API

```bash
curl -X POST "http://127.0.0.1:8081/api/rooms/test-room-001/start" \
  -H "Content-Type: application/json"
```

レスポンス例:
```json
{
  "status": "VOTING",
  "turn": 1,
  "currentPolicyIds": ["policy_003", "policy_001", "policy_002"]
}
```

#### 投票集計 API

投票データをセットしてから実行:

```bash
# 投票を設定（Firestoreに直接）
curl -X PATCH "http://127.0.0.1:8080/v1/projects/demo-project/databases/(default)/documents/rooms/test-room-001" \
  -H "Content-Type: application/json" \
  -d '{
    "fields": {
      "votes": {
        "mapValue": {
          "fields": {
            "user_host": {"stringValue": "policy_001"}
          }
        }
      }
    }
  }'

# 投票集計を実行
curl -X POST "http://127.0.0.1:8081/api/rooms/test-room-001/resolve" \
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
    "actualEffects": {...},
    "newsFlash": "【速報】政策が可決されました！"
  }
}
```

## Emulator UIでのデータ確認

ブラウザで http://127.0.0.1:4040 を開くと、Firestoreエミュレータに保存されたデータを確認・編集できます。

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
