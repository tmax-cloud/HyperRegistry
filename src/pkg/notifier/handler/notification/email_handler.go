package notification

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/goharbor/harbor/src/common/job/models"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/pkg/notification"
	"github.com/goharbor/harbor/src/pkg/notifier/model"
)

// EmailHandler preprocess http event data and start the hook processing
type EmailHandler struct {
}

// Name ...
func (h *EmailHandler) Name() string {
	return "SMTP"
}

// Handle handles http event
func (h *EmailHandler) Handle(ctx context.Context, value interface{}) error {
	if value == nil {
		return errors.New("EmailHandler cannot handle nil value")
	}

	event, ok := value.(*model.HookEvent)
	if !ok || event == nil {
		return errors.New("invalid notification http event")
	}
	return h.process(ctx, event)
}

// IsStateful ...
func (h *EmailHandler) IsStateful() bool {
	return false
}

func (h *EmailHandler) process(ctx context.Context, event *model.HookEvent) error {
	j := &models.JobData{
		Metadata: &models.JobMetadata{
			JobKind: job.KindGeneric,
		},
	}
	j.Name = job.EmailJob

	payload, err := json.Marshal(event.Payload)
	if err != nil {
		return fmt.Errorf("marshal from payload %v failed: %v", event.Payload, err)
	}

	j.Parameters = map[string]interface{}{
		"payload": string(payload),
		"address": event.Target.Address,
		// Users can define a auth header in http statement in notification(webhook) policy.
		// So it will be sent in header in http request.
		"auth_header":      event.Target.AuthHeader,
		"skip_cert_verify": event.Target.SkipCertVerify,
	}
	return notification.HookManager.StartHook(ctx, event, j)
}
