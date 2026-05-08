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
