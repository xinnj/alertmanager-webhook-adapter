package senders

import (
	"os"

	"github.com/bougou/alertmanager-webhook-adapter/pkg/webhook-adapter/models"
	restful "github.com/emicklei/go-restful/v3"
)

type ChannelSenderCreator func(request *restful.Request) (models.Sender, error)

var ChannelsSenderCreatorMap = map[string]ChannelSenderCreator{}

func RegisterChannelsSenderCreator(channel string, creator ChannelSenderCreator) {
	ChannelsSenderCreatorMap[channel] = creator
}

// credentialFrom returns the query parameter value, falling back to the
// environment variable when the query parameter is absent or empty. This lets
// Alertmanager receiver URLs stay free of credentials.
func credentialFrom(request *restful.Request, queryKey string, envKey string) string {
	if value := request.QueryParameter(queryKey); value != "" {
		return value
	}

	return os.Getenv(envKey)
}
