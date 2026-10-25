package dingtalk

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/bougou/alertmanager-webhook-adapter/pkg/webhook-adapter/utils"
)

// DingtalkBot can send messages to dingtalk group
// ref: https://developers.dingtalk.com/document/app/message-types-and-data-format
type DingtalkGroupBot struct {
	addr         string
	access_token string
	secret       string
	client       *http.Client
}

func NewDingtalkGroupBot(access_token string) *DingtalkGroupBot {
	return NewDingtalkGroupBotWithSecret(access_token, "")
}

func NewDingtalkGroupBotWithSecret(access_token string, secret string) *DingtalkGroupBot {
	return &DingtalkGroupBot{
		addr:         "https://oapi.dingtalk.com",
		access_token: access_token,
		secret:       secret,
		client:       &http.Client{},
	}
}

func (bot *DingtalkGroupBot) Addr() string {
	if bot.secret == "" {
		return fmt.Sprintf("%s/robot/send?access_token=%s", bot.addr, bot.access_token)
	}

	timestamp := time.Now().UnixMilli()
	// GenSign cannot fail for well-formed inputs; the error return only mirrors
	// the Feishu helper. Never fall back to an unsigned URL on error.
	sign, _ := GenSign(bot.secret, timestamp)

	q := url.Values{}
	q.Set("access_token", bot.access_token)
	q.Set("timestamp", strconv.FormatInt(timestamp, 10))
	q.Set("sign", sign)
	return fmt.Sprintf("%s/robot/send?%s", bot.addr, q.Encode())
}

func (bot *DingtalkGroupBot) AddrForUpload() string {
	return fmt.Sprintf("%s/robot/upload_media?key=%s&type=file", bot.addr, bot.access_token)
}

func (bot *DingtalkGroupBot) Send(msg *Msg) error {
	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", bot.Addr(), bytes.NewBuffer(msgBytes))
	if err != nil {
		return fmt.Errorf("failed to construct request, got %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	res, err := bot.client.Do(req)
	if err != nil {
		return fmt.Errorf("send msg error, %v", utils.SanitizeRequestError(err))
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return fmt.Errorf("send msg response error, status: %d", res.StatusCode)
	}

	// DingTalk returns HTTP 200 with a non-zero errcode for rejected requests
	// (invalid token, bad signature, stale timestamp, ...).
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("read response body failed: %v", err)
	}
	if len(bytes.TrimSpace(body)) == 0 {
		return nil
	}

	var result struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("unexpected response body from dingtalk (status %d)", res.StatusCode)
	}
	if result.ErrCode != 0 {
		return fmt.Errorf("dingtalk send failed: errcode=%d errmsg=%s", result.ErrCode, result.ErrMsg)
	}

	return nil
}

func (bot *DingtalkGroupBot) UploadFile(filename string, fileReader io.Reader) (meidaID string, err error) {

	// Todo
	// 要求文件大小在 5B~20M 之间

	body := new(bytes.Buffer)
	w := multipart.NewWriter(body)
	fileWriter, err := w.CreateFormFile("media", filename)
	if err != nil {
		return "", fmt.Errorf("create file writer error, got %v", err)
	}
	io.Copy(fileWriter, fileReader)

	req, err := http.NewRequest("POST", bot.AddrForUpload(), body)
	if err != nil {
		return "", fmt.Errorf("failed to construct request, got %v", err)
	}
	req.Header.Set("Content-Type", w.FormDataContentType())

	res, err := bot.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("send error, %v", utils.SanitizeRequestError(err))
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return "", fmt.Errorf("upload file failed, status: %d", res.StatusCode)
	}

	type ResponseBody struct {
		ErrCode   int       `json:"errcode"`
		ErrMsg    string    `json:"errmsg"`
		Type      string    `json:"type"`       // 媒体文件类型，分别有图片（image）、语音（voice）、视频（video），普通文件 (file)
		MediaID   string    `json:"media_id"`   // 媒体文件上传后获取的唯一标识，3 天内有效
		CreatedAt time.Time `json:"created_at"` // 媒体文件上传时间戳
	}

	r := &ResponseBody{}
	if err := json.NewDecoder(res.Body).Decode(r); err != nil {
		return "", fmt.Errorf("can decode res body, err: %v", err)
	}

	return r.MediaID, nil
}
