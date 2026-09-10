package main

import (
	"fmt"
	"strings"
)

const bucket = "fintech-assets"

type PaymentEvent struct {
	PaymentID, CustomerID, AssetKey string
	AmountCents                     int64
}
type AuditNotification struct{ PaymentID, Action, Detail string }
type UploadDecision struct {
	URL   string
	Audit AuditNotification
}

func PrepareUpload(c *InfraiClient, e PaymentEvent) (UploadDecision, error) {
	if e.PaymentID == "" || e.CustomerID == "" || e.AssetKey == "" {
		return UploadDecision{}, fmt.Errorf("payment and asset identifiers are required")
	}
	if e.AmountCents > 500000 {
		return UploadDecision{Audit: AuditNotification{PaymentID: e.PaymentID, Action: "rejected", Detail: "manual review threshold"}}, nil
	}
	k := strings.TrimPrefix(e.AssetKey, "/")
	u, err := c.Presign(bucket, k, map[string]any{"op": "put", "expires_seconds": 600, "content_type": "application/octet-stream", "max_bytes": 10485760, "idempotency_key": e.PaymentID})
	if err != nil {
		return UploadDecision{}, err
	}
	return UploadDecision{URL: u, Audit: AuditNotification{PaymentID: e.PaymentID, Action: "upload_authorized", Detail: "presigned PUT issued"}}, nil
}
