# sdk-go

Official UniPost API client for Go.
Post to 7 social platforms with one API call.

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
