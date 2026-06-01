package song

import "context"

// SongUseCase はプレゼンテーション層が呼び出すインターフェース
type SongUseCase interface {
	CreateSong(ctx context.Context, cmd CreateSongCommand) (*SongDTO, error)
	GetSong(ctx context.Context, query GetSongQuery) (*SongDTO, error)
	ListSongs(ctx context.Context, query ListSongQuery) (*SongSearchResult, error)
	UpdateSong(ctx context.Context, cmd UpdateSongCommand) error
	DeleteSong(ctx context.Context, cmd DeleteSongCommand) error
}
