package idol

import (
	"errors"
	"strings"
)

// SocialLinks はアイドルのSNS/外部リンク情報を表す値オブジェクト
type SocialLinks struct {
	twitter   *string
	instagram *string
	tiktok    *string
	youtube   *string
	facebook  *string
	official  *string
	fanClub   *string
}

// NewSocialLinks は新しいSocialLinksを生成する
func NewSocialLinks() *SocialLinks {
	return &SocialLinks{}
}

// Getters
func (s *SocialLinks) Twitter() *string {
	return s.twitter
}

func (s *SocialLinks) Instagram() *string {
	return s.instagram
}

func (s *SocialLinks) TikTok() *string {
	return s.tiktok
}

func (s *SocialLinks) YouTube() *string {
	return s.youtube
}

func (s *SocialLinks) Facebook() *string {
	return s.facebook
}

func (s *SocialLinks) Official() *string {
	return s.official
}

func (s *SocialLinks) FanClub() *string {
	return s.fanClub
}

// SetTwitter はTwitter URLを設定する
func (s *SocialLinks) SetTwitter(url string) error {
	if url == "" {
		s.twitter = nil
		return nil
	}
	if err := validateURL(url); err != nil {
		return err
	}
	if !strings.Contains(url, "twitter.com") && !strings.Contains(url, "x.com") {
		return errors.New("無効なTwitter URLです")
	}
	s.twitter = &url
	return nil
}

// SetInstagram はInstagram URLを設定する
func (s *SocialLinks) SetInstagram(url string) error {
	if url == "" {
		s.instagram = nil
		return nil
	}
	if err := validateURL(url); err != nil {
		return err
	}
	if !strings.Contains(url, "instagram.com") {
		return errors.New("無効なInstagram URLです")
	}
	s.instagram = &url
	return nil
}

// SetTikTok はTikTok URLを設定する
func (s *SocialLinks) SetTikTok(url string) error {
	if url == "" {
		s.tiktok = nil
		return nil
	}
	if err := validateURL(url); err != nil {
		return err
	}
	if !strings.Contains(url, "tiktok.com") {
		return errors.New("無効なTikTok URLです")
	}
	s.tiktok = &url
	return nil
}

// SetYouTube はYouTube URLを設定する
func (s *SocialLinks) SetYouTube(url string) error {
	if url == "" {
		s.youtube = nil
		return nil
	}
	if err := validateURL(url); err != nil {
		return err
	}
	if !strings.Contains(url, "youtube.com") && !strings.Contains(url, "youtu.be") {
		return errors.New("無効なYouTube URLです")
	}
	s.youtube = &url
	return nil
}

// SetFacebook はFacebook URLを設定する
func (s *SocialLinks) SetFacebook(url string) error {
	if url == "" {
		s.facebook = nil
		return nil
	}
	if err := validateURL(url); err != nil {
		return err
	}
	if !strings.Contains(url, "facebook.com") {
		return errors.New("無効なFacebook URLです")
	}
	s.facebook = &url
	return nil
}

// SetOfficial は公式サイトURLを設定する
func (s *SocialLinks) SetOfficial(url string) error {
	if url == "" {
		s.official = nil
		return nil
	}
	if err := validateURL(url); err != nil {
		return err
	}
	s.official = &url
	return nil
}

// SetFanClub はファンクラブURLを設定する
func (s *SocialLinks) SetFanClub(url string) error {
	if url == "" {
		s.fanClub = nil
		return nil
	}
	if err := validateURL(url); err != nil {
		return err
	}
	s.fanClub = &url
	return nil
}

// validateURL は基本的なURL検証を行う
func validateURL(url string) error {
	if url == "" {
		return nil
	}
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return errors.New("URLはhttp://またはhttps://で始まる必要があります")
	}
	return nil
}
