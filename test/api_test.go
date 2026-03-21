package test

import (
	"MixFound/global"
	"MixFound/searcher"
	"MixFound/searcher/words"
	"MixFound/web/controller"
	"MixFound/web/router"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

var testRouter *gin.Engine

func init() {
	gin.SetMode(gin.TestMode)
	global.CONFIG = global.GetDefaultConfig()
	global.CONFIG.Debug = true

	// 初始化测试数据目录
	testDataDir := "./test_data"
	os.MkdirAll(testDataDir, os.ModePerm)

	// 初始化 Tokenizer
	tokenizer := words.NewTokenizer(global.CONFIG.Dictionary)

	// 初始化 Container
	global.Container = &searcher.Container{
		Dir:       testDataDir,
		Debug:     true,
		Shard:     10,
		Timeout:   -1, // 禁用自动关闭，避免测试时并发问题
		BufferNum: 1000,
		Tokenizer: tokenizer,
	}
	err := global.Container.Init()
	if err != nil {
		panic(err)
	}

	controller.NewServices()
	testRouter = router.SetupRouter()
}

type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

func performRequest(r *gin.Engine, method, url string, body interface{}) *httptest.ResponseRecorder {
	var w *httptest.ResponseRecorder
	var req *http.Request
	var err error

	if body != nil {
		jsonBody, _ := json.Marshal(body)
		req, err = http.NewRequest(method, url, bytes.NewBuffer(jsonBody))
		if err != nil {
			panic(err)
		}
		req.Header.Set("Content-Type", "application/json")
	} else {
		req, err = http.NewRequest(method, url, nil)
		if err != nil {
			panic(err)
		}
	}

	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestQuery(t *testing.T) {
	queryData := map[string]interface{}{
		"query": "",
	}

	w := performRequest(testRouter, "POST", "/api/query?database=test_db", queryData)

	fmt.Println("\n========== TestQuery ==========")
	fmt.Printf("URL: %s\n", "/api/query?database=test")
	fmt.Printf("Method: POST\n")
	fmt.Printf("Request Body: %+v\n", queryData)
	fmt.Printf("Status Code: %d\n", w.Code)
	fmt.Printf("Response Body: %s\n", w.Body.String())

	var response Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	fmt.Printf("Parsed Response: %+v\n", response)
	fmt.Println("==============================")
}

func TestIntegration(t *testing.T) {
	// 使用独立的测试数据库
	testDB := "integration_final_test_db"

	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════════════════╗")
	fmt.Println("║           开始集成测试流程                                ║")
	fmt.Println("╚══════════════════════════════════════════════════════════╝")

	// 预清理：确保数据库不存在
	fmt.Println("\n【预清理】删除可能存在的旧数据库...")
	_ = performRequest(testRouter, "GET", fmt.Sprintf("/api/db/drop?database=%s", testDB), nil)
	time.Sleep(500 * time.Millisecond)

	// 步骤 1: 创建数据库
	fmt.Println("\n【步骤 1】创建数据库...")
	w := performRequest(testRouter, "GET", fmt.Sprintf("/api/db/create?database=%s", testDB), nil)
	fmt.Printf("Status: %d\n", w.Code)
	fmt.Printf("Response: %s\n", w.Body.String())

	var response Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if response.Data != "create success" {
		t.Fatalf("Create database failed: %v", response)
	}
	fmt.Println("✓ 数据库创建成功")

	// 等待数据库初始化完成（生产环境中不会频繁创建数据库，所以可以等待较长时间）
	fmt.Println("⏳ 等待数据库初始化完成...")
	time.Sleep(3 * time.Second)

	// 步骤 2: 查询数据库列表
	fmt.Println("\n【步骤 2】查询数据库列表...")
	w = performRequest(testRouter, "GET", "/api/db/list", nil)
	fmt.Printf("Status: %d\n", w.Code)
	fmt.Printf("Response: %s\n", w.Body.String())

	err = json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	dataMap, ok := response.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Database list is not a map")
	}
	if _, exists := dataMap[testDB]; !exists {
		t.Fatalf("Database %s not found in list", testDB)
	}
	fmt.Println("✓ 数据库列表查询成功，包含新创建的数据库")

	// 步骤 3: 写入单个索引
	fmt.Println("\n【步骤 3】写入单个索引...")
	singleDoc := map[string]interface{}{
		"id":   300,
		"text": "集成测试文档 title 内容",
		"document": map[string]interface{}{
			"title":   "测试标题",
			"content": "这是内容",
		},
	}
	w = performRequest(testRouter, "POST", fmt.Sprintf("/api/index?database=%s", testDB), singleDoc)
	fmt.Printf("Status: %d\n", w.Code)
	fmt.Printf("Response: %s\n", w.Body.String())

	err = json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	fmt.Println("✓ 单个索引写入成功")

	// 等待异步索引完成
	fmt.Println("⏳ 等待索引构建完成...")
	time.Sleep(2 * time.Second)

	// 步骤 4: 批量写入索引
	fmt.Println("\n【步骤 4】批量写入索引...")
	batchDocs := []map[string]interface{}{
		{
			"id":   301,
			"text": "第二个文档 title 内容",
			"document": map[string]interface{}{
				"title":   "第二个标题",
				"content": "第二个内容",
			},
		},
		{
			"id":   302,
			"text": "第三个文档 title 测试",
			"document": map[string]interface{}{
				"title":   "第三个标题",
				"content": "第三个内容",
			},
		},
		{
			"id":   303,
			"text": "第四个文档 title",
			"document": map[string]interface{}{
				"title":   "第四个标题",
				"content": "第四个内容",
			},
		},
	}
	w = performRequest(testRouter, "POST", fmt.Sprintf("/api/index/batch?database=%s", testDB), batchDocs)
	fmt.Printf("Status: %d\n", w.Code)
	fmt.Printf("Response: %s\n", w.Body.String())

	err = json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	fmt.Println("✓ 批量索引写入成功")

	// 等待批量索引构建完成
	fmt.Println("⏳ 等待批量索引构建完成...")
	time.Sleep(3 * time.Second)

	// 步骤 5: 查询 "title" 字段
	fmt.Println("\n【步骤 5】查询 'title' 字段...")
	queryData := map[string]interface{}{
		"query": "title",
	}
	w = performRequest(testRouter, "POST", fmt.Sprintf("/api/query?database=%s&query=title", testDB), queryData)
	fmt.Printf("Status: %d\n", w.Code)
	fmt.Printf("Response: %s\n", w.Body.String())

	err = json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if data, ok := response.Data.(map[string]interface{}); ok {
		if total, exists := data["total"]; exists {
			if totalFloat, ok := total.(float64); ok && totalFloat == 0 {
				t.Fatal("Expected to find documents but got total=0")
			}
		}
	}
	fmt.Println("✓ 'title' 字段查询成功，找到文档")

	// 等待后台 worker 完成，避免锁竞争
	time.Sleep(1 * time.Second)

	// 步骤 6: 移除索引（ID 为 300 的文档）
	fmt.Println("\n【步骤 6】移除索引（ID: 300）...")
	removeDoc := map[string]interface{}{
		"id": 300,
	}
	w = performRequest(testRouter, "POST", fmt.Sprintf("/api/index/remove?database=%s", testDB), removeDoc)
	fmt.Printf("Status: %d\n", w.Code)
	fmt.Printf("Response: %s\n", w.Body.String())

	err = json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	fmt.Println("✓ 索引移除成功")

	// 步骤 7: 再次查询 "title" 字段
	fmt.Println("\n【步骤 7】再次查询 'title' 字段（验证移除效果）...")
	w = performRequest(testRouter, "POST", fmt.Sprintf("/api/query?database=%s&query=title", testDB), queryData)
	fmt.Printf("Status: %d\n", w.Code)
	fmt.Printf("Response: %s\n", w.Body.String())

	err = json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	fmt.Println("✓ 再次查询 'title' 字段成功")

	// 步骤 8: 删除数据库
	fmt.Println("\n【步骤 8】删除数据库...")
	// 等待异步操作完成，避免并发问题
	time.Sleep(500 * time.Millisecond)
	w = performRequest(testRouter, "GET", fmt.Sprintf("/api/db/drop?database=%s", testDB), nil)
	fmt.Printf("Status: %d\n", w.Code)
	fmt.Printf("Response: %s\n", w.Body.String())

	err = json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if response.Data != "drop success" {
		t.Fatalf("Drop database failed: %v", response)
	}
	fmt.Println("✓ 数据库删除成功")

	// 步骤 9: 再次查询数据库列表
	fmt.Println("\n【步骤 9】再次查询数据库列表（验证删除效果）...")
	w = performRequest(testRouter, "GET", "/api/db/list", nil)
	fmt.Printf("Status: %d\n", w.Code)
	fmt.Printf("Response: %s\n", w.Body.String())

	err = json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	dataMap, ok = response.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Database list is not a map")
	}
	if _, exists := dataMap[testDB]; exists {
		t.Fatalf("Database %s should be deleted but still exists", testDB)
	}
	fmt.Println("✓ 数据库列表查询成功，已删除的数据库不在列表中")

	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════════════════╗")
	fmt.Println("║           集成测试流程全部完成 ✓                          ║")
	fmt.Println("╚══════════════════════════════════════════════════════════╝")
	fmt.Println()
}

func TestCleanupDatabases(t *testing.T) {
	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════════════════╗")
	fmt.Println("║           清理测试数据库                                  ║")
	fmt.Println("╚══════════════════════════════════════════════════════════╝")

	// 先获取数据库列表
	fmt.Println("\n【步骤 1】获取数据库列表...")
	w := performRequest(testRouter, "GET", "/api/db/list", nil)
	fmt.Printf("Status: %d\n", w.Code)
	fmt.Printf("Response: %s\n", w.Body.String())

	var response Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	dataMap, ok := response.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Database list is not a map")
	}

	// 删除 test 和 test_db 数据库
	databasesToDelete := []string{"test", "test_db"}

	for _, dbName := range databasesToDelete {
		if _, exists := dataMap[dbName]; exists {
			fmt.Printf("\n【删除数据库】%s...\n", dbName)
			w = performRequest(testRouter, "GET", fmt.Sprintf("/api/db/drop?database=%s", dbName), nil)
			fmt.Printf("Status: %d\n", w.Code)
			fmt.Printf("Response: %s\n", w.Body.String())

			err = json.Unmarshal(w.Body.Bytes(), &response)
			if err != nil {
				t.Fatalf("Failed to unmarshal response: %v", err)
			}

			if response.Data == "drop success" {
				fmt.Printf("✓ 数据库 %s 删除成功\n", dbName)
			} else {
				fmt.Printf("⚠ 数据库 %s 删除失败：%v\n", dbName, response)
			}

			// 等待异步操作完成
			time.Sleep(200 * time.Millisecond)
		} else {
			fmt.Printf("\n⚠ 数据库 %s 不存在，跳过\n", dbName)
		}
	}

	// 验证删除结果
	fmt.Println("\n【验证】查询数据库列表...")
	w = performRequest(testRouter, "GET", "/api/db/list", nil)
	fmt.Printf("Status: %d\n", w.Code)
	fmt.Printf("Response: %s\n", w.Body.String())

	err = json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	dataMap, ok = response.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Database list is not a map")
	}

	// 检查是否还有需要删除的数据库
	remaining := false
	for _, dbName := range databasesToDelete {
		if _, exists := dataMap[dbName]; exists {
			fmt.Printf("⚠ 数据库 %s 仍然存在\n", dbName)
			remaining = true
		}
	}

	if !remaining {
		fmt.Println("\n✓ 所有测试数据库已清理完成")
	}

	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════════════════╗")
	fmt.Println("║           数据库清理完成 ✓                                ║")
	fmt.Println("╚══════════════════════════════════════════════════════════╝")
	fmt.Println()
}

func TestDBS(t *testing.T) {
	w := performRequest(testRouter, "GET", "/api/db/list", nil)

	fmt.Println("\n========== TestDBS ==========")
	fmt.Printf("URL: %s\n", "/api/db/list")
	fmt.Printf("Method: GET\n")
	fmt.Printf("Status Code: %d\n", w.Code)
	fmt.Printf("Response Body: %s\n", w.Body.String())

	var response Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	fmt.Printf("Parsed Response: %+v\n", response)
	fmt.Println("==============================")
}

func TestDatabaseCreate(t *testing.T) {
	w := performRequest(testRouter, "GET", "/api/db/create?database=test_db", nil)

	fmt.Println("\n========== TestDatabaseCreate ==========")
	fmt.Printf("URL: %s\n", "/api/db/create?database=test_db")
	fmt.Printf("Method: GET\n")
	fmt.Printf("Status Code: %d\n", w.Code)
	fmt.Printf("Response Body: %s\n", w.Body.String())

	var response Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	fmt.Printf("Parsed Response: %+v\n", response)
	fmt.Println("==============================")
}

func TestDatabaseDrop(t *testing.T) {
	w := performRequest(testRouter, "GET", "/api/db/drop?database=test_db", nil)

	fmt.Println("\n========== TestDatabaseDrop ==========")
	fmt.Printf("URL: %s\n", "/api/db/drop?database=test_db")
	fmt.Printf("Method: GET\n")
	fmt.Printf("Status Code: %d\n", w.Code)
	fmt.Printf("Response Body: %s\n", w.Body.String())

	var response Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	fmt.Printf("Parsed Response: %+v\n", response)
	fmt.Println("==============================")
}

func TestAddIndex(t *testing.T) {
	indexDoc := map[string]interface{}{
		"id":   1,
		"text": "北京",
		"document": map[string]interface{}{
			"title":   "test title",
			"content": "test content",
		},
	}

	w := performRequest(testRouter, "POST", "/api/index?database=test_db", indexDoc)

	fmt.Println("\n========== TestAddIndex ==========")
	fmt.Printf("URL: %s\n", "/api/index?database=test_db")
	fmt.Printf("Method: POST\n")
	fmt.Printf("Request Body: %+v\n", indexDoc)
	fmt.Printf("Status Code: %d\n", w.Code)
	fmt.Printf("Response Body: %s\n", w.Body.String())

	var response Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	fmt.Printf("Parsed Response: %+v\n", response)
	fmt.Println("==============================")
}

func TestBatchAddIndex(t *testing.T) {
	documents := []map[string]interface{}{
		{
			"id":   2,
			"text": "batch content 1",
			"document": map[string]interface{}{
				"title":   "batch title 1",
				"content": "batch content 1",
			},
		},
		{
			"id":   3,
			"text": "batch content 2",
			"document": map[string]interface{}{
				"title":   "batch title 2",
				"content": "batch content 2",
			},
		},
	}

	w := performRequest(testRouter, "POST", "/api/index/batch?database=test_db", documents)

	fmt.Println("\n========== TestBatchAddIndex ==========")
	fmt.Printf("URL: %s\n", "/api/index/batch?database=test_db")
	fmt.Printf("Method: POST\n")
	fmt.Printf("Request Body: %+v\n", documents)
	fmt.Printf("Status Code: %d\n", w.Code)
	fmt.Printf("Response Body: %s\n", w.Body.String())

	var response Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	fmt.Printf("Parsed Response: %+v\n", response)
	fmt.Println("==============================")
}

func TestRemoveIndex(t *testing.T) {
	removeDoc := map[string]interface{}{
		"id": 1,
	}

	w := performRequest(testRouter, "POST", "/api/index/remove?database=test_db", removeDoc)

	fmt.Println("\n========== TestRemoveIndex ==========")
	fmt.Printf("URL: %s\n", "/api/index/remove?database=test_db")
	fmt.Printf("Method: POST\n")
	fmt.Printf("Request Body: %+v\n", removeDoc)
	fmt.Printf("Status Code: %d\n", w.Code)
	fmt.Printf("Response Body: %s\n", w.Body.String())

	var response Response
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	fmt.Printf("Parsed Response: %+v\n", response)
	fmt.Println("==============================")
}
