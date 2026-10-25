package dingtalk

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
)

// GenSign returns the DingTalk custom robot signature for the given
// timestamp (in milliseconds) and secret.
//
// Per the DingTalk docs, the string to sign is "timestamp\nsecret" and it is
// used as the HMAC-SHA256 key with the secret as the message, which is the
// mirror image of Feishu's algorithm.
// ref: https://open.dingtalk.com/document/orgapp/custom-robot-access
func GenSign(secret string, timestampMs int64) (string, error) {
	stringToSign := fmt.Sprintf("%d\n%s", timestampMs, secret)
	h := hmac.New(sha256.New, []byte(secret))
	if _, err := h.Write([]byte(stringToSign)); err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(h.Sum(nil)), nil
}
