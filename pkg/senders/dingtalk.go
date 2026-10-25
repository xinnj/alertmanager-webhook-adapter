package senders

import (
	"fmt"

	"github.com/bougou/alertmanager-webhook-adapter/pkg/webhook-adapter/channels/dingtalk"
	"github.com/bougou/alertmanager-webhook-adapter/pkg/webhook-adapter/models"
	restful "github.com/emicklei/go-restful/v3"
)

const (
	ChannelTypeDingtalk = "dingtalk"
)

func init() {
	RegisterChannelsSenderCreator(ChannelTypeDingtalk, createDingtalkSender)
}

func createDingtalkSender(request *restful.Request) (models.Sender, error) {
	token := credentialFrom(request, "token", "DINGTALK_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("not token found for dingtalk channel")
	}

	secret := credentialFrom(request, "secret", "DINGTALK_SECRET")

	msgType := request.QueryParameter("msg_type")
	if msgType == "" {
		msgType = "markdown"
	}

	var sender models.Sender = dingtalk.NewSenderWithSecret(token, msgType, secret)
	return sender, nil
}
