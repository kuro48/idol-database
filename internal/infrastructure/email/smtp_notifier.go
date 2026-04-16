package email

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"net/smtp"
	"strings"

	"github.com/kuro48/idol-api/internal/usecase/submission"
)

// SMTPNotifier は SMTP を使ったメール通知の実装
type SMTPNotifier struct {
	host     string
	port     int
	username string
	password string
	from     string
	fromName string
}

// SMTPConfig は SMTPNotifier の設定
type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
	FromName string
}

// NewSMTPNotifier は SMTPNotifier を作成する
func NewSMTPNotifier(cfg SMTPConfig) *SMTPNotifier {
	return &SMTPNotifier{
		host:     cfg.Host,
		port:     cfg.Port,
		username: cfg.Username,
		password: cfg.Password,
		from:     cfg.From,
		fromName: cfg.FromName,
	}
}

// NotifyStatusChanged は投稿審査ステータス変更を投稿者にメール通知する
func (n *SMTPNotifier) NotifyStatusChanged(ctx context.Context, notification submission.StatusNotification) error {
	subject, body := buildMessage(notification)

	if err := n.send(notification.To, subject, body); err != nil {
		return fmt.Errorf("メール送信エラー: %w", err)
	}

	slog.Info("メール通知送信完了",
		"to", notification.To,
		"submission_id", notification.SubmissionID,
		"status", notification.Status,
	)
	return nil
}

// send は SMTP でメールを送信する（STARTTLS対応）
func (n *SMTPNotifier) send(to, subject, body string) error {
	addr := fmt.Sprintf("%s:%d", n.host, n.port)

	header := strings.Builder{}
	header.WriteString(fmt.Sprintf("From: %s <%s>\r\n", n.fromName, n.from))
	header.WriteString(fmt.Sprintf("To: %s\r\n", to))
	header.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
	header.WriteString("MIME-Version: 1.0\r\n")
	header.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	header.WriteString("\r\n")
	header.WriteString(body)

	msg := []byte(header.String())

	// 認証情報が設定されている場合のみ AUTH を使用
	var auth smtp.Auth
	if n.username != "" {
		auth = smtp.PlainAuth("", n.username, n.password, n.host)
	}

	// port 465 は直接 TLS、それ以外は STARTTLS を試みる
	if n.port == 465 {
		return n.sendTLS(addr, auth, to, msg)
	}

	return smtp.SendMail(addr, auth, n.from, []string{to}, msg)
}

// sendTLS は TLS 接続でメールを送信する（ポート465用）
func (n *SMTPNotifier) sendTLS(addr string, auth smtp.Auth, to string, msg []byte) error {
	tlsCfg := &tls.Config{
		ServerName: n.host,
		MinVersion: tls.VersionTLS12,
	}

	conn, err := tls.Dial("tcp", addr, tlsCfg)
	if err != nil {
		return fmt.Errorf("TLS接続エラー: %w", err)
	}

	client, err := smtp.NewClient(conn, n.host)
	if err != nil {
		return fmt.Errorf("SMTPクライアント作成エラー: %w", err)
	}
	defer client.Close()

	if auth != nil {
		if err := client.Auth(auth); err != nil {
			return fmt.Errorf("SMTP認証エラー: %w", err)
		}
	}

	if err := client.Mail(n.from); err != nil {
		return fmt.Errorf("MAIL FROMエラー: %w", err)
	}
	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("RCPT TOエラー: %w", err)
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("DATA開始エラー: %w", err)
	}
	if _, err := w.Write(msg); err != nil {
		return fmt.Errorf("メール本文書き込みエラー: %w", err)
	}
	return w.Close()
}

// buildMessage はステータスに応じた件名と本文を返す
func buildMessage(n submission.StatusNotification) (subject, body string) {
	targetLabel := targetTypeLabel(n.TargetType)

	switch n.Status {
	case "approved":
		subject = fmt.Sprintf("【Idol API】投稿が承認されました（%s）", targetLabel)
		body = fmt.Sprintf(`投稿が承認されました。

投稿ID: %s
対象種別: %s
ステータス: 承認済み

ご投稿いただきありがとうございました。
審査を通過しましたので、情報はデータベースに反映されます。

---
Idol API
`, n.SubmissionID, targetLabel)

	case "rejected":
		subject = fmt.Sprintf("【Idol API】投稿が却下されました（%s）", targetLabel)
		body = fmt.Sprintf(`投稿が却下されました。

投稿ID: %s
対象種別: %s
ステータス: 却下

公式サイトとの照合ができなかったか、投稿ポリシーに合致しない内容が含まれていました。
別の情報での再投稿をお待ちしています。

---
Idol API
`, n.SubmissionID, targetLabel)

	case "needs_revision":
		subject = fmt.Sprintf("【Idol API】投稿の修正をお願いします（%s）", targetLabel)
		body = fmt.Sprintf(`投稿の内容に修正が必要です。

投稿ID: %s
対象種別: %s
ステータス: 修正依頼

【修正依頼内容】
%s

上記の内容を修正のうえ、再投稿してください。

---
Idol API
`, n.SubmissionID, targetLabel, n.RevisionNote)

	default:
		subject = "【Idol API】投稿のステータスが更新されました"
		body = fmt.Sprintf("投稿ID %s のステータスが %s に更新されました。\n", n.SubmissionID, n.Status)
	}

	return subject, body
}

// targetTypeLabel は投稿タイプを日本語ラベルに変換する
func targetTypeLabel(t string) string {
	switch t {
	case "idol":
		return "アイドル"
	case "group":
		return "グループ"
	case "agency":
		return "事務所"
	case "event":
		return "イベント"
	default:
		return t
	}
}
