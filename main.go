package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func main() {
	bearerToken := os.Getenv("BEARER_TOKEN")
	if bearerToken == "" {
		fmt.Println("Error: BEARER_TOKEN not set!")
		return
	}

	userID := getUserID(bearerToken, "uiryuu_")
	if userID != "" {
		fetchTweets(bearerToken, userID)
	}
}

func getUserID(bearerToken string, username string) string {
	endpoint := fmt.Sprintf("https://api.twitter.com/2/users/by/username/%s", username)
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return ""
	}

	req.Header.Add("Authorization", "Bearer "+bearerToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Error fetching user ID:", err)
		return ""
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Println("Error decoding response:", err)
		return ""
	}

	if user, ok := result["data"].(map[string]interface{}); ok {
		return user["id"].(string)
	}
	return ""
}

func fetchTweets(bearerToken string, userID string) {
	var nextToken string

	// 使用循环进行翻页
	for {
		endpoint := fmt.Sprintf("https://api.twitter.com/2/users/%s/tweets", userID)

		// 如果存在nextToken，则添加到endpoint
		if nextToken != "" {
			endpoint = fmt.Sprintf("%s?pagination_token=%s", endpoint, nextToken)
		}

		req, err := http.NewRequest("GET", endpoint, nil)
		if err != nil {
			fmt.Println("Error creating request:", err)
			return
		}

		req.Header.Add("Authorization", "Bearer "+bearerToken)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			fmt.Println("Error fetching tweets:", err)
			return
		}
		defer resp.Body.Close()

		bodyBytes, _ := io.ReadAll(resp.Body)
		fmt.Println("Response:", string(bodyBytes))
		resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

		var result map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			fmt.Println("Error decoding response:", err)
			return
		}

		if data, ok := result["data"].([]interface{}); ok {
			for _, item := range data {
				if tweet, ok := item.(map[string]interface{}); ok {
					fmt.Println(tweet["text"].(string))
				}
			}
		}

		// 检查是否存在next_token，如果不存在，退出循环
		if meta, ok := result["meta"].(map[string]interface{}); ok {
			if token, exists := meta["next_token"].(string); exists {
				nextToken = token
			} else {
				break
			}
		} else {
			break
		}
	}
}
