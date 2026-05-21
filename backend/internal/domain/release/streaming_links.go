package release

import (
	"errors"
	"strings"
)

// StreamingLinks はリリースのストリーミングサービスリンク（値オブジェクト）
type StreamingLinks struct {
	spotify      *string // open.spotify.com
	appleMusic   *string // music.apple.com
	youtubeMusic *string // music.youtube.com
	youtube      *string // youtube.com / youtu.be（MV等）
	lineMusic    *string // music.line.me
	amazonMusic  *string // music.amazon.co.jp / amazon.com/music
	official     *string // 公式サイト（制限なし）
}

func NewStreamingLinks() *StreamingLinks {
	return &StreamingLinks{}
}

func (s *StreamingLinks) Spotify() *string      { return s.spotify }
func (s *StreamingLinks) AppleMusic() *string   { return s.appleMusic }
func (s *StreamingLinks) YouTubeMusic() *string { return s.youtubeMusic }
func (s *StreamingLinks) YouTube() *string      { return s.youtube }
func (s *StreamingLinks) LineMusic() *string    { return s.lineMusic }
func (s *StreamingLinks) AmazonMusic() *string  { return s.amazonMusic }
func (s *StreamingLinks) Official() *string     { return s.official }

func (s *StreamingLinks) SetSpotify(url string) error {
	if url == "" {
		s.spotify = nil
		return nil
	}
	if err := validateStreamingURL(url); err != nil {
		return err
	}
	if !strings.Contains(url, "spotify.com") {
		return errors.New("無効なSpotify URLです")
	}
	s.spotify = &url
	return nil
}

func (s *StreamingLinks) SetAppleMusic(url string) error {
	if url == "" {
		s.appleMusic = nil
		return nil
	}
	if err := validateStreamingURL(url); err != nil {
		return err
	}
	if !strings.Contains(url, "music.apple.com") {
		return errors.New("無効なApple Music URLです")
	}
	s.appleMusic = &url
	return nil
}

func (s *StreamingLinks) SetYouTubeMusic(url string) error {
	if url == "" {
		s.youtubeMusic = nil
		return nil
	}
	if err := validateStreamingURL(url); err != nil {
		return err
	}
	if !strings.Contains(url, "music.youtube.com") {
		return errors.New("無効なYouTube Music URLです")
	}
	s.youtubeMusic = &url
	return nil
}

func (s *StreamingLinks) SetYouTube(url string) error {
	if url == "" {
		s.youtube = nil
		return nil
	}
	if err := validateStreamingURL(url); err != nil {
		return err
	}
	if !strings.Contains(url, "youtube.com") && !strings.Contains(url, "youtu.be") {
		return errors.New("無効なYouTube URLです")
	}
	s.youtube = &url
	return nil
}

func (s *StreamingLinks) SetLineMusic(url string) error {
	if url == "" {
		s.lineMusic = nil
		return nil
	}
	if err := validateStreamingURL(url); err != nil {
		return err
	}
	if !strings.Contains(url, "music.line.me") {
		return errors.New("無効なLINE Music URLです")
	}
	s.lineMusic = &url
	return nil
}

func (s *StreamingLinks) SetAmazonMusic(url string) error {
	if url == "" {
		s.amazonMusic = nil
		return nil
	}
	if err := validateStreamingURL(url); err != nil {
		return err
	}
	if !strings.Contains(url, "amazon") {
		return errors.New("無効なAmazon Music URLです")
	}
	s.amazonMusic = &url
	return nil
}

func (s *StreamingLinks) SetOfficial(url string) error {
	if url == "" {
		s.official = nil
		return nil
	}
	if err := validateStreamingURL(url); err != nil {
		return err
	}
	s.official = &url
	return nil
}

func validateStreamingURL(url string) error {
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return errors.New("URLはhttp://またはhttps://で始まる必要があります")
	}
	return nil
}
