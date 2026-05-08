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

	if strings.Contains(string(source), "planAuth.OptionalAuth()") {
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
