# sdk-go

Official UniPost API client for Go.
Post to 7 social platforms with one API call.

## Latest release: v0.5.0

Media uploads now support custom audio overlay jobs and optional reserve-time file sizes.

- Use `client.Media.AudioOverlays.Create(...)` to combine one uploaded video with one uploaded audio file.
- Poll the job with `client.Media.AudioOverlays.Get(...)`, then publish the returned `OutputMediaID`.
- Leave `SizeBytes` as zero when reserving media if your app cannot know the raw file length up front.
- Post failure responses also include the typed v0.4.1 error contract fields.

Supported analytics surfaces include Instagram, Threads, Pinterest, and TikTok when connected account permissions allow them. See `Analytics Explorer` below for code.

## Installation

```bash
go get github.com/unipost-dev/sdk-go
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

```go
post, err := client.Posts.Create(ctx, params)
if err != nil {
    switch e := err.(type) {
    case *unipost.AuthError:
        fmt.Println("API key invalid")
    case *unipost.RateLimitError:
        fmt.Printf("Rate limited, retry after %ds\n", e.RetryAfter)
    case *unipost.ValidationError:
        fmt.Println("Validation failed:", e.Errors)
    case *unipost.UniPostError:
        fmt.Printf("API error: %d %s %s\n", e.Status, e.Code, e.Message)
    default:
        fmt.Println("Error:", err)
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
