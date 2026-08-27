package client

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/laixhe/gonet/network/header"
)

// Request HTTP 请求, 通过 Client.Get 等创建后链式构建
type Request struct {
	c           *Client
	err         error
	Method      string
	URL         *url.URL
	QueryParams url.Values
	Header      http.Header
	Result      any

	ctx         context.Context
	body        []byte
	bodyReader  io.Reader
	contentType string
	basicAuth   string
	bearerToken string
	formData    url.Values
	files       []filePart

	// 重试配置, 0/未设置表示继承客户端配置
	retryCount   int
	retryWait    time.Duration
	retryMaxWait time.Duration
	retryConds   []RetryCondition
	// 非 2xx 状态码校验开关, checkStatusSet 标记是否显式设置
	checkStatus    bool
	checkStatusSet bool
	// 响应体读取上限, 0 表示继承客户端配置
	maxBodySize int64
	// 非 2xx 响应的解码目标, 需实现 error 接口
	errTarget any
}

// filePart multipart 文件
type filePart struct {
	param    string
	filename string
	data     []byte
	reader   io.Reader
}

// Context 返回请求上下文, 未设置时返回 context.Background
func (r *Request) Context() context.Context {
	if r.ctx == nil {
		return context.Background()
	}
	return r.ctx
}

// SetContext 设置请求上下文, 支持取消与超时
func (r *Request) SetContext(ctx context.Context) *Request {
	if ctx != nil {
		r.ctx = ctx
	}
	return r
}

// SetQueryParam 设置单个查询参数
func (r *Request) SetQueryParam(param, value string) *Request {
	r.QueryParams.Set(param, value)
	return r
}

// SetQueryParams 批量设置查询参数
func (r *Request) SetQueryParams(params map[string]string) *Request {
	for k, v := range params {
		r.SetQueryParam(k, v)
	}
	return r
}

// SetHeader 设置单个请求头
func (r *Request) SetHeader(k, v string) *Request {
	r.Header.Set(k, v)
	return r
}

// SetHeaders 批量设置请求头
func (r *Request) SetHeaders(headers map[string]string) *Request {
	for k, v := range headers {
		r.Header.Set(k, v)
	}
	return r
}

// SetContentType 设置 Content-Type
func (r *Request) SetContentType(contentType string) *Request {
	r.contentType = contentType
	return r
}

// SetUserAgent 设置 User-Agent, 覆盖默认值
func (r *Request) SetUserAgent(ua string) *Request {
	r.Header.Set(header.UserAgent, ua)
	return r
}

// SetBasicAuth 设置 Basic 认证
func (r *Request) SetBasicAuth(username, password string) *Request {
	r.basicAuth = "Basic " + base64.StdEncoding.EncodeToString([]byte(username+":"+password))
	return r
}

// SetBearerToken 设置 Bearer Token 认证
func (r *Request) SetBearerToken(token string) *Request {
	r.bearerToken = token
	return r
}

// SetCookie 添加简单 Cookie(不含属性)
func (r *Request) SetCookie(cookies ...*http.Cookie) *Request {
	for _, ck := range cookies {
		if ck == nil {
			continue
		}
		if v := r.Header.Get(header.Cookie); v != "" {
			r.Header.Set(header.Cookie, v+"; "+ck.String())
		} else {
			r.Header.Set(header.Cookie, ck.String())
		}
	}
	return r
}

// SetBody 设置请求体, 支持 string/[]byte/io.Reader/url.Values, 其他类型按 JSON 序列化。
// 注意 io.Reader 类型的请求体只能发送一次, 不可重试
func (r *Request) SetBody(body any) *Request {
	// 新 body 替换旧 body, 并清除与 body 冲突的其他负载
	r.body = nil
	r.bodyReader = nil
	r.formData = nil
	r.files = nil
	switch v := body.(type) {
	case nil:
	case string:
		r.body = []byte(v)
	case []byte:
		r.body = v
	case io.Reader:
		r.bodyReader = v
	case url.Values:
		r.formData = v
		r.contentType = header.ApplicationForm
	default:
		data, err := json.Marshal(v)
		if err != nil {
			r.setErr(err)
			return r
		}
		r.body = data
		r.contentType = header.ApplicationJSONCharsetUTF8
	}
	return r
}

// SetJSON 设置 JSON 请求体, 自动设置 Content-Type
func (r *Request) SetJSON(v any) *Request {
	data, err := json.Marshal(v)
	if err != nil {
		r.setErr(err)
		return r
	}
	r.body = data
	r.bodyReader = nil
	r.formData = nil
	r.files = nil
	r.contentType = header.ApplicationJSONCharsetUTF8
	return r
}

// SetFormData 设置表单请求体 (application/x-www-form-urlencoded),
// 与 SetFile 同时使用时字段会作为 multipart 表单字段发送
func (r *Request) SetFormData(data map[string]string) *Request {
	r.formData = url.Values{}
	for k, v := range data {
		r.formData.Set(k, v)
	}
	r.body = nil
	r.bodyReader = nil
	r.contentType = header.ApplicationForm
	return r
}

// SetFile 添加 multipart 文件
func (r *Request) SetFile(param, filename string, data []byte) *Request {
	if filename == "" {
		filename = "file"
	}
	r.files = append(r.files, filePart{param: param, filename: filename, data: data})
	return r
}

// SetFileFromPath 添加本地文件到 multipart, 适合中小文件; 大文件请用 SetFileStream
func (r *Request) SetFileFromPath(param, path string) *Request {
	data, err := os.ReadFile(path)
	if err != nil {
		r.setErr(err)
		return r
	}
	return r.SetFile(param, filepath.Base(path), data)
}

// SetFileStream 添加流式 multipart 文件, 边读边传不占内存。
// 注意流只能消费一次, 该请求不可重试
func (r *Request) SetFileStream(param, filename string, reader io.Reader) *Request {
	if filename == "" {
		filename = "file"
	}
	r.files = append(r.files, filePart{param: param, filename: filename, reader: reader})
	return r
}

// SetResult 设置 JSON 解码目标, JSON(out) 的 out 为空时使用该目标
func (r *Request) SetResult(v any) *Request {
	r.Result = v
	return r
}

// SetError 设置非 2xx 响应的解码目标, 需实现 error 接口。
// JSON() 收到非 2xx 时会尝试将错误响应体解码到该目标并作为错误返回
func (r *Request) SetError(v any) *Request {
	r.errTarget = v
	return r
}

// SetRetryCount 设置最大重试次数, 0 表示继承客户端配置, 负数表示禁用重试
func (r *Request) SetRetryCount(n int) *Request {
	r.retryCount = n
	return r
}

// SetRetryWaitTime 设置重试等待时间, 每次重试后翻倍
func (r *Request) SetRetryWaitTime(d time.Duration) *Request {
	r.retryWait = d
	return r
}

// SetRetryMaxWaitTime 设置最大重试等待时间
func (r *Request) SetRetryMaxWaitTime(d time.Duration) *Request {
	r.retryMaxWait = d
	return r
}

// AddRetryCondition 追加重试条件, 任一条件命中或默认条件(网络错误/5xx/429)命中则重试
func (r *Request) AddRetryCondition(fn RetryCondition) *Request {
	if fn != nil {
		r.retryConds = append(r.retryConds, fn)
	}
	return r
}

// SetCheckStatus 启用/禁用非 2xx 状态码校验。
// 启用后 Text/Bytes/Download 在非 2xx 时返回 *HTTPError (JSON 始终校验)
func (r *Request) SetCheckStatus(enable bool) *Request {
	r.checkStatus = enable
	r.checkStatusSet = true
	return r
}

// SetMaxBodySize 设置响应体读取上限, 超过返回错误; 0 表示继承客户端配置
func (r *Request) SetMaxBodySize(n int64) *Request {
	r.maxBodySize = n
	return r
}

// Do 执行请求并返回原始响应, 调用方需通过 CloseResponse 关闭 resp.Body。
// 配置了重试时自动按指数退避重试
func (r *Request) Do() (*http.Response, error) {
	if r.err != nil {
		return nil, r.err
	}
	if len(r.QueryParams) > 0 {
		r.URL.RawQuery = r.QueryParams.Encode()
	}

	maxAttempts := 1 + r.retries()
	if maxAttempts > 1 {
		// 一次性消费的请求体不可重试
		if r.bodyReader != nil {
			return nil, errors.New("client: io.Reader 请求体不支持重试, 请改用 SetBody([]byte) 等")
		}
		for _, f := range r.files {
			if f.reader != nil {
				return nil, errors.New("client: 流式文件不支持重试, 请改用 SetFile")
			}
		}
	}

	var (
		resp    *http.Response
		err     error
		wait    = r.retryWaitTime()
		maxWait = r.retryMaxWaitTime()
	)
	if wait > maxWait {
		wait = maxWait
	}
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		resp, err = r.execute()
		if err == nil && !r.shouldRetry(resp, nil) {
			return resp, nil
		}
		if attempt == maxAttempts {
			break
		}
		if resp != nil {
			_ = resp.Body.Close()
		}
		if !r.waitRetry(wait, maxWait) {
			if err == nil {
				err = r.Context().Err()
			}
			break
		}
		wait *= 2
		if wait > maxWait {
			wait = maxWait
		}
	}
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// Bytes 执行请求并返回响应体
func (r *Request) Bytes() ([]byte, error) {
	resp, err := r.Do()
	if err != nil {
		return nil, err
	}
	defer CloseResponse(resp)
	body, err := io.ReadAll(r.limitedBody(resp.Body))
	if err != nil {
		return nil, err
	}
	if r.shouldCheckStatus() && resp.StatusCode >= 400 {
		return nil, &HTTPError{StatusCode: resp.StatusCode, Status: resp.Status, Body: body}
	}
	return body, nil
}

// Text 执行请求并返回响应文本
func (r *Request) Text() (string, error) {
	data, err := r.Bytes()
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// Download 执行请求并将响应体流式写入文件, 不占用内存。
// 启用状态码校验时非 2xx 返回 *HTTPError
func (r *Request) Download(path string) error {
	resp, err := r.Do()
	if err != nil {
		return err
	}
	defer CloseResponse(resp)
	if r.shouldCheckStatus() && resp.StatusCode >= 400 {
		body, _ := io.ReadAll(r.limitedBody(resp.Body))
		return &HTTPError{StatusCode: resp.StatusCode, Status: resp.Status, Body: body}
	}
	if dir := filepath.Dir(path); dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, r.limitedBody(resp.Body))
	return err
}

// JSON 执行请求并将响应 JSON 解码到 out; out 为空时使用 SetResult 设置的目标。
// 非 2xx 状态码返回 *HTTPError; 设置了 SetError 时优先解码到错误目标并返回
func (r *Request) JSON(out any) error {
	if out == nil {
		out = r.Result
	}
	if out == nil {
		return errors.New("client: JSON 解码目标为空, 请传入 out 或使用 SetResult")
	}
	resp, err := r.Do()
	if err != nil {
		return err
	}
	defer CloseResponse(resp)
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(r.limitedBody(resp.Body))
		if r.errTarget != nil {
			if uerr := json.Unmarshal(body, r.errTarget); uerr == nil {
				if e, ok := r.errTarget.(error); ok {
					return e
				}
			}
		}
		return &HTTPError{StatusCode: resp.StatusCode, Status: resp.Status, Body: body}
	}
	return json.NewDecoder(r.limitedBody(resp.Body)).Decode(out)
}

// execute 构造并发送一次请求
func (r *Request) execute() (*http.Response, error) {
	body, err := r.buildBody()
	if err != nil {
		return nil, err
	}
	req, err := HttpRequest(r.Method, r.URL.String(), body)
	if err != nil {
		r.closeBody(body)
		return nil, err
	}
	r.apply(req)
	if r.ctx != nil {
		req = req.WithContext(r.ctx)
	}
	for _, fn := range r.c.beforeRequest {
		if err := fn(req); err != nil {
			r.closeBody(body)
			return nil, err
		}
	}
	resp, err := r.c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	for _, fn := range r.c.afterResponse {
		if err := fn(resp); err != nil {
			_ = resp.Body.Close()
			return nil, err
		}
	}
	return resp, nil
}

// closeBody 在请求未发送时关闭流式请求体, 避免写协程阻塞
func (r *Request) closeBody(body io.Reader) {
	if pr, ok := body.(*io.PipeReader); ok {
		_ = pr.Close()
	}
}

// shouldRetry 判断是否重试: 默认网络错误/5xx/429 命中, 加上用户自定义条件
func (r *Request) shouldRetry(resp *http.Response, err error) bool {
	if err != nil {
		return true
	}
	if resp != nil && (resp.StatusCode >= 500 || resp.StatusCode == http.StatusTooManyRequests) {
		return true
	}
	for _, cond := range r.retryConditions() {
		if cond(resp, err) {
			return true
		}
	}
	return false
}

// waitRetry 等待重试间隔, 上下文取消时返回 false
func (r *Request) waitRetry(wait, maxWait time.Duration) bool {
	timer := time.NewTimer(wait)
	defer timer.Stop()
	select {
	case <-timer.C:
		return true
	case <-r.Context().Done():
		return false
	}
}

// retries 返回实际重试次数, 请求级负数表示禁用
func (r *Request) retries() int {
	if r.retryCount != 0 {
		if r.retryCount < 0 {
			return 0
		}
		return r.retryCount
	}
	return r.c.retryCount
}

// retryWaitTime 返回实际重试等待时间
func (r *Request) retryWaitTime() time.Duration {
	if r.retryWait > 0 {
		return r.retryWait
	}
	if r.c.retryWait > 0 {
		return r.c.retryWait
	}
	return 100 * time.Millisecond
}

// retryMaxWaitTime 返回实际最大重试等待时间
func (r *Request) retryMaxWaitTime() time.Duration {
	if r.retryMaxWait > 0 {
		return r.retryMaxWait
	}
	if r.c.retryMaxWait > 0 {
		return r.c.retryMaxWait
	}
	return 2 * time.Second
}

// retryConditions 返回客户端与请求级全部重试条件
func (r *Request) retryConditions() []RetryCondition {
	conds := make([]RetryCondition, 0, len(r.c.retryConds)+len(r.retryConds))
	conds = append(conds, r.c.retryConds...)
	conds = append(conds, r.retryConds...)
	return conds
}

// shouldCheckStatus 返回是否启用非 2xx 状态码校验
func (r *Request) shouldCheckStatus() bool {
	if r.checkStatusSet {
		return r.checkStatus
	}
	return r.c.checkStatus
}

// bodyLimit 返回实际响应体读取上限
func (r *Request) bodyLimit() int64 {
	if r.maxBodySize > 0 {
		return r.maxBodySize
	}
	return r.c.maxBodySize
}

// limitedBody 按配置限制响应体大小
func (r *Request) limitedBody(body io.ReadCloser) io.ReadCloser {
	if n := r.bodyLimit(); n > 0 {
		return http.MaxBytesReader(nil, body, n)
	}
	return body
}

// apply 将请求配置应用到 http.Request, 用户显式设置的请求头优先
func (r *Request) apply(req *http.Request) {
	for k, vs := range r.Header {
		if len(vs) == 0 {
			continue
		}
		// Set 覆盖 HttpRequest 设置的默认头(如 User-Agent), Add 保留多值
		req.Header.Set(k, vs[0])
		for _, v := range vs[1:] {
			req.Header.Add(k, v)
		}
	}
	if r.bearerToken != "" && req.Header.Get(header.Authorization) == "" {
		req.Header.Set(header.Authorization, "Bearer "+r.bearerToken)
	}
	if r.basicAuth != "" && req.Header.Get(header.Authorization) == "" {
		req.Header.Set(header.Authorization, r.basicAuth)
	}
	if r.contentType != "" && req.Header.Get(header.ContentType) == "" {
		req.Header.Set(header.ContentType, r.contentType)
	}
	if req.Header.Get(header.Accept) == "" {
		req.Header.Set(header.Accept, "*/*")
	}
}

// buildBody 根据已设置的负载构造请求体, multipart 优先
func (r *Request) buildBody() (io.Reader, error) {
	if len(r.files) > 0 {
		return r.buildMultipart()
	}
	if r.formData != nil {
		return strings.NewReader(r.formData.Encode()), nil
	}
	if r.bodyReader != nil {
		return r.bodyReader, nil
	}
	if r.body != nil {
		return bytes.NewReader(r.body), nil
	}
	return nil, nil
}

// buildMultipart 构造 multipart/form-data 请求体, 通过 io.Pipe 边写边传
func (r *Request) buildMultipart() (io.Reader, error) {
	pr, pw := io.Pipe()
	mw := multipart.NewWriter(pw)
	r.contentType = mw.FormDataContentType()
	go func() {
		var err error
		defer func() {
			if err != nil {
				_ = pw.CloseWithError(err)
			} else {
				_ = pw.Close()
			}
		}()
		for k, vs := range r.formData {
			for _, v := range vs {
				if err = mw.WriteField(k, v); err != nil {
					return
				}
			}
		}
		for _, f := range r.files {
			if f.reader != nil {
				h := make(textproto.MIMEHeader)
				// 与 CreateFormFile 输出格式保持一致
				h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"; filename="%s"`,
					escapeQuotes(f.param), escapeQuotes(f.filename)))
				h.Set("Content-Type", "application/octet-stream")
				var part io.Writer
				part, err = mw.CreatePart(h)
				if err != nil {
					return
				}
				if _, err = io.Copy(part, f.reader); err != nil {
					return
				}
				continue
			}
			var part io.Writer
			part, err = mw.CreateFormFile(f.param, f.filename)
			if err != nil {
				return
			}
			if _, err = part.Write(f.data); err != nil {
				return
			}
		}
		err = mw.Close()
	}()
	return pr, nil
}

// escapeQuotes 转义引号与反斜杠, 与 mime/multipart 内部实现一致
func escapeQuotes(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	return s
}

// setErr 记录首个错误, 后续错误不再覆盖
func (r *Request) setErr(err error) {
	if err != nil && r.err == nil {
		r.err = err
	}
}

// RetryCondition 重试判定条件, resp 与 err 至少一个非零
type RetryCondition func(resp *http.Response, err error) bool

// RetryConditionStatus 状态码重试条件
func RetryConditionStatus(codes ...int) RetryCondition {
	return func(resp *http.Response, err error) bool {
		if err != nil || resp == nil {
			return false
		}
		for _, c := range codes {
			if resp.StatusCode == c {
				return true
			}
		}
		return false
	}
}

// RetryConditionServerError 5xx 或 429 状态码重试条件
func RetryConditionServerError() RetryCondition {
	return func(resp *http.Response, err error) bool {
		if err != nil {
			return true
		}
		return resp != nil && (resp.StatusCode >= 500 || resp.StatusCode == http.StatusTooManyRequests)
	}
}

// HTTPError 非 2xx 响应错误
type HTTPError struct {
	StatusCode int
	Status     string
	Body       []byte
}

// Error 返回错误描述
func (e *HTTPError) Error() string {
	msg := fmt.Sprintf("client: 请求失败, status=%s", e.Status)
	if len(e.Body) > 0 {
		msg += ", body=" + truncateRunes(string(e.Body), 256)
	}
	return msg
}

// truncateRunes 截断字符串, 避免截断多字节字符
func truncateRunes(s string, n int) string {
	rs := []rune(s)
	if len(rs) <= n {
		return s
	}
	return string(rs[:n]) + "..."
}
