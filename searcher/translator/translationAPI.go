package translator

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
	"time"
)

var httpClient = &http.Client{
	Timeout: 10 * time.Second,
}

// 有道翻译 API 端点
const youdaoAPIURL = "https://dict-trans.youdao.com/translate/enhance"

// 有道翻译请求参数
type YoudaoRequest struct {
	SrcArticle    string `json:"srcArticle"`
	TgtArticle    string `json:"tgtArticle"`
	From          string `json:"from"`
	To            string `json:"to"`
	Product       string `json:"product"`
	AppVersion    string `json:"appVersion"`
	Client        string `json:"client"`
	Mid           string `json:"mid"`
	Vendor        string `json:"vendor"`
	Screen        string `json:"screen"`
	Model         string `json:"model"`
	Imei          string `json:"imei"`
	Network       string `json:"network"`
	Keyfrom       string `json:"keyfrom"`
	Keyid         string `json:"keyid"`
	MysticTime    string `json:"mysticTime"`
	Yduuid        string `json:"yduuid"`
	Abtest        string `json:"abtest"`
	Sign          string `json:"sign"`
	SignSecretKey string `json:"signSecretKey"`
	KeyId         string `json:"keyId"`
	Token         string `json:"token"`
	Source        string `json:"source"`
	PointParam    string `json:"pointParam"`
}

// 有道翻译响应
type YoudaoResponse struct {
	ErrorCode int `json:"errorCode"`
	Data      struct {
		Src    string   `json:"src"`
		Tgt    string   `json:"tgt"`
		SrcTgt []string `json:"srcTgt"`
	} `json:"data"`
}

// CallTranslationAPIToEn 调用有道翻译 API 将中文翻译成英文
func CallTranslationAPIToEn(word string) string {
	translated := callYoudaoTranslate(word, "zh-CHS", "en")
	fmt.Println(translated)
	return translated
}

// CallTranslationAPIToCn 调用有道翻译 API 将英文翻译成中文
func CallTranslationAPIToCn(word string) string {
	return callYoudaoTranslate(word, "en", "zh-CHS")
}

// callYoudaoTranslate 调用有道翻译 API 的通用方法
func callYoudaoTranslate(text, from, to string) string {
	// 构建 multipart/form-data 请求
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// 添加所有表单字段
	_ = writer.WriteField("srcArticle", text)
	_ = writer.WriteField("tgtArticle", "") // 目标文本留空，由服务器翻译
	_ = writer.WriteField("from", from)
	_ = writer.WriteField("to", to)
	_ = writer.WriteField("product", "webfanyi")
	_ = writer.WriteField("appVersion", "12.0.0")
	_ = writer.WriteField("client", "webmain")
	_ = writer.WriteField("mid", "1")
	_ = writer.WriteField("vendor", "web")
	_ = writer.WriteField("screen", "1")
	_ = writer.WriteField("model", "1")
	_ = writer.WriteField("imei", "1")
	_ = writer.WriteField("network", "wifi")
	_ = writer.WriteField("keyfrom", "webfanyi.webmain")
	_ = writer.WriteField("keyid", "translate-webfanyi-webmain")
	_ = writer.WriteField("mysticTime", strconv.FormatInt(time.Now().UnixMilli(), 10))
	_ = writer.WriteField("yduuid", generateUUID())
	_ = writer.WriteField("abtest", "0")
	_ = writer.WriteField("signSecretKey", "BdCYRtHAJxO7yNm9RHwU2JiFISIk62Ts")
	_ = writer.WriteField("keyId", "translate-webfanyi-webmain")
	_ = writer.WriteField("token", "10eda3e853af4394892dbbd897c4680c")
	_ = writer.WriteField("source", "webmain")
	_ = writer.WriteField("sign", generateSign(text, from, to))
	_ = writer.WriteField("pointParam", "abtest,appVersion,client,imei,keyId,keyfrom,keyid,mid,model,mysticTime,network,product,screen,signSecretKey,source,token,vendor,yduuid,key")

	err := writer.Close()
	if err != nil {
		fmt.Printf("关闭 multipart writer 失败：%v\n", err)
		return ""
	}

	// 创建 HTTP 请求
	req, err := http.NewRequest("POST", youdaoAPIURL, body)
	if err != nil {
		fmt.Printf("创建翻译请求失败：%v\n", err)
		return ""
	}

	// 设置请求头
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Origin", "https://fanyi.youdao.com")
	req.Header.Set("Referer", "https://fanyi.youdao.com/")
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "same-site")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/143.0.0.0 Safari/537.36")
	req.Header.Set("sec-ch-ua", `"Microsoft Edge";v="143", "Chromium";v="143", "Not A(Brand";v="24"`)
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("sec-ch-ua-platform", `"macOS"`)

	// 发送请求
	resp, err := httpClient.Do(req)
	if err != nil {
		fmt.Printf("发送翻译请求失败：%v\n", err)
		return ""
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("读取翻译响应失败：%v\n", err)
		return ""
	}

	// 解析响应
	return parseYoudaoResponse(respBody)
}

// parseYoudaoResponse 解析有道翻译 API 的响应
func parseYoudaoResponse(response []byte) string {
	var resp YoudaoResponse
	err := json.Unmarshal(response, &resp)
	if err != nil {
		fmt.Printf("解析翻译响应失败：%v\n", err)
		return ""
	}

	if resp.ErrorCode != 0 {
		fmt.Printf("有道翻译 API 错误码：%d\n", resp.ErrorCode)
		return ""
	}

	// 返回翻译结果
	if resp.Data.Tgt != "" {
		return resp.Data.Tgt
	}

	return ""
}

// generateUUID 生成简单的 UUID（用于 yduuid 字段）
func generateUUID() string {
	// 简单实现，实际应该使用更复杂的 UUID 生成算法
	return fmt.Sprintf("%x", time.Now().UnixNano())
}

// generateSign 生成签名（简化版本，实际可能需要更复杂的逻辑）
func generateSign(text, from, to string) string {
	// 这里使用固定的 sign，实际可能需要根据请求参数动态计算
	// 根据观察，sign 可能是基于 mysticTime、token 等参数计算的
	return "7ae25a2516df2497dbffcf751b7489e6"
}
