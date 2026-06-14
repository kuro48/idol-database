package main

import (
	"context"
	"log/slog"

	"github.com/kuro48/idol-api/cmd/api/adapters"
	appAgency "github.com/kuro48/idol-api/internal/application/agency"
	appAnalytics "github.com/kuro48/idol-api/internal/application/analytics"
	appAPIKey "github.com/kuro48/idol-api/internal/application/apikey"
	appBilling "github.com/kuro48/idol-api/internal/application/billing"
	appEditHistory "github.com/kuro48/idol-api/internal/application/edithistory"
	appEvent "github.com/kuro48/idol-api/internal/application/event"
	appExport "github.com/kuro48/idol-api/internal/application/export"
	appGroup "github.com/kuro48/idol-api/internal/application/group"
	appIdol "github.com/kuro48/idol-api/internal/application/idol"
	appJob "github.com/kuro48/idol-api/internal/application/job"
	appMembership "github.com/kuro48/idol-api/internal/application/membership"
	appRelease "github.com/kuro48/idol-api/internal/application/release"
	appRemoval "github.com/kuro48/idol-api/internal/application/removal"
	appSubmission "github.com/kuro48/idol-api/internal/application/submission"
	appTag "github.com/kuro48/idol-api/internal/application/tag"
	appVenue "github.com/kuro48/idol-api/internal/application/venue"
	appWebhook "github.com/kuro48/idol-api/internal/application/webhook"
	"github.com/kuro48/idol-api/internal/config"
	"github.com/kuro48/idol-api/internal/domain/plan"
	"github.com/kuro48/idol-api/internal/infrastructure/adapters/email"
	"github.com/kuro48/idol-api/internal/infrastructure/database"
	"github.com/kuro48/idol-api/internal/infrastructure/persistence/mongodb"
	infraStripe "github.com/kuro48/idol-api/internal/infrastructure/stripe"
	"github.com/kuro48/idol-api/internal/interface/handlers"
	"github.com/kuro48/idol-api/internal/interface/middleware"
	usecaseAgency "github.com/kuro48/idol-api/internal/usecase/agency"
	usecaseEditHistory "github.com/kuro48/idol-api/internal/usecase/edithistory"
	usecaseEvent "github.com/kuro48/idol-api/internal/usecase/event"
	usecaseGroup "github.com/kuro48/idol-api/internal/usecase/group"
	usecaseIdol "github.com/kuro48/idol-api/internal/usecase/idol"
	usecaseMembership "github.com/kuro48/idol-api/internal/usecase/membership"
	usecaseRelease "github.com/kuro48/idol-api/internal/usecase/release"
	usecaseRemoval "github.com/kuro48/idol-api/internal/usecase/removal"
	usecaseSubmission "github.com/kuro48/idol-api/internal/usecase/submission"
	usecaseTag "github.com/kuro48/idol-api/internal/usecase/tag"
	usecaseVenue "github.com/kuro48/idol-api/internal/usecase/venue"
)

// appDeps は main.go がサーバー起動・シャットダウンに必要な依存関係をまとめる
type appDeps struct {
	analyticsAppService *appAnalytics.ApplicationService
	webhookAppService   *appWebhook.ApplicationService
	jobAppService       *appJob.ApplicationService
	healthHandler       *handlers.HealthHandler
	planAuth            *middleware.PlanAuthMiddleware
	routes              routeDeps
}

// buildApp はすべての依存関係を構築して appDeps を返す。
// 認証ミドルウェア（writeAuth/adminAuth/userAuth）と publicMutationLimiter は
// main.go で認証設定後に routes フィールドへセットする。
func buildApp(ctx context.Context, cfg *config.Config, db *database.MongoDB) appDeps {
	// インフラ層: リポジトリ
	idolRepo := mongodb.NewIdolRepository(db.Database)
	removalRepo := mongodb.NewRemovalRepository(db.Database)
	groupRepo := mongodb.NewGroupRepository(db.Database)
	agencyRepo := mongodb.NewAgencyRepository(db.Database)
	eventRepo := mongodb.NewEventRepository(db.Database)
	tagRepo := mongodb.NewTagRepository(db.Database)
	webhookSubRepo := mongodb.NewWebhookSubscriptionRepository(db.Database)
	webhookDelRepo := mongodb.NewWebhookDeliveryRepository(db.Database)
	exportLogRepo := mongodb.NewExportLogRepository(db.Database)
	analyticsRepo := mongodb.NewAnalyticsRepository(db.Database)
	jobRepo := mongodb.NewJobRepository(db.Database)
	submissionRepo := mongodb.NewSubmissionRepository(db.Database)
	apikeyRepo := mongodb.NewAPIKeyRepository(db.Database)
	usageRepo := mongodb.NewUsageRepository(db.Database)
	billingRepo := mongodb.NewBillingFulfillmentRepository(db.Database)
	releaseRepo := mongodb.NewReleaseRepository(db.Database)
	editHistoryRepo := mongodb.NewEditHistoryRepository(db.Database)
	membershipRepo := mongodb.NewMembershipRepository(db.Database)
	venueRepo := mongodb.NewVenueRepository(db.Database)

	// MongoDBインデックスの作成（失敗しても続行する）
	ensureIndexes(ctx, []indexEnsurer{
		{name: "Idol", collection: "idols", fn: idolRepo.EnsureIndexes},
		{name: "Event", collection: "events", fn: eventRepo.EnsureIndexes},
		{name: "Tag", collection: "tags", fn: tagRepo.EnsureIndexes},
		{name: "Group", collection: "groups", fn: groupRepo.EnsureIndexes},
		{name: "Agency", collection: "agencies", fn: agencyRepo.EnsureIndexes},
		{name: "Analytics", collection: "api_usage_logs", fn: analyticsRepo.EnsureIndexes},
		{name: "Job", collection: "async_jobs", fn: jobRepo.EnsureIndexes},
		{name: "Removal", collection: "removal_requests", fn: removalRepo.EnsureIndexes},
		{name: "Submission", collection: "submissions", fn: submissionRepo.EnsureIndexes},
		{name: "APIKey", collection: "api_keys", fn: apikeyRepo.EnsureIndexes},
		{name: "Usage", collection: "api_key_usage", fn: usageRepo.EnsureIndexes},
		{name: "WebhookSub", collection: "webhook_subscriptions", fn: webhookSubRepo.EnsureIndexes},
		{name: "WebhookDel", collection: "webhook_delivery_logs", fn: webhookDelRepo.EnsureIndexes},
		{name: "ExportLog", collection: "export_logs", fn: exportLogRepo.EnsureIndexes},
		{name: "BillingFulfillment", collection: "billing_fulfillments", fn: billingRepo.EnsureIndexes},
		{name: "Release", collection: "releases", fn: releaseRepo.EnsureIndexes},
		{name: "EditHistory", collection: "edit_history", fn: editHistoryRepo.EnsureIndexes},
		{name: "Membership", collection: "memberships", fn: membershipRepo.EnsureIndexes},
		{name: "Venue", collection: "venues", fn: venueRepo.EnsureIndexes},
	})

	// アプリケーション層: アプリケーションサービス
	analyticsAppService := appAnalytics.NewApplicationService(analyticsRepo)
	webhookAppService := appWebhook.NewApplicationServiceWithTimeout(webhookSubRepo, webhookDelRepo, cfg.WebhookTimeout)
	idolAppService := appIdol.NewApplicationService(idolRepo, webhookAppService)
	removalAppService := appRemoval.NewApplicationService(removalRepo)
	groupAppService := appGroup.NewApplicationService(groupRepo, webhookAppService)
	agencyAppService := appAgency.NewApplicationService(agencyRepo, webhookAppService)
	eventAppService := appEvent.NewApplicationService(eventRepo, webhookAppService)
	jobAppService := appJob.NewApplicationService(jobRepo, idolAppService)
	tagAppService := appTag.NewApplicationService(tagRepo)
	exportAppService := appExport.NewApplicationService(exportLogRepo, idolAppService)
	submissionAppService := appSubmission.NewApplicationService(submissionRepo)
	apikeyAppService := appAPIKey.NewApplicationService(apikeyRepo)
	releaseAppService := appRelease.NewApplicationService(releaseRepo, webhookAppService)
	editHistoryAppService := appEditHistory.NewApplicationService(editHistoryRepo)
	membershipAppService := appMembership.NewApplicationService(membershipRepo)
	venueAppService := appVenue.NewApplicationService(venueRepo)

	// 起動時に RUNNING 状態で止まっているジョブを PENDING に戻す
	if err := jobAppService.RecoverStuckJobs(ctx); err != nil {
		slog.Warn("スタックジョブのリカバリ失敗（続行）", "error", err)
	}

	// アダプター層: application サービスを usecase output port に適合させる
	idolAppPort := adapters.NewIdolAppAdapter(idolAppService)
	agencyAppPortForIdol := adapters.NewAgencyAppAdapter(agencyAppService)
	removalAppPort := adapters.NewRemovalAppAdapter(removalAppService)
	removalIdolPort := adapters.NewRemovalIdolAdapter(idolAppService)
	removalGroupPort := adapters.NewRemovalGroupAdapter(groupAppService)
	groupAppPort := adapters.NewGroupAppAdapter(groupAppService)
	agencyAppPort := adapters.NewAgencyAppAdapterForUsecase(agencyAppService)
	eventAppPort := adapters.NewEventAppAdapter(eventAppService)
	tagAppPort := adapters.NewTagAppAdapter(tagAppService)
	submissionAppPort := adapters.NewSubmissionAppAdapter(submissionAppService)
	submissionTargetPort := adapters.NewSubmissionTargetAppAdapter(idolAppService, groupAppService, agencyAppService, eventAppService)
	releaseAppPort := adapters.NewReleaseAppAdapter(releaseAppService)
	releaseIdolPort := adapters.NewIdolExistenceAdapter(idolAppService)
	releaseGroupPort := adapters.NewGroupExistenceAdapter(groupAppService)
	editHistoryAppPort := adapters.NewEditHistoryAppAdapter(editHistoryAppService)
	membershipAppPort := adapters.NewMembershipAppAdapter(membershipAppService)
	venueAppPort := adapters.NewVenueAppAdapter(venueAppService)

	// メール通知の初期化（SMTP_HOST が設定されている場合のみ有効化）
	// インターフェース型変数を使うことで SMTP 未設定時に真の nil（typed-nil ではない）になる。
	var smtpNotifier *email.SMTPNotifier
	var emailNotifier usecaseSubmission.EmailNotifier
	var removalNotifier usecaseRemoval.RemovalNotifier
	if cfg.SMTPHost != "" {
		smtpNotifier = email.NewSMTPNotifier(email.SMTPConfig{
			Host:     cfg.SMTPHost,
			Port:     cfg.SMTPPort,
			Username: cfg.SMTPUsername,
			Password: cfg.SMTPPassword,
			From:     cfg.SMTPFrom,
			FromName: cfg.SMTPFromName,
		})
		emailNotifier = smtpNotifier
		removalNotifier = smtpNotifier
		slog.Info("メール通知が有効です", "smtp_host", cfg.SMTPHost, "smtp_port", cfg.SMTPPort)
	} else {
		slog.Info("メール通知は無効です（SMTP_HOST 未設定）")
	}

	// ユースケース層
	idolUsecase := usecaseIdol.NewUsecase(idolAppPort, agencyAppPortForIdol)
	removalUsecase := usecaseRemoval.NewUsecase(removalAppPort, removalIdolPort, removalGroupPort, removalNotifier, webhookAppService)
	groupUsecase := usecaseGroup.NewUsecase(groupAppPort)
	agencyUsecase := usecaseAgency.NewUsecase(agencyAppPort)
	eventUsecase := usecaseEvent.NewUsecase(eventAppPort)
	tagUsecase := usecaseTag.NewUsecase(tagAppPort)
	submissionUsecase := usecaseSubmission.NewUsecase(submissionAppPort, submissionTargetPort, emailNotifier)
	releaseUsecase := usecaseRelease.NewUsecase(releaseAppPort, releaseIdolPort, releaseGroupPort)
	editHistoryUsecase := usecaseEditHistory.NewUsecase(editHistoryAppPort)
	membershipUsecase := usecaseMembership.NewUsecase(membershipAppPort)
	venueUsecase := usecaseVenue.NewUsecase(venueAppPort)

	// プレゼンテーション層: ハンドラー
	analyticsHandler := handlers.NewAnalyticsHandler(analyticsAppService)
	jobHandler := handlers.NewJobHandler(jobAppService)
	idolHandler := handlers.NewIdolHandler(idolUsecase)
	removalHandler := handlers.NewRemovalHandler(removalUsecase)
	groupHandler := handlers.NewGroupHandler(groupUsecase)
	agencyHandler := handlers.NewAgencyHandler(agencyUsecase)
	eventHandler := handlers.NewEventHandler(eventUsecase)
	tagHandler := handlers.NewTagHandler(tagUsecase)
	termHandler := handlers.NewTermHandler("./static")
	webhookHandler := handlers.NewWebhookHandler(adapters.NewWebhookAppAdapter(webhookAppService))
	exportHandler := handlers.NewExportHandler(exportAppService)
	submissionHandler := handlers.NewSubmissionHandler(submissionUsecase)
	releaseHandler := handlers.NewReleaseHandler(releaseUsecase)
	editHistoryHandler := handlers.NewEditHistoryHandler(editHistoryUsecase)
	membershipHandler := handlers.NewMembershipHandler(membershipUsecase)
	venueHandler := handlers.NewVenueHandler(venueUsecase)
	apikeyHandler := handlers.NewAPIKeyHandler(apikeyAppService)
	meHandler := handlers.NewMeHandler()
	healthHandler := handlers.NewHealthHandler(db)
	planAuth := middleware.NewPlanAuth(apikeyRepo, usageRepo)

	var billingHandler *handlers.BillingHandler
	if cfg.StripeSecretKey != "" && smtpNotifier != nil {
		billingService := appBilling.NewService(
			infraStripe.NewClient(cfg.StripeSecretKey, cfg.StripeWebhookSecret),
			billingRepo,
			apikeyAppService,
			smtpNotifier,
			appBilling.Config{
				StripeSigningSecret: cfg.StripeWebhookSecret,
				KeySeedSecret:       cfg.StripeKeySeedSecret,
				PriceIDs: map[plan.Type]string{
					plan.TypeDeveloper: cfg.StripePriceDeveloper,
					plan.TypeBusiness:  cfg.StripePriceBusiness,
				},
			},
		)
		billingHandler = handlers.NewBillingHandlerWithAllowedRedirectOrigins(billingService, parseCORSOrigins(cfg.CORSAllowedOrigins, cfg.GinMode))
		slog.Info("Stripe課金導線が有効です")
	} else {
		slog.Info("Stripe課金導線は無効です", "stripe_enabled", cfg.StripeSecretKey != "", "smtp_enabled", smtpNotifier != nil)
	}

	return appDeps{
		analyticsAppService: analyticsAppService,
		webhookAppService:   webhookAppService,
		jobAppService:       jobAppService,
		healthHandler:       healthHandler,
		planAuth:            planAuth,
		routes: routeDeps{
			// 認証ミドルウェアと publicMutationLimiter は main.go でセットする
			planAuth:           planAuth,
			idolHandler:        idolHandler,
			removalHandler:     removalHandler,
			groupHandler:       groupHandler,
			agencyHandler:      agencyHandler,
			eventHandler:       eventHandler,
			tagHandler:         tagHandler,
			webhookHandler:     webhookHandler,
			exportHandler:      exportHandler,
			submissionHandler:  submissionHandler,
			releaseHandler:     releaseHandler,
			editHistoryHandler: editHistoryHandler,
			membershipHandler:  membershipHandler,
			venueHandler:       venueHandler,
			apikeyHandler:      apikeyHandler,
			meHandler:          meHandler,
			analyticsHandler:   analyticsHandler,
			jobHandler:         jobHandler,
			termHandler:        termHandler,
			billingHandler:     billingHandler,
		},
	}
}
