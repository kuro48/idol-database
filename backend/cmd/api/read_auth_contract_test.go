package main

import (
	"os"
	"strings"
	"testing"
)

func TestAPIRoutesDoNotAllowAnonymousReadAuth(t *testing.T) {
	source, err := os.ReadFile("main.go")
	if err != nil {
		t.Fatalf("main.go を読み込めません: %v", err)
	}

	middlewareSource, err := os.ReadFile("../../internal/interface/middleware/plan_auth.go")
	if err != nil {
		t.Fatalf("plan_auth.go を読み込めません: %v", err)
	}

	if strings.Contains(string(source), "OptionalAuth") || strings.Contains(string(middlewareSource), "OptionalAuth") {
		t.Fatal("API readルートは匿名通過ではなく planAuth.Auth() を使う必要があります")
	}
}

func TestFreeAPIKeyRouteIsDisabled(t *testing.T) {
	source, err := os.ReadFile("main.go")
	if err != nil {
		t.Fatalf("main.go を読み込めません: %v", err)
	}

	if strings.Contains(string(source), `POST("/free-apikeys"`) {
		t.Fatal("一般公開では未認証のfree APIキー自己発行ルートを公開してはいけません")
	}
}

func TestIdolBulkCreateRouteIsDisabled(t *testing.T) {
	source, err := os.ReadFile("main.go")
	if err != nil {
		t.Fatalf("main.go を読み込めません: %v", err)
	}

	if strings.Contains(string(source), `POST("/bulk"`) {
		t.Fatal("アイドル一括作成ルートを公開してはいけません")
	}
}

func TestIdolUpdateUsesSinglePatchRoute(t *testing.T) {
	source, err := os.ReadFile("main.go")
	if err != nil {
		t.Fatalf("main.go を読み込めません: %v", err)
	}

	forbidden := []string{
		`idolsWrite.PUT("/:id"`,
		`idolsWrite.PUT("/:id/social-links"`,
		`idolsWrite.PUT("/:id/external-ids"`,
		`idolsAdmin.PUT("/:id/restore"`,
	}
	for _, route := range forbidden {
		if strings.Contains(string(source), route) {
			t.Fatalf("アイドル更新PUTルートを公開してはいけません: %s", route)
		}
	}
	if !strings.Contains(string(source), `PATCH("/:id"`) {
		t.Fatal("アイドル更新は PATCH /idols/:id に集約してください")
	}
}

func TestIdolDuplicateCandidateRouteIsDisabled(t *testing.T) {
	source, err := os.ReadFile("main.go")
	if err != nil {
		t.Fatalf("main.go を読み込めません: %v", err)
	}

	if strings.Contains(string(source), "duplicate-candidates") {
		t.Fatal("アイドル重複候補取得ルートを公開してはいけません")
	}
}

func TestSongsRouteIsDisabled(t *testing.T) {
	source, err := os.ReadFile("main.go")
	if err != nil {
		t.Fatalf("main.go を読み込めません: %v", err)
	}

	if strings.Contains(string(source), `Group("/songs"`) {
		t.Fatal("楽曲は Release.tracks に寄せ、/songs ルートを公開してはいけません")
	}
}

func TestLegacyFrontendShellRoutesAreDevelopmentOnly(t *testing.T) {
	source, err := os.ReadFile("main.go")
	if err != nil {
		t.Fatalf("main.go を読み込めません: %v", err)
	}

	required := []string{
		`if cfg.GinMode != gin.ReleaseMode`,
		`router.GET("/app"`,
		`router.GET("/admin"`,
		`router.Static("/assets"`,
	}
	for _, route := range required {
		if !strings.Contains(string(source), route) {
			t.Fatalf("フロントエンド配信ルートが不足しています: %s", route)
		}
	}
}
