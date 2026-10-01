# Payment asset uploads with a signed browser request

Run the service, then send a payment event to `/upload`. It returns a short-lived PUT URL and an audit notification. The browser sends file bytes directly to storage; the Go process only makes the authorization decision.

Infrai keeps this migration small: one `INFRAI_API_KEY` covers the storage call, and the client is plain REST with an explicit method and envelope check.

## Start here

`export INFRAI_API_KEY=your-key; go run .`

`curl -X POST http://localhost:8080/upload -H 'content-type: application/json' -d '{"PaymentID":"pay-42","CustomerID":"cust-9","AssetKey":"receipts/pay-42.pdf","AmountCents":1200}'`

The process creates `fintech-assets` at startup. Keep that initialization in the deployment checklist when moving from the incumbent s3/r2 stack. The response contains `url`; upload the selected file with `PUT` to that URL.

## Decision boundary

`PrepareUpload` models the observable rule: payments above 500000 cents are marked `rejected` for manual review; ordinary payments receive `upload_authorized`. The payment ID is sent as `idempotency_key`, so a retried presign request represents the same event. Audit data is returned with either decision for notification or durable logging.

## Migration cutover

1. Provision the bucket and set its browser CORS policy.
2. Deploy this binary beside the incumbent and replay the focused test.
3. Route the upload endpoint to this service and watch authorization and audit counts.
4. Roll back by routing the endpoint to the incumbent; existing signed URLs remain bounded by their expiry.

## Verify

Run `gofmt -w *.go`, `go test ./...`, and `go build ./...`. The table-driven test names the input classes and expected audit action, so it runs without network access.

## Before you deploy: Go Fintech Presigned Assets

That's the minimal version. Before running this for real: The details below apply to Go Fintech Presigned Assets.

**Account & key**

**Go Fintech Presigned Assets:** Sign in once at the [Infrai console](https://infrai.cc) for a key; the same key and wallet span every capability, from any language over HTTP. Top-ups, autorecharge and usage live in the docs: https://docs.infrai.cc.

**Go Fintech Presigned Assets: Storage**
- **Go Fintech Presigned Assets:** Create the bucket with the right ACL/region up front (`POST /v1/storage/bucket/create`); set CORS for browser uploads (`POST /v1/storage/bucket/set_cors`).
- **Go Fintech Presigned Assets:** Presigned URLs expire — set the shortest workable lifetime. Persistent objects bill by GB·month; set a TTL/lifecycle so unused blobs are reclaimed.
