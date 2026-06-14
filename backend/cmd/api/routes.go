package main

import (
	"github.com/gin-gonic/gin"
	"github.com/kuro48/idol-api/internal/interface/handlers"
	"github.com/kuro48/idol-api/internal/interface/middleware"
)

// routeDeps はルート登録に必要なハンドラーとミドルウェアをまとめる
type routeDeps struct {
	// ミドルウェア
	writeAuth             gin.HandlerFunc
	adminAuth             gin.HandlerFunc
	userAuth              gin.HandlerFunc
	publicMutationLimiter *middleware.RateLimiter
	planAuth              *middleware.PlanAuthMiddleware

	// ハンドラー
	idolHandler        *handlers.IdolHandler
	removalHandler     *handlers.RemovalHandler
	groupHandler       *handlers.GroupHandler
	agencyHandler      *handlers.AgencyHandler
	eventHandler       *handlers.EventHandler
	tagHandler         *handlers.TagHandler
	webhookHandler     *handlers.WebhookHandler
	exportHandler      *handlers.ExportHandler
	submissionHandler  *handlers.SubmissionHandler
	releaseHandler     *handlers.ReleaseHandler
	editHistoryHandler *handlers.EditHistoryHandler
	membershipHandler  *handlers.MembershipHandler
	venueHandler       *handlers.VenueHandler
	apikeyHandler      *handlers.APIKeyHandler
	meHandler          *handlers.MeHandler
	analyticsHandler   *handlers.AnalyticsHandler
	jobHandler         *handlers.JobHandler
	termHandler        *handlers.TermHandler
	billingHandler     *handlers.BillingHandler // nil の場合は課金ルートを登録しない
}

// registerAPIRoutes は /api/v1 以下のすべての API ルートを登録する。
// ミドルウェアの適用順序・グループ構成は main.go での設計から変えない。
func registerAPIRoutes(v1 *gin.RouterGroup, d routeDeps) {
	v1.GET("/me", d.userAuth, d.meHandler.GetMe)
	v1.GET("/me/submissions", d.userAuth, d.submissionHandler.ListMySubmissions)
	v1.GET("/me/removal-requests", d.userAuth, d.removalHandler.ListMyRemovalRequests)

	// アイドル: 読み取りは公開、書き込みは write スコープ必須
	idols := v1.Group("/idols")
	{
		idols.GET("", d.idolHandler.ListIdols)                                 // 一覧取得
		idols.GET("/:id", d.idolHandler.GetIdol)                               // 詳細取得
		idols.GET("/:id/external-ids", d.idolHandler.GetExternalIDs)           // 外部IDマッピング取得
		idols.GET("/:id/memberships", d.membershipHandler.ListIdolMemberships) // メンバーシップ一覧
	}
	idolsWrite := v1.Group("/idols", d.writeAuth)
	{
		idolsWrite.POST("", d.idolHandler.CreateIdol)       // 新規作成
		idolsWrite.PATCH("/:id", d.idolHandler.PatchIdol)   // 更新
		idolsWrite.DELETE("/:id", d.idolHandler.DeleteIdol) // 削除
	}

	// 削除申請: 申請はログイン必須、参照は投稿者トークン、管理は admin スコープ必須
	removalRequests := v1.Group("/removal-requests")
	{
		removalRequests.POST("", d.userAuth, d.publicMutationLimiter.Limit(), d.removalHandler.CreateRemovalRequest) // 削除申請作成
		removalRequests.GET("/:id", d.userAuth, d.removalHandler.GetRemovalRequest)                                  // 削除申請詳細取得（ログイン必須）
	}
	adminRemoval := v1.Group("/removal-requests", d.adminAuth)
	{
		adminRemoval.GET("", d.removalHandler.ListAllRemovalRequests)             // 全削除申請取得
		adminRemoval.GET("/pending", d.removalHandler.ListPendingRemovalRequests) // 保留中取得
		adminRemoval.GET("/overdue", d.removalHandler.ListOverdueRemovalRequests) // SLA超過取得
		adminRemoval.PUT("/:id", d.removalHandler.UpdateStatus)                   // ステータス更新
	}

	// APIキー管理（admin スコープ必須）
	adminAPIKeys := v1.Group("/admin/apikeys", d.adminAuth)
	{
		adminAPIKeys.POST("", d.apikeyHandler.CreateAPIKey)       // APIキー作成
		adminAPIKeys.GET("", d.apikeyHandler.ListAPIKeys)         // APIキー一覧（?email=）
		adminAPIKeys.DELETE("/:id", d.apikeyHandler.RevokeAPIKey) // APIキー無効化
	}

	// API利用分析（admin スコープ必須）
	adminAnalytics := v1.Group("/admin/analytics", d.adminAuth)
	{
		adminAnalytics.GET("/usage", d.analyticsHandler.GetUsageSummary) // API利用サマリー取得
	}

	// 非同期ジョブ管理（admin スコープ必須）
	adminJobs := v1.Group("/admin/jobs", d.adminAuth)
	{
		adminJobs.POST("/bulk-import", d.jobHandler.EnqueueBulkImport) // バルクインポートジョブ作成
		adminJobs.GET("/:id", d.jobHandler.GetJobStatus)               // ジョブステータス取得
		adminJobs.POST("/:id/retry", d.jobHandler.RetryJob)            // ジョブリトライ
	}

	// Webhook管理（admin スコープ必須）
	adminWebhooks := v1.Group("/admin/webhooks", d.adminAuth)
	{
		adminWebhooks.POST("", d.webhookHandler.CreateSubscription)       // 購読作成
		adminWebhooks.GET("", d.webhookHandler.ListSubscriptions)         // 購読一覧
		adminWebhooks.DELETE("/:id", d.webhookHandler.DeleteSubscription) // 購読削除
	}

	// Webhook受信エンドポイント（公開: 外部からの受信）
	v1.POST("/webhooks/receive/:subscription_id", d.publicMutationLimiter.Limit(), d.webhookHandler.ReceiveWebhook)

	// 編集履歴（admin スコープ必須）
	adminEditHistory := v1.Group("/admin/edit-history", d.adminAuth)
	{
		adminEditHistory.GET("", d.editHistoryHandler.ListEditHistory)    // 編集履歴一覧
		adminEditHistory.GET("/:id", d.editHistoryHandler.GetEditHistory) // 編集履歴詳細
	}

	// エクスポート（admin スコープ必須）
	adminExport := v1.Group("/admin/export", d.adminAuth)
	{
		adminExport.GET("/idols", d.exportHandler.ExportIdols)   // アイドルエクスポート
		adminExport.GET("/logs", d.exportHandler.ListExportLogs) // 実行履歴
	}

	if d.billingHandler != nil {
		billing := v1.Group("/billing")
		{
			billing.POST("/checkout-sessions", d.publicMutationLimiter.Limit(), d.billingHandler.CreateCheckoutSession)
			billing.POST("/webhooks/stripe", d.billingHandler.HandleStripeWebhook)
		}

		billingAuth := v1.Group("/billing", d.planAuth.Auth())
		{
			billingAuth.POST("/portal-sessions", d.billingHandler.CreatePortalSession)
		}
	}

	// グループ: 読み取りは公開、書き込みは write スコープ必須
	groups := v1.Group("/groups")
	{
		groups.GET("", d.groupHandler.ListGroup)
		groups.GET("/:id", d.groupHandler.GetGroup)
		groups.GET("/:id/memberships", d.membershipHandler.ListGroupMemberships) // メンバーシップ一覧
	}
	groupsWrite := v1.Group("/groups", d.writeAuth)
	{
		groupsWrite.POST("", d.groupHandler.CreateGroup)
		groupsWrite.PUT("/:id", d.groupHandler.UpdateGroup)
		groupsWrite.DELETE("/:id", d.groupHandler.DeleteGroup)
	}

	// メンバーシップ: 読み取りは公開、書き込みは write スコープ必須
	memberships := v1.Group("/memberships")
	{
		memberships.GET("", d.membershipHandler.ListMemberships)
		memberships.GET("/:id", d.membershipHandler.GetMembership)
	}
	membershipsWrite := v1.Group("/memberships", d.writeAuth)
	{
		membershipsWrite.POST("", d.membershipHandler.CreateMembership)
		membershipsWrite.PUT("/:id", d.membershipHandler.UpdateMembership)
		membershipsWrite.DELETE("/:id", d.membershipHandler.DeleteMembership)
	}

	// 会場: 読み取りは公開、書き込みは write スコープ必須
	venues := v1.Group("/venues")
	{
		venues.GET("", d.venueHandler.ListVenues)
		venues.GET("/:id", d.venueHandler.GetVenue)
	}
	venuesWrite := v1.Group("/venues", d.writeAuth)
	{
		venuesWrite.POST("", d.venueHandler.CreateVenue)
		venuesWrite.PUT("/:id", d.venueHandler.UpdateVenue)
		venuesWrite.DELETE("/:id", d.venueHandler.DeleteVenue)
	}

	// 事務所: 読み取りは公開、書き込みは write スコープ必須
	agencies := v1.Group("/agencies")
	{
		agencies.GET("", d.agencyHandler.ListAgencies)
		agencies.GET("/:id", d.agencyHandler.GetAgency)
	}
	agenciesWrite := v1.Group("/agencies", d.writeAuth)
	{
		agenciesWrite.POST("", d.agencyHandler.CreateAgency)
		agenciesWrite.PUT("/:id", d.agencyHandler.UpdateAgency)
		agenciesWrite.DELETE("/:id", d.agencyHandler.DeleteAgency)
	}

	terms := v1.Group("/terms")
	{
		terms.GET("/service", d.termHandler.ShowTermsOfService)
		terms.GET("/privacy", d.termHandler.ShowPrivacyPolicy)
	}

	// イベント: 読み取りは公開、書き込みは write スコープ必須
	events := v1.Group("/events")
	{
		events.GET("", d.eventHandler.ListEvents)                 // イベント一覧取得（検索機能付き）
		events.GET("/upcoming", d.eventHandler.GetUpcomingEvents) // 今後のイベント取得
		events.GET("/:id", d.eventHandler.GetEvent)               // イベント詳細取得
	}
	eventsWrite := v1.Group("/events", d.writeAuth)
	{
		eventsWrite.POST("", d.eventHandler.CreateEvent)                                    // イベント作成
		eventsWrite.PUT("/:id", d.eventHandler.UpdateEvent)                                 // イベント更新
		eventsWrite.DELETE("/:id", d.eventHandler.DeleteEvent)                              // イベント削除
		eventsWrite.POST("/:id/performers", d.eventHandler.AddPerformer)                    // パフォーマー追加
		eventsWrite.DELETE("/:id/performers/:performer_id", d.eventHandler.RemovePerformer) // パフォーマー削除
	}

	// リリース: 読み取りは公開、書き込みは write スコープ必須
	releases := v1.Group("/releases")
	{
		releases.GET("", d.releaseHandler.ListReleases)
		releases.GET("/:id", d.releaseHandler.GetRelease)
	}
	releasesWrite := v1.Group("/releases", d.writeAuth)
	{
		releasesWrite.POST("", d.releaseHandler.CreateRelease)
		releasesWrite.PUT("/:id", d.releaseHandler.UpdateRelease)
		releasesWrite.DELETE("/:id", d.releaseHandler.DeleteRelease)
		releasesWrite.PUT("/:id/streaming-links", d.releaseHandler.UpdateStreamingLinks)
		releasesWrite.PUT("/:id/external-ids", d.releaseHandler.UpdateExternalIDs)
	}
	releasesAdmin := v1.Group("/releases", d.adminAuth)
	{
		releasesAdmin.PUT("/:id/restore", d.releaseHandler.RestoreRelease)
	}

	// 投稿審査: 作成はログイン必須、取得は投稿者トークン、審査は admin スコープ必須
	submissions := v1.Group("/submissions")
	{
		submissions.POST("", d.userAuth, d.publicMutationLimiter.Limit(), d.submissionHandler.CreateSubmission)           // 投稿作成
		submissions.GET("/:id", d.userAuth, d.submissionHandler.GetSubmission)                                            // 投稿詳細取得（ログイン必須）
		submissions.PUT("/:id/revise", d.userAuth, d.publicMutationLimiter.Limit(), d.submissionHandler.ReviseSubmission) // 差し戻し後の再投稿（ログイン必須）
	}
	adminSubmissions := v1.Group("/submissions", d.adminAuth)
	{
		adminSubmissions.GET("", d.submissionHandler.ListAllSubmissions)             // 全投稿一覧
		adminSubmissions.GET("/pending", d.submissionHandler.ListPendingSubmissions) // 審査待ち一覧
		adminSubmissions.PUT("/:id/status", d.submissionHandler.UpdateStatus)        // ステータス更新
	}

	// タグ: 読み取りは公開、書き込みは write スコープ必須
	tags := v1.Group("/tags")
	{
		tags.GET("", d.tagHandler.ListTags)   // タグ一覧取得
		tags.GET("/:id", d.tagHandler.GetTag) // タグ詳細取得
	}
	tagsWrite := v1.Group("/tags", d.writeAuth)
	{
		tagsWrite.POST("", d.tagHandler.CreateTag)       // タグ作成
		tagsWrite.PUT("/:id", d.tagHandler.UpdateTag)    // タグ更新
		tagsWrite.DELETE("/:id", d.tagHandler.DeleteTag) // タグ削除
	}
}
