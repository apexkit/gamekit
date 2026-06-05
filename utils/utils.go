package utils

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/apexkit/gamekit/models"
	"gorm.io/gorm"
)

// AppIdCache 用于缓存通过 AppId 查询的 AppInfo
type AppIdCache struct {
	cache map[string]*models.AppInfo
	mu    sync.RWMutex
}

// 初始化 AppId 缓存
var appIdCache = &AppIdCache{
	cache: make(map[string]*models.AppInfo),
}

// GetAppInfoByAppId 通过 AppId 获取 AppInfo，使用 AppId 缓存
func GetAppInfoByAppId(db *gorm.DB, appId string) (*models.AppInfo, error) {
	// 先尝试从 AppId 缓存读取
	appIdCache.mu.RLock()
	appInfo, exists := appIdCache.cache[appId]
	appIdCache.mu.RUnlock()
	if exists {
		return appInfo, nil
	}

	// 缓存中不存在，从数据库查询
	var dbAppInfo models.AppInfo
	result := db.Where("app_id = ?", appId).First(&dbAppInfo)
	if result.Error != nil {
		return nil, result.Error
	}

	// 将查询结果写入 AppId 缓存
	appIdCache.mu.Lock()
	appIdCache.cache[appId] = &dbAppInfo
	appIdCache.mu.Unlock()

	return &dbAppInfo, nil
}

// DeleteByAppId 按 AppId 指定清除 AppIdCache 中的缓存
func (c *AppIdCache) DeleteByAppId(appId string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.cache, appId)
}

// ===========================================================================
// generateNonce 生成指定长度的随机字符串

func generateNonce(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.IntN(len(charset))] // 注意：v2 中方法名是 IntN（大写 N）
	}
	return string(b)

}

// generateSign: HMAC-SHA256(appId + nonce + timestamp + body, accessSecret)
func generateSign(appId, accessSecret, nonce string, timestamp int64, body string) string {
	parameter := appId + nonce + strconv.FormatInt(timestamp, 10) + body
	mac := hmac.New(sha256.New, []byte(accessSecret))
	mac.Write([]byte(parameter))
	return hex.EncodeToString(mac.Sum(nil))
}

func SendRequest(appId, accessSecret, apiURL string, data []byte) ([]byte, error) {
	nonce := generateNonce(8)
	timestamp := time.Now().Unix()
	timestampStr := strconv.FormatInt(timestamp, 10)
	signature := generateSign(appId, accessSecret, nonce, timestamp, string(data))

	client := &http.Client{}
	req, err := http.NewRequest("POST", apiURL, bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("Error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("x-appid", appId)
	req.Header.Set("x-sign", signature)
	req.Header.Set("x-nonce", nonce)
	req.Header.Set("x-timestamp", timestampStr)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Error sending request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Error reading response body: %w", err)
	}

	return body, nil
}

// ==========================================================================
func GetRespData(respData []byte) ([]byte, error) {

	// 定义通用的响应结构
	var response map[string]interface{}
	err := json.Unmarshal(respData, &response)
	if err != nil {
		return nil, err
	}

	// 提取 resp_msg 部分
	respMsg, ok := response["resp_msg"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("resp_msg 格式不正确")
	}

	// 获取 code 和 message
	code, ok := respMsg["code"].(float64)
	if !ok {
		return nil, fmt.Errorf("resp_msg 中的 code 格式不正确")
	}
	message, ok := respMsg["message"].(string)
	if !ok {
		return nil, fmt.Errorf("resp_msg 中的 message 格式不正确")
	}

	// 根据 code 和 message 进行判断
	if code != 200 {
		return nil, fmt.Errorf("请求失败，code: %d, message: %s", int(code), message)
	}

	// 提取 resp_data 部分
	respDataMap, ok := response["resp_data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("resp_data 格式不正确")
	}

	// 将 resp_data 转换为 JSON 字节
	respDataJSON, err := json.Marshal(respDataMap)
	if err != nil {
		return nil, err
	}
	return respDataJSON, err
}

type OperationApiError struct {
	Code int    // 业务错误码
	Msg  string // 自定义错误信息（覆盖默认）
}

func (e *OperationApiError) Error() string {
	if e.Msg != "" {
		return e.Msg
	}
	return "unknown error"
}

// ==========================================================================
func GetApiRespData(respData []byte) ([]byte, error) {

	// 定义通用的响应结构
	var response map[string]interface{}
	err := json.Unmarshal(respData, &response)
	if err != nil {
		return nil, err
	}

	// 获取 code 和 message
	code, ok := response["code"].(float64)
	if !ok {
		return nil, fmt.Errorf("resp_msg 中的 code 格式不正确")
	}
	message, ok := response["msg"].(string)
	if !ok {
		return nil, fmt.Errorf("resp_msg 中的 message 格式不正确")
	}

	// 根据 code 和 message 进行判断
	if code != 0 {
		return nil, &OperationApiError{Code: int(code), Msg: message}
	}

	// 提取 resp_data 部分
	respDataMap, ok := response["data"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("resp_data 格式不正确")
	}

	// 将 resp_data 转换为 JSON 字节
	respDataJSON, err := json.Marshal(respDataMap)
	if err != nil {
		return nil, err
	}
	return respDataJSON, err
}
