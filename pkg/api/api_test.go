package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bougou/alertmanager-webhook-adapter/pkg/senders"
	webhookModels "github.com/bougou/alertmanager-webhook-adapter/pkg/webhook-adapter/models"
	restful "github.com/emicklei/go-restful/v3"
)

const testChannel = "test-guard"

type recordedSender struct {
	payloads []*webhookModels.Payload
}

func (s *recordedSender) Send(payload *webhookModels.Payload) error {
	s.payloads = append(s.payloads, payload)
	return nil
}

func (s *recordedSender) SendMsg(msg interface{}) error { return nil }

func (s *recordedSender) SendMsgT(msgType string, msg interface{}) error { return nil }

func newTestController(sender *recordedSender) *Controller {
	senders.RegisterChannelsSenderCreator(testChannel, func(request *restful.Request) (webhookModels.Sender, error) {
		return sender, nil
	})

	return NewController("test")
}

func postAlert(t *testing.T, c *Controller, body string) *httptest.ResponseRecorder {
	t.Helper()

	httpReq := httptest.NewRequest(http.MethodPost, "/webhook/send?channel_type="+testChannel, strings.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()
	c.send(restful.NewRequest(httpReq), restful.NewResponse(rec))

	return rec
}

func TestSendRejectsEmptyAlertList(t *testing.T) {
	sender := &recordedSender{}
	c := newTestController(sender)

	rec := postAlert(t, c, `{"version":"4","status":"firing","alerts":[]}`)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d; body: %s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
	if len(sender.payloads) != 0 {
		t.Errorf("sender called %d times, want 0", len(sender.payloads))
	}
}

func TestSendRejectsPayloadWithoutAlerts(t *testing.T) {
	sender := &recordedSender{}
	c := newTestController(sender)

	rec := postAlert(t, c, `{}`)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d; body: %s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
	if len(sender.payloads) != 0 {
		t.Errorf("sender called %d times, want 0", len(sender.payloads))
	}
}

func TestSendAcceptsPayloadWithAlerts(t *testing.T) {
	sender := &recordedSender{}
	c := newTestController(sender)

	body := `{
	  "version": "4",
	  "status": "firing",
	  "receiver": "test",
	  "groupLabels": {"alertname": "GuardTest"},
	  "commonLabels": {"alertname": "GuardTest", "severity": "critical"},
	  "commonAnnotations": {"summary": "guard test"},
	  "alerts": [{
	    "status": "firing",
	    "labels": {"alertname": "GuardTest", "severity": "critical"},
	    "annotations": {"summary": "guard test"},
	    "startsAt": "2026-10-26T07:00:00Z",
	    "endsAt": "0001-01-01T00:00:00Z"
	  }]
	}`

	rec := postAlert(t, c, body)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d; body: %s", rec.Code, http.StatusNoContent, rec.Body.String())
	}
	if len(sender.payloads) != 1 {
		t.Fatalf("sender called %d times, want 1", len(sender.payloads))
	}
	if strings.TrimSpace(sender.payloads[0].Markdown) == "" {
		t.Error("payload markdown is empty")
	}
}
