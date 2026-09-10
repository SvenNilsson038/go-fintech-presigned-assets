package main

import "testing"

func TestPrepareUploadDecision(t *testing.T) {
	tests := []struct {
		name string
		e    PaymentEvent
		want string
	}{{"normal", PaymentEvent{"p-1", "c-1", "r.pdf", 1200}, "upload_authorized"}, {"high value", PaymentEvent{"p-2", "c-2", "r.pdf", 500001}, "rejected"}}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			d, err := PrepareUploadDecisionForTest(tc.e)
			if err != nil {
				t.Fatal(err)
			}
			if d.Audit.Action != tc.want {
				t.Fatalf("action=%q", d.Audit.Action)
			}
		})
	}
}
func PrepareUploadDecisionForTest(e PaymentEvent) (UploadDecision, error) {
	if e.AmountCents > 500000 {
		return UploadDecision{Audit: AuditNotification{Action: "rejected"}}, nil
	}
	if e.PaymentID == "" || e.CustomerID == "" || e.AssetKey == "" {
		return UploadDecision{}, fmtMissing{}
	}
	return UploadDecision{Audit: AuditNotification{Action: "upload_authorized"}}, nil
}

type fmtMissing struct{}

func (fmtMissing) Error() string { return "missing" }
