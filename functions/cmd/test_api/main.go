package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

var baseURL string

func main() {
	baseURL = os.Getenv("API_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8081"
	}
	fmt.Printf("Testing API at: %s\n\n", baseURL)

	client := &http.Client{Timeout: 30 * time.Second}

	// 1. ヘルスチェック
	fmt.Println("=== 1. Health Check ===")
	resp, err := client.Get(baseURL + "/health")
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
	} else {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		fmt.Printf("✅ Status: %d, Body: %s\n", resp.StatusCode, string(body))
	}

	// 2. 部屋作成
	fmt.Println("\n=== 2. Create Room ===")
	createRoomReq := map[string]interface{}{
		"displayName": "テストプレイヤー1",
	}
	roomResp, err := postJSON(client, "/api/rooms", createRoomReq)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		return
	}
	fmt.Printf("✅ Room created: %s\n", prettyJSON(roomResp))

	roomID, _ := roomResp["roomId"].(string)
	player1ID, _ := roomResp["playerId"].(string)
	if roomID == "" {
		fmt.Println("❌ roomId not found in response")
		return
	}
	fmt.Printf("   roomId: %s, playerId: %s\n", roomID, player1ID)

	// 3. 部屋参加（2人目）
	fmt.Println("\n=== 3. Join Room ===")
	joinReq := map[string]interface{}{
		"displayName": "テストプレイヤー2",
	}
	joinResp, err := postJSON(client, fmt.Sprintf("/api/rooms/%s/join", roomID), joinReq)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
	} else {
		fmt.Printf("✅ Joined: %s\n", prettyJSON(joinResp))
	}
	player2ID, _ := joinResp["playerId"].(string)

	// 4. Ready状態トグル（プレイヤー1）- 初期値trueなので2回トグルしてtrueに戻す
	fmt.Println("\n=== 4. Toggle Ready (Player 1) ===")
	readyReq1 := map[string]interface{}{
		"playerId": player1ID,
	}
	// 1回目: true -> false
	postJSON(client, fmt.Sprintf("/api/rooms/%s/ready", roomID), readyReq1)
	// 2回目: false -> true
	readyResp1, err := postJSON(client, fmt.Sprintf("/api/rooms/%s/ready", roomID), readyReq1)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
	} else {
		fmt.Printf("✅ Ready toggled: %s\n", prettyJSON(readyResp1))
	}

	// 5. Ready状態トグル（プレイヤー2）
	fmt.Println("\n=== 5. Toggle Ready (Player 2) ===")
	readyReq2 := map[string]interface{}{
		"playerId": player2ID,
	}
	// 1回目: true -> false
	postJSON(client, fmt.Sprintf("/api/rooms/%s/ready", roomID), readyReq2)
	// 2回目: false -> true
	readyResp2, err := postJSON(client, fmt.Sprintf("/api/rooms/%s/ready", roomID), readyReq2)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
	} else {
		fmt.Printf("✅ Ready toggled: %s\n", prettyJSON(readyResp2))
	}

	// 6. ゲーム開始
	fmt.Println("\n=== 6. Start Game ===")
	startReq := map[string]interface{}{
		"playerId": player1ID,
	}
	startResp, err := postJSON(client, fmt.Sprintf("/api/rooms/%s/start", roomID), startReq)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		fmt.Println("\n=== Test aborted (game not started) ===")
		return
	}
	fmt.Printf("✅ Game started: %s\n", prettyJSON(startResp))

	// currentPolicyIdsを取得（配列）
	var currentPolicyIDs []string
	if policyIDs, ok := startResp["currentPolicyIds"].([]interface{}); ok {
		for _, id := range policyIDs {
			if s, ok := id.(string); ok {
				currentPolicyIDs = append(currentPolicyIDs, s)
			}
		}
	}
	fmt.Printf("   currentPolicyIds: %v\n", currentPolicyIDs)

	if len(currentPolicyIDs) == 0 {
		fmt.Println("⚠️  No currentPolicyIds, skipping vote tests")
	} else {
		// 最初のpolicyIdを使用
		targetPolicyID := currentPolicyIDs[0]

		// 7. 投票（プレイヤー1）
		fmt.Println("\n=== 7. Vote (Player 1) ===")
		voteReq := map[string]interface{}{
			"playerId": player1ID,
			"policyId": targetPolicyID,
			"vote":     true,
		}
		voteResp, err := postJSON(client, fmt.Sprintf("/api/rooms/%s/vote", roomID), voteReq)
		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)
		} else {
			fmt.Printf("✅ Voted for %s: %s\n", targetPolicyID, prettyJSON(voteResp))
		}

		// 8. 投票（プレイヤー2）
		fmt.Println("\n=== 8. Vote (Player 2) ===")
		voteReq2 := map[string]interface{}{
			"playerId": player2ID,
			"policyId": targetPolicyID,
			"vote":     true,
		}
		voteResp2, err := postJSON(client, fmt.Sprintf("/api/rooms/%s/vote", roomID), voteReq2)
		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)
		} else {
			fmt.Printf("✅ Voted for %s: %s\n", targetPolicyID, prettyJSON(voteResp2))
		}
	}

	// 9. 投票集計
	fmt.Println("\n=== 9. Resolve Vote ===")
	resolveReq := map[string]interface{}{}
	resolveResp, err := postJSON(client, fmt.Sprintf("/api/rooms/%s/resolve", roomID), resolveReq)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
	} else {
		fmt.Printf("✅ Resolved: %s\n", prettyJSON(resolveResp))
	}

	// 10. 次ターン
	fmt.Println("\n=== 10. Next Turn ===")
	nextReq := map[string]interface{}{}
	nextResp, err := postJSON(client, fmt.Sprintf("/api/rooms/%s/next", roomID), nextReq)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
	} else {
		fmt.Printf("✅ Next turn: %s\n", prettyJSON(nextResp))
	}

	// 11. 陳情（AI）- オプション
	fmt.Println("\n=== 11. Submit Petition (AI) ===")
	if os.Getenv("SKIP_PETITION") == "true" {
		fmt.Println("⏭️  Skipped (SKIP_PETITION=true)")
	} else {
		petitionReq := map[string]interface{}{
			"playerId": player1ID,
			"text":     "学校への支援を増やしてほしいです",
		}
		petitionResp, err := postJSON(client, fmt.Sprintf("/api/rooms/%s/petition", roomID), petitionReq)
		if err != nil {
			fmt.Printf("❌ Error: %v\n", err)
		} else {
			fmt.Printf("✅ Petition: %s\n", prettyJSON(petitionResp))
		}
	}

	// 12. 退出
	fmt.Println("\n=== 12. Leave Room ===")
	leaveReq := map[string]interface{}{
		"playerId": player2ID,
	}
	leaveResp, err := postJSON(client, fmt.Sprintf("/api/rooms/%s/leave", roomID), leaveReq)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
	} else {
		fmt.Printf("✅ Left: %s\n", prettyJSON(leaveResp))
	}

	fmt.Println("\n=== Test Complete ===")
}

func postJSON(client *http.Client, path string, data map[string]interface{}) (map[string]interface{}, error) {
	body, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", baseURL+path, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("status %d: %s", resp.StatusCode, string(respBody))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %s", string(respBody))
	}

	return result, nil
}

func prettyJSON(data map[string]interface{}) string {
	b, err := json.MarshalIndent(data, "   ", "  ")
	if err != nil {
		return fmt.Sprintf("%v", data)
	}
	return string(b)
}
