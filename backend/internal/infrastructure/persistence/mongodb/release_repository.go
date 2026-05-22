package mongodb

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/kuro48/idol-api/internal/domain/release"
	"github.com/kuro48/idol-api/internal/shared/audit"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// ReleaseRepository はMongoDBを使用したリリースリポジトリの実装
type ReleaseRepository struct {
	collection *mongo.Collection
}

func NewReleaseRepository(db *mongo.Database) *ReleaseRepository {
	return &ReleaseRepository{
		collection: db.Collection("releases"),
	}
}

// ---- ドキュメント構造 ----

type releaseDocument struct {
	ID             bson.ObjectID           `bson:"_id,omitempty"`
	Title          string                  `bson:"title"`
	ReleaseType    string                  `bson:"release_type"`
	ReleaseDate    time.Time               `bson:"release_date"`
	Artists        []artistRefDocument     `bson:"artists"`
	Tracks         []trackDocument         `bson:"tracks,omitempty"`
	StreamingLinks *streamingLinksDocument `bson:"streaming_links,omitempty"`
	ExternalIDs    map[string]string       `bson:"external_ids,omitempty"`
	CoverImageURL  *string                 `bson:"cover_image_url,omitempty"`
	Aliases        []string                `bson:"aliases,omitempty"`
	TagIDs         []string                `bson:"tag_ids,omitempty"`
	CreatedAt      time.Time               `bson:"created_at"`
	UpdatedAt      time.Time               `bson:"updated_at"`
	CreatedBy      string                  `bson:"created_by,omitempty"`
	UpdatedBy      string                  `bson:"updated_by,omitempty"`
	IsDeleted      bool                    `bson:"is_deleted,omitempty"`
	DeletedAt      *time.Time              `bson:"deleted_at,omitempty"`
	DeletedBy      string                  `bson:"deleted_by,omitempty"`
}

type artistRefDocument struct {
	Kind string `bson:"kind"`
	ID   string `bson:"id"`
	Role string `bson:"role,omitempty"`
}

type trackDocument struct {
	TrackNumber   int                        `bson:"track_number"`
	Title         string                     `bson:"title"`
	DurationSec   *int                       `bson:"duration_sec,omitempty"`
	ISRC          *string                    `bson:"isrc,omitempty"`
	CoverImageURL *string                    `bson:"cover_image_url,omitempty"`
	Participants  []trackParticipantDocument `bson:"participants,omitempty"`
}

type trackParticipantDocument struct {
	IdolID   string  `bson:"idol_id"`
	Status   string  `bson:"status"`
	Position *string `bson:"position,omitempty"`
}

type streamingLinksDocument struct {
	Spotify      *string `bson:"spotify,omitempty"`
	AppleMusic   *string `bson:"apple_music,omitempty"`
	YouTubeMusic *string `bson:"youtube_music,omitempty"`
	YouTube      *string `bson:"youtube,omitempty"`
	LineMusic    *string `bson:"line_music,omitempty"`
	AmazonMusic  *string `bson:"amazon_music,omitempty"`
	Official     *string `bson:"official,omitempty"`
}

// ---- toDocument ----

func toReleaseDocument(r *release.Release) (*releaseDocument, error) {
	objectID, err := bson.ObjectIDFromHex(r.ID().Value())
	if err != nil && r.ID().Value() != "" {
		return nil, fmt.Errorf("無効なリリースID %q: %w", r.ID().Value(), err)
	}

	artists := make([]artistRefDocument, len(r.Artists()))
	for i, a := range r.Artists() {
		artists[i] = artistRefDocument{Kind: string(a.Kind()), ID: a.ID(), Role: a.Role()}
	}

	tracks := make([]trackDocument, len(r.Tracks()))
	for i, t := range r.Tracks() {
		tracks[i] = trackDocument{
			TrackNumber:   t.TrackNumber(),
			Title:         t.Title(),
			DurationSec:   t.DurationSec(),
			ISRC:          t.ISRC(),
			CoverImageURL: t.CoverImageURL(),
			Participants:  toTrackParticipantDocuments(t.Participants()),
		}
	}

	var sl *streamingLinksDocument
	if links := r.StreamingLinks(); links != nil {
		sl = toStreamingLinksDocument(links)
	}

	var extIDs map[string]string
	if ids := r.ExternalIDs(); !ids.IsEmpty() {
		raw := ids.All()
		extIDs = make(map[string]string, len(raw))
		for k, v := range raw {
			extIDs[string(k)] = v
		}
	}

	return &releaseDocument{
		ID:             objectID,
		Title:          r.Title().Value(),
		ReleaseType:    r.ReleaseType().Value(),
		ReleaseDate:    r.ReleaseDate().Value(),
		Artists:        artists,
		Tracks:         tracks,
		StreamingLinks: sl,
		ExternalIDs:    extIDs,
		CoverImageURL:  r.CoverImageURL(),
		Aliases:        r.Aliases(),
		TagIDs:         r.TagIDs(),
		CreatedAt:      r.CreatedAt(),
		UpdatedAt:      r.UpdatedAt(),
	}, nil
}

func toTrackParticipantDocuments(participants []release.TrackParticipant) []trackParticipantDocument {
	if len(participants) == 0 {
		return nil
	}
	docs := make([]trackParticipantDocument, 0, len(participants))
	for _, p := range participants {
		docs = append(docs, trackParticipantDocument{
			IdolID:   p.IdolID(),
			Status:   p.Status().Value(),
			Position: p.Position(),
		})
	}
	return docs
}

func toStreamingLinksDocument(links *release.StreamingLinks) *streamingLinksDocument {
	return &streamingLinksDocument{
		Spotify:      links.Spotify(),
		AppleMusic:   links.AppleMusic(),
		YouTubeMusic: links.YouTubeMusic(),
		YouTube:      links.YouTube(),
		LineMusic:    links.LineMusic(),
		AmazonMusic:  links.AmazonMusic(),
		Official:     links.Official(),
	}
}

// ---- toDomain ----

func toReleaseDomain(doc *releaseDocument) (*release.Release, error) {
	id, err := release.NewReleaseID(doc.ID.Hex())
	if err != nil {
		return nil, err
	}
	title, err := release.NewReleaseTitle(doc.Title)
	if err != nil {
		return nil, err
	}
	releaseType, err := release.NewReleaseType(doc.ReleaseType)
	if err != nil {
		return nil, err
	}
	releaseDate := release.NewReleaseDate(doc.ReleaseDate)

	artists := make([]release.ArtistRef, 0, len(doc.Artists))
	for _, a := range doc.Artists {
		ref, err := release.NewArtistRef(release.ArtistKind(a.Kind), a.ID, a.Role)
		if err != nil {
			return nil, fmt.Errorf("アーティスト参照変換エラー: %w", err)
		}
		artists = append(artists, ref)
	}

	tracks := make([]release.Track, 0, len(doc.Tracks))
	for _, t := range doc.Tracks {
		participants, err := toTrackParticipantsDomain(t.Participants)
		if err != nil {
			return nil, fmt.Errorf("楽曲参加情報変換エラー: %w", err)
		}
		track, err := release.NewTrack(t.TrackNumber, t.Title, t.DurationSec, t.ISRC, t.CoverImageURL, participants)
		if err != nil {
			return nil, fmt.Errorf("楽曲変換エラー: %w", err)
		}
		tracks = append(tracks, track)
	}

	var streamingLinks *release.StreamingLinks
	if doc.StreamingLinks != nil {
		streamingLinks = toStreamingLinksDomain(doc.StreamingLinks)
	}

	var extIDs *release.ReleaseExternalIDs
	if len(doc.ExternalIDs) > 0 {
		typed := make(map[release.ReleaseExternalIDKind]string, len(doc.ExternalIDs))
		for k, v := range doc.ExternalIDs {
			typed[release.ReleaseExternalIDKind(k)] = v
		}
		extIDs = release.ReconstructReleaseExternalIDs(typed)
	}

	if doc.Aliases == nil {
		doc.Aliases = []string{}
	}
	if doc.TagIDs == nil {
		doc.TagIDs = []string{}
	}

	return release.Reconstruct(
		id, title, releaseType, releaseDate,
		artists, tracks, streamingLinks, extIDs,
		doc.CoverImageURL, doc.Aliases, doc.TagIDs,
		doc.CreatedAt, doc.UpdatedAt,
	), nil
}

func toTrackParticipantsDomain(docs []trackParticipantDocument) ([]release.TrackParticipant, error) {
	if docs == nil {
		return nil, nil
	}
	participants := make([]release.TrackParticipant, 0, len(docs))
	for _, doc := range docs {
		status, err := release.NewParticipationStatus(doc.Status)
		if err != nil {
			return nil, err
		}
		participant, err := release.NewTrackParticipant(doc.IdolID, status, doc.Position)
		if err != nil {
			return nil, err
		}
		participants = append(participants, participant)
	}
	return participants, nil
}

func toStreamingLinksDomain(doc *streamingLinksDocument) *release.StreamingLinks {
	links := release.NewStreamingLinks()
	if doc.Spotify != nil {
		_ = links.SetSpotify(*doc.Spotify)
	}
	if doc.AppleMusic != nil {
		_ = links.SetAppleMusic(*doc.AppleMusic)
	}
	if doc.YouTubeMusic != nil {
		_ = links.SetYouTubeMusic(*doc.YouTubeMusic)
	}
	if doc.YouTube != nil {
		_ = links.SetYouTube(*doc.YouTube)
	}
	if doc.LineMusic != nil {
		_ = links.SetLineMusic(*doc.LineMusic)
	}
	if doc.AmazonMusic != nil {
		_ = links.SetAmazonMusic(*doc.AmazonMusic)
	}
	if doc.Official != nil {
		_ = links.SetOfficial(*doc.Official)
	}
	return links
}

// ---- CRUD ----

func (r *ReleaseRepository) Save(ctx context.Context, rel *release.Release) error {
	doc, err := toReleaseDocument(rel)
	if err != nil {
		return fmt.Errorf("ドキュメント変換エラー: %w", err)
	}
	if doc.ID.IsZero() {
		doc.ID = bson.NewObjectID()
		doc.CreatedAt = time.Now()
		doc.UpdatedAt = time.Now()
		doc.CreatedBy = audit.ActorFrom(ctx)
		doc.UpdatedBy = audit.ActorFrom(ctx)

		newID, err := release.NewReleaseID(doc.ID.Hex())
		if err != nil {
			return fmt.Errorf("ID生成エラー: %w", err)
		}
		rel.SetID(newID)
	}
	if _, err := r.collection.InsertOne(ctx, doc); err != nil {
		return fmt.Errorf("リリースの保存エラー: %w", err)
	}
	return nil
}

func (r *ReleaseRepository) FindByID(ctx context.Context, id release.ReleaseID) (*release.Release, error) {
	objectID, err := bson.ObjectIDFromHex(id.Value())
	if err != nil {
		return nil, fmt.Errorf("無効なID形式: %w", err)
	}
	var doc releaseDocument
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID, "is_deleted": bson.M{"$ne": true}}).Decode(&doc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("リリースが見つかりません")
		}
		return nil, fmt.Errorf("リリース取得エラー: %w", err)
	}
	return toReleaseDomain(&doc)
}

func (r *ReleaseRepository) Update(ctx context.Context, rel *release.Release) error {
	objectID, err := bson.ObjectIDFromHex(rel.ID().Value())
	if err != nil {
		return fmt.Errorf("無効なID形式: %w", err)
	}
	doc, err := toReleaseDocument(rel)
	if err != nil {
		return fmt.Errorf("ドキュメント変換エラー: %w", err)
	}

	artists := make(bson.A, len(doc.Artists))
	for i, a := range doc.Artists {
		m := bson.M{"kind": a.Kind, "id": a.ID}
		if a.Role != "" {
			m["role"] = a.Role
		}
		artists[i] = m
	}
	tracks := make(bson.A, len(doc.Tracks))
	for i, t := range doc.Tracks {
		m := bson.M{"track_number": t.TrackNumber, "title": t.Title}
		if t.DurationSec != nil {
			m["duration_sec"] = t.DurationSec
		}
		if t.ISRC != nil {
			m["isrc"] = t.ISRC
		}
		if t.CoverImageURL != nil {
			m["cover_image_url"] = t.CoverImageURL
		}
		if len(t.Participants) > 0 {
			participants := make(bson.A, 0, len(t.Participants))
			for _, p := range t.Participants {
				pm := bson.M{"idol_id": p.IdolID, "status": p.Status}
				if p.Position != nil {
					pm["position"] = p.Position
				}
				participants = append(participants, pm)
			}
			m["participants"] = participants
		}
		tracks[i] = m
	}

	setFields := bson.M{
		"title":           doc.Title,
		"release_type":    doc.ReleaseType,
		"release_date":    doc.ReleaseDate,
		"artists":         artists,
		"tracks":          tracks,
		"cover_image_url": doc.CoverImageURL,
		"aliases":         doc.Aliases,
		"tag_ids":         doc.TagIDs,
		"updated_at":      time.Now(),
		"updated_by":      audit.ActorFrom(ctx),
	}
	if doc.StreamingLinks != nil {
		setFields["streaming_links"] = doc.StreamingLinks
	}
	if doc.ExternalIDs != nil {
		setFields["external_ids"] = doc.ExternalIDs
	}

	result, err := r.collection.UpdateOne(ctx, bson.M{"_id": objectID}, bson.M{"$set": setFields})
	if err != nil {
		return fmt.Errorf("リリース更新エラー: %w", err)
	}
	if result.MatchedCount == 0 {
		return errors.New("リリースが見つかりません")
	}
	return nil
}

func (r *ReleaseRepository) Delete(ctx context.Context, id release.ReleaseID) error {
	objectID, err := bson.ObjectIDFromHex(id.Value())
	if err != nil {
		return fmt.Errorf("無効なID形式: %w", err)
	}
	now := time.Now()
	result, err := r.collection.UpdateOne(ctx,
		bson.M{"_id": objectID, "is_deleted": bson.M{"$ne": true}},
		bson.M{"$set": bson.M{
			"is_deleted": true,
			"deleted_at": now,
			"deleted_by": audit.ActorFrom(ctx),
			"updated_at": now,
		}},
	)
	if err != nil {
		return fmt.Errorf("リリース削除エラー: %w", err)
	}
	if result.MatchedCount == 0 {
		return errors.New("リリースが見つかりません")
	}
	return nil
}

func (r *ReleaseRepository) Restore(ctx context.Context, id release.ReleaseID) error {
	objectID, err := bson.ObjectIDFromHex(id.Value())
	if err != nil {
		return fmt.Errorf("無効なID形式: %w", err)
	}
	now := time.Now()
	result, err := r.collection.UpdateOne(ctx,
		bson.M{"_id": objectID, "is_deleted": true},
		bson.M{
			"$set":   bson.M{"is_deleted": false, "updated_at": now, "updated_by": audit.ActorFrom(ctx)},
			"$unset": bson.M{"deleted_at": "", "deleted_by": ""},
		},
	)
	if err != nil {
		return fmt.Errorf("リリース復元エラー: %w", err)
	}
	if result.MatchedCount == 0 {
		return errors.New("削除済みリリースが見つかりません")
	}
	return nil
}

// ---- Search ----

func (r *ReleaseRepository) Search(ctx context.Context, criteria release.SearchCriteria) ([]*release.Release, error) {
	filter := buildReleaseFilter(criteria)
	opts := options.Find()

	sortOrder := 1
	if criteria.Order == "desc" {
		sortOrder = -1
	}
	sortField := criteria.Sort
	if sortField == "" {
		sortField = "release_date"
		sortOrder = -1
	}
	opts.SetSort(bson.D{{Key: sortField, Value: sortOrder}})
	opts.SetSkip(int64(criteria.Offset))
	if criteria.Limit > 0 {
		opts.SetLimit(int64(criteria.Limit))
	}

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("リリース検索エラー: %w", err)
	}
	defer cursor.Close(ctx)

	var docs []releaseDocument
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, fmt.Errorf("データ変換エラー: %w", err)
	}

	releases := make([]*release.Release, 0, len(docs))
	for _, doc := range docs {
		rel, err := toReleaseDomain(&doc)
		if err != nil {
			return nil, fmt.Errorf("ドメインモデル変換エラー: %w", err)
		}
		releases = append(releases, rel)
	}
	return releases, nil
}

func (r *ReleaseRepository) Count(ctx context.Context, criteria release.SearchCriteria) (int64, error) {
	return r.collection.CountDocuments(ctx, buildReleaseFilter(criteria))
}

func buildReleaseFilter(criteria release.SearchCriteria) bson.M {
	filter := bson.M{"is_deleted": bson.M{"$ne": true}}

	if criteria.Title != nil {
		titleRegex := bson.M{"$regex": safePartialMatchRegex(*criteria.Title), "$options": "i"}
		filter["$or"] = bson.A{
			bson.M{"title": titleRegex},
			bson.M{"aliases": titleRegex},
		}
	}
	if criteria.ReleaseType != nil {
		filter["release_type"] = criteria.ReleaseType.Value()
	}
	if criteria.ArtistID != nil {
		if criteria.ArtistKind != nil {
			filter["artists"] = bson.M{"$elemMatch": bson.M{
				"id":   *criteria.ArtistID,
				"kind": string(*criteria.ArtistKind),
			}}
		} else {
			filter["artists.id"] = *criteria.ArtistID
		}
	}
	if criteria.ReleaseDateFrom != nil || criteria.ReleaseDateTo != nil {
		dateFilter := bson.M{}
		if criteria.ReleaseDateFrom != nil {
			dateFilter["$gte"] = *criteria.ReleaseDateFrom
		}
		if criteria.ReleaseDateTo != nil {
			dateFilter["$lte"] = *criteria.ReleaseDateTo
		}
		filter["release_date"] = dateFilter
	}
	return filter
}

func (r *ReleaseRepository) FindByExternalID(ctx context.Context, kind release.ReleaseExternalIDKind, value string) (*release.Release, error) {
	field := "external_ids." + string(kind)
	var doc releaseDocument
	err := r.collection.FindOne(ctx, bson.M{field: value, "is_deleted": bson.M{"$ne": true}}).Decode(&doc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("外部ID検索エラー: %w", err)
	}
	return toReleaseDomain(&doc)
}

func (r *ReleaseRepository) FindByArtistID(ctx context.Context, artistID string) ([]*release.Release, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"artists.id": artistID, "is_deleted": bson.M{"$ne": true}})
	if err != nil {
		return nil, fmt.Errorf("アーティストIDによるリリース検索エラー: %w", err)
	}
	defer cursor.Close(ctx)

	var docs []releaseDocument
	if err := cursor.All(ctx, &docs); err != nil {
		return nil, fmt.Errorf("データ変換エラー: %w", err)
	}

	releases := make([]*release.Release, 0, len(docs))
	for _, doc := range docs {
		rel, err := toReleaseDomain(&doc)
		if err != nil {
			return nil, fmt.Errorf("ドメインモデル変換エラー: %w", err)
		}
		releases = append(releases, rel)
	}
	return releases, nil
}

// EnsureIndexes は検索パフォーマンス向上のためのインデックスを作成する
func (r *ReleaseRepository) EnsureIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "release_date", Value: -1}}},
		{Keys: bson.D{{Key: "release_type", Value: 1}}},
		{Keys: bson.D{{Key: "artists.id", Value: 1}}},
		{Keys: bson.D{{Key: "tracks.participants.idol_id", Value: 1}}},
		{Keys: bson.D{{Key: "title", Value: 1}}},
		{Keys: bson.D{{Key: "created_at", Value: -1}}},
		{Keys: bson.D{{Key: "tag_ids", Value: 1}}},
		{Keys: bson.D{{Key: "release_type", Value: 1}, {Key: "release_date", Value: -1}}},
	}

	for _, kind := range []string{"spotify_album_id", "apple_music_album_id", "upc", "jan_code"} {
		field := "external_ids." + kind
		indexes = append(indexes, mongo.IndexModel{
			Keys:    bson.D{{Key: field, Value: 1}},
			Options: options.Index().SetSparse(true).SetUnique(true).SetName("unique_ext_" + kind),
		})
	}

	if _, err := r.collection.Indexes().CreateMany(ctx, indexes); err != nil {
		return fmt.Errorf("インデックス作成エラー: %w", err)
	}
	return nil
}
