# sdk-go

Official UniPost API client for Go.
Post to 7 social platforms with one API call.

## Latest release: v0.7.0

v0.6.0 adds the production Inbox client for direct messages, comments, and
replies. Every Inbox operation is explicitly bound to either one managed user
or the owner/admin workspace aggregate. Replies expose completed and accepted-
for-reconciliation outcomes, and X backfill uses typed estimate, confirmation,
in-progress, and completed results.

Media audio overlays, analytics, and the other v0.5.0 APIs remain available.

## Installation

```bash
go get github.com/unipost-dev/sdk-go@v0.7.0
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "github.com/unipost-dev/sdk-go/unipost"
)

func main() {
    // Reads UNIPOST_API_KEY from environment automatically
    client := unipost.NewClient()

    ctx := context.Background()

    post, err := client.Posts.Create(ctx, &unipost.CreatePostParams{
        Caption:    "Hello from UniPost!",
        AccountIDs: []string{"sa_twitter_xxx", "sa_linkedin_xxx"},
    })
    if err != nil {
        panic(err)
    }

    fmt.Println("Post created:", post.ID)
}
```

## Usage

### List Accounts

```go
accounts, err := client.Accounts.List(ctx, nil)

// Filter by platform
accounts, err := client.Accounts.List(ctx, &unipost.ListAccountsParams{
    Platform: "twitter",
})
```

### Get Connect URL (Your Own Accounts)

```go
connect, err := client.Connect.GetConnectURL(ctx, &unipost.GetConnectURLParams{
    ProfileID:   "pr_brand_us",
    Platform:    "linkedin",
    RedirectURL: "https://app.acme.com/integrations/done", // optional
})
if err != nil {
    panic(err)
}

fmt.Println(connect.AuthURL)
```

### Connect (Managed Users)

```go
session, err := client.Connect.CreateSession(ctx, &unipost.CreateConnectSessionParams{
    Platform:              "twitter",
    ExternalUserID:        "your_user_123",
    ReturnURL:             "https://yourapp.com/callback",
    AllowQuickstartCreds:  true, // optional
})
if err != nil {
    panic(err)
}

fmt.Println(session.URL)
```

### Production Inbox Integration

Inbox API keys are backend credentials. Keep the UniPost workspace key in your
server-side secret store and create the SDK client only in trusted backend code.
Never return the key to a managed user or embed it in browser JavaScript, a
mobile app, a WebSocket URL, or application logs.

Derive the stable external user ID from your app's authenticated session. Do not
accept an arbitrary `external_user_id` from a request body or query string:

```go
externalUserID := authenticatedSession.UserID // established by auth middleware
managedInbox, err := client.Inbox.ManagedUser(externalUserID)
if err != nil {
    return err
}
```

`ManagedUser(externalUserID)` never falls back to workspace access: a blank ID
returns an error before a scoped resource is created or a request is sent. Use
`Workspace()` only for an explicit owner/admin aggregate view. Workspace access
is authorized by the UniPost API
key and is allowed only while that key's creator remains an owner or admin of
the UniPost workspace. That authorization is separate from your end-app roles;
an end-app "admin" label does not grant UniPost workspace access.

#### List a managed user's Inbox

`List` supports exactly `Source`, `IsRead`, `IsOwn`, and `Limit`. Boolean filters
are pointers so an explicit `false` is transmitted instead of being omitted.
The collection is limit-only: there is no cursor, offset, or pagination API.
When `limit` is omitted, invalid, zero, or negative, the server uses 50; values
above 500 are clamped to 500.

```go
managedInbox, err := client.Inbox.ManagedUser(authenticatedSession.UserID)
if err != nil {
    return err
}
isRead := false
isOwn := false

items, err := managedInbox.List(ctx, &unipost.InboxListParams{
    Source: unipost.InboxSourceXDM,
    IsRead: &isRead,
    IsOwn:  &isOwn,
    Limit:  100,
})
if err != nil {
    return err
}
for _, item := range items.Data {
    fmt.Println(item.ID, item.Source, item.IsRead)
}
```

The same scoped resource exposes the remaining read and workflow operations:

```go
unread, err := managedInbox.UnreadCount(ctx)
item, err := managedInbox.Get(ctx, "inbox_item_123")
err = managedInbox.MarkRead(ctx, item.ID)
marked, err := managedInbox.MarkAllRead(ctx)

assignee := "support-agent-42"
updated, err := managedInbox.UpdateThreadState(ctx, item.ID, &unipost.InboxThreadStateRequest{
    ThreadStatus: unipost.InboxThreadStatusAssigned,
    AssignedTo:   &assignee,
})
media, err := managedInbox.MediaContext(ctx, item.ID)

fmt.Println(unread.Count, marked.Marked, updated.ThreadStatus, media.Permalink)
```

#### Reply exactly once

Create one stable idempotency key for a logical reply and reuse that same key
if your own job retries. Never resend the same logical reply with a new key.
The SDK itself performs one POST and does not automatically retry or follow a
redirect.

HTTP 200 produces `InboxReplyStateCompleted` with an Inbox item. A valid HTTP 202
produces `InboxReplyStateReconciling`; poll the returned operation ID with
`XOutboundStatus` instead of sending the reply again.

```go
idempotencyKey := "reply-order-8721-comment-4" // persist with the logical job
reply, err := managedInbox.Reply(
    ctx,
    "inbox_item_123",
    &unipost.InboxReplyRequest{Text: "Thanks—we are looking into this."},
    unipost.WithIdempotencyKey(idempotencyKey),
)
if err != nil {
    return err
}

switch reply.State {
case unipost.InboxReplyStateCompleted:
    fmt.Println("reply item", reply.Item.ID)
case unipost.InboxReplyStateReconciling:
    status, err := managedInbox.XOutboundStatus(ctx, reply.OperationID)
    if err != nil {
        return err
    }
    fmt.Println("reconciliation", status.Status)
}
```

#### Backend WebSocket connection details

`WebSocketConnectionDetails()` is local-only: it performs no network request
and has no WebSocket runtime dependency. It converts the configured HTTP base
URL to `/v1/inbox/ws`, includes only the bound scope in the URL, and returns the
key only in the `Authorization` header map.

```go
details, err := managedInbox.WebSocketConnectionDetails()
if err != nil {
    return err
}
// Pass details.URL and details.Headers to a backend WebSocket implementation.
```

Native browser WebSocket APIs cannot attach an `Authorization` header. Do not
work around that limitation by putting the workspace key in the URL or sending
it to browser/mobile code; terminate or proxy the connection in your backend.

#### Sync and metered X backfill

Ordinary sync has no backfill request and returns a typed result:

```go
syncResult, err := managedInbox.Sync(ctx)
if err != nil {
    return err
}
fmt.Println(syncResult.AccountsChecked, syncResult.NewItems)
```

`SyncXBackfill` is a separate, metered operation. Review the X credit estimate,
account selection, lookback, and maximum item count before confirmation. A
managed-user scope limits the account blast radius to that managed user; a
workspace scope can cover every eligible account when `AccountID` is omitted.
Never schedule an unreviewed workspace-wide backfill.

The first call can return a confirmation token. Treat it as a short-lived
secret: do not log it, persist it in analytics, or expose it to clients. Submit
the token only after the operator approves the displayed estimate and scope.

```go
lookbackDays := 7
maxItems := 250
estimate, err := managedInbox.SyncXBackfill(ctx, &unipost.XInboxBackfillRequest{
    LookbackDays:   &lookbackDays,
    MaxItems:       &maxItems,
    IncludeReplies: true,
    IncludeDMs:     true,
})
if err != nil {
    return err
}

confirmation, ok := estimate.(*unipost.XInboxBackfillConfirmationRequired)
if ok {
    if confirmation.EstimatedXCredits != nil {
        fmt.Println("estimated X credits", *confirmation.EstimatedXCredits)
    }
    // Continue only after an operator approves the estimate and scope.
    confirmationToken := confirmation.ConfirmationToken
    estimate, err = managedInbox.SyncXBackfill(ctx, &unipost.XInboxBackfillRequest{
        LookbackDays:      &lookbackDays,
        MaxItems:          &maxItems,
        IncludeReplies:    true,
        IncludeDMs:        true,
        ConfirmationToken: &confirmationToken,
    })
    if err != nil {
        return err
    }
}

switch result := estimate.(type) {
case *unipost.XInboxBackfillConfirmationRequired:
    fmt.Println("approval required; accounts", result.AccountsChecked)
case *unipost.XInboxBackfillInProgress:
    fmt.Println("backfill operation", result.ConfirmationOperationID)
case *unipost.XInboxBackfillCompleted:
    fmt.Println("backfill complete", result.Accepted, result.Suppressed)
}
```

For a deliberately reviewed owner/admin aggregate, replace `managedInbox` with
`client.Inbox.Workspace()` and keep the same methods. Do not use workspace scope
as an error fallback for a missing managed-user ID.

### Create Posts

```go
// Immediate publish
post, err := client.Posts.Create(ctx, &unipost.CreatePostParams{
    Caption:    "Hello world!",
    AccountIDs: []string{"sa_twitter_xxx"},
})

// Scheduled
post, err := client.Posts.Create(ctx, &unipost.CreatePostParams{
    Caption:     "Later!",
    AccountIDs:  []string{"sa_twitter_xxx"},
    ScheduledAt: "2026-04-28T09:00:00Z",
})

// Per-platform captions
post, err := client.Posts.Create(ctx, &unipost.CreatePostParams{
    PlatformPosts: []unipost.CreatePlatformPost{
        {AccountID: "sa_twitter_xxx", Caption: "Short tweet"},
        {AccountID: "sa_linkedin_xxx", Caption: "Longer LinkedIn version"},
    },
})

// Draft
post, err := client.Posts.Create(ctx, &unipost.CreatePostParams{
    Caption:    "Work in progress",
    AccountIDs: []string{"sa_twitter_xxx"},
    Status:     "draft",
})
```

### Analytics Explorer

```go
posts, err := client.Analytics.Posts(ctx, &unipost.AnalyticsPostsParams{
    Platform: "tiktok",
    Limit:    25,
    Sort:     "engagement_rate",
})

platforms, err := client.Analytics.Platforms(ctx, nil)
tiktok, err := client.Analytics.Platform(ctx, "tiktok", nil)
csv, err := client.Analytics.ExportPostsCSV(ctx, &unipost.AnalyticsPostsParams{
    Platform: "pinterest",
})

_, err = client.Analytics.Refresh(ctx, &unipost.AnalyticsRefreshParams{
    Platform: "threads",
    Limit:    100,
})
```

### Developer Logs

```go
logs, err := client.Logs.List(ctx, &unipost.LogListParams{
    Status: "error",
    Limit:  50,
})
if err != nil {
    panic(err)
}

afterID := int64(0)
if len(logs.Data) > 0 {
    afterID = logs.Data[0].ID - 1
    entry, err := client.Logs.Get(ctx, logs.Data[0].ID)
    if err != nil {
        panic(err)
    }
    fmt.Println(entry.Action, entry.RequestPayload)
}

stream, err := client.Logs.Stream(ctx, &unipost.LogStreamParams{
    Status:  "error",
    AfterID: afterID,
})
if err != nil {
    panic(err)
}
defer stream.Close()

if stream.Next() {
    fmt.Println(stream.Event().ID, stream.Event().Action)
}
```

### Media Upload

```go
reserved, err := client.Media.Upload(ctx, &unipost.MediaUploadRequest{
    Filename:    "voiceover.mp3",
    ContentType: "audio/mpeg",
    // SizeBytes is optional.
})
if err != nil {
    panic(err)
}
fmt.Println(reserved.MediaID)
```

### Custom Audio Overlay

```go
videoVolume := int32(70)
job, err := client.Media.AudioOverlays.Create(ctx, &unipost.AudioOverlayCreateRequest{
    VideoMediaID: "media_video_123",
    AudioMediaID: "media_audio_456",
    Mode:         "mix",
    VideoVolume:  &videoVolume,
    Fit:          "trim_to_video",
}, unipost.WithIdempotencyKey("overlay-demo-001"))
if err != nil {
    panic(err)
}

for job.Status == "queued" || job.Status == "processing" {
    time.Sleep(1500 * time.Millisecond)
    job, err = client.Media.AudioOverlays.Get(ctx, job.ID)
    if err != nil {
        panic(err)
    }
}

if job.Status != "succeeded" || job.OutputMediaID == nil {
    panic("audio overlay failed")
}

post, err := client.Posts.Create(ctx, &unipost.CreatePostParams{
    Caption:  "Video with custom audio",
    AccountIDs: []string{"sa_tiktok_xxx"},
    MediaIDs: []string{*job.OutputMediaID},
})
```

### Error Handling

All non-2xx API responses use `*unipost.APIError`. `Code` normally prefers the
canonical `NormalizedCode` when the API supplies one. Inbox reply errors preserve
the raw server `Code` so callers can distinguish X reconciliation outcomes;
`NormalizedCode` remains available separately. Validation details and rate-limit
delay are populated only when applicable.

```go
package main

import (
    "errors"
    "fmt"

    "github.com/unipost-dev/sdk-go/unipost"
)

func logUniPostError(err error) {
    var apiErr *unipost.APIError
    if !errors.As(err, &apiErr) {
        fmt.Println("Request error:", err)
        return
    }

    fmt.Printf(
        "API error: status=%d code=%s normalized_code=%s message=%s\n",
        apiErr.Status,
        apiErr.Code,
        apiErr.NormalizedCode,
        apiErr.Message,
    )
    if len(apiErr.Errors) > 0 {
        fmt.Println("Validation errors:", apiErr.Errors)
    }
    if apiErr.RetryAfter > 0 {
        fmt.Printf("Retry after %d seconds\n", apiErr.RetryAfter)
    }
}
```

### Webhook Verification

```go
isValid := unipost.VerifyWebhookSignature(
    requestBody,
    r.Header.Get("X-UniPost-Signature"),
    os.Getenv("UNIPOST_WEBHOOK_SECRET"),
)
```

## License

MIT
