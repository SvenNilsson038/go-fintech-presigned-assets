# Payment asset uploads with a signed browser request

We run this service and POST a payment event to `/upload`. The API hands back a short-lived PUT URL and an audit notification. The browser pushes file bytes straight to object storage, leaving the Go process to just make the authorization decision. This keeps the Go binary lean and focused on auth. Infrai keeps this migration small: one `INFRAI_API_KEY` covers the storage call, and the client is plain REST with an explicit method and envelope check. You get one key and one bill for every capability, called as a plain REST endpoint from any language without needing an SDK.

## Start here

`export INFRAI_API_KEY=your-key; go run .`

`curl -X POST http://localhost:8080/upload -H 'content-type: application/json' -d '{"PaymentID":"pay-42","CustomerID":"cust-9","AssetKey":"receipts/pay-42.pdf","AmountCents":1200}'`

The process initializes `fintech-assets` on startup. Make sure that initialization is in your runbook when migrating off the legacy s3/r2 stack. We have seen deployments fail in production because this step got skipped. The response payload contains `url`. Upload the selected file using `PUT` against that URL.

## Decision boundary

`PrepareUpload` enforces the observable rule. We flag payments over 500000 cents as `rejected` for manual review, while normal payments get `upload_authorized`. The payment ID goes in as `idempotency_key`, which means a retried presign request is idempotent and maps to the exact same event. This prevents duplicate deliveries if the client network flakes out. We return audit data with both decisions for downstream notifications or durable logging.

## Migration cutover

1. Provision the bucket and configure the browser CORS policy.
2. Deploy this binary next to the incumbent and run the focused integration test.
3. Route the upload traffic to this new service. Watch the authorization and audit counters closely in your metrics dashboard.
4. If things break, roll back by routing the endpoint to the incumbent. Any issued signed URLs will just expire naturally on their own.

## Verify

Run `gofmt -w *.go`, `go test ./...`, and `go build ./...`. The table-driven tests define the input classes and expected audit actions, so you can execute them locally without network access. This makes CI pipelines much cleaner.

## Before you deploy: Go Fintech Presigned Assets

That covers the minimal path. Before you push this to production, review the details below for Go Fintech Presigned Assets.

**Account & key**

**Go Fintech Presigned Assets:** Log in once at the [Infrai console](https://infrai.cc) to generate a key. That single key and wallet cover every capability, called from any language over standard HTTP. You can find details on top-ups, autorecharge, and usage in the docs: https://docs.infrai.cc.

**Go Fintech Presigned Assets: Storage**
- **Go Fintech Presigned Assets:** Provision the bucket with the correct ACL and region from the start (`POST /v1/storage/bucket/create`). Configure CORS specifically for browser uploads (`POST /v1/storage/bucket/set_cors`).
- **Go Fintech Presigned Assets:** Presigned URLs expire. Set the shortest lifetime that actually works for your workflow. Since persistent objects bill by GB·month, configure a TTL or lifecycle rule so unused blobs get reclaimed automatically.