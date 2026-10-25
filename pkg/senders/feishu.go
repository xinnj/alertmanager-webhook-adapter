package senders

import (
	"fmt"

	"github.com/bougou/alertmanager-webhook-adapter/pkg/webhook-adapter/channels/feishu"
	"github.com/bougou/alertmanager-webhook-adapter/pkg/webhook-adapter/models"
	restful "github.com/emicklei/go-restful/v3"
)

const (
	ChannelTypeFeishu = "feishu"
)

func init() {
	RegisterChannelsSenderCreator(ChannelTypeFeishu, createFeishuSender)
}

func createFeishuSender(request *restful.Request) (models.Sender, error) {
	token := credentialFrom(request, "token", "FEISHU_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("not token found for feishu channel")
	}

	secret := credentialFrom(request, "secret", "FEISHU_SECRET")

	msgType := request.QueryParameter("msg_type")
	if msgType == "" {
		msgType = "markdown"
	}

	var sender models.Sender = feishu.NewSenderWithSecret(token, msgType, secret)
	return sender, nil
}
