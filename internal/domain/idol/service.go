package idol

import (
	"context"
	"errors"
	"strings"
	"unicode/utf8"
)

// DomainService はアイドルドメインのドメインサービス
type DomainService struct {
	repository Repository
}

// NewDomainService はドメインサービスを作成する
func NewDomainService(repository Repository) *DomainService {
	return &DomainService{
		repository: repository,
	}
}

// CanCreate はアイドルを作成可能かを判定する
func (s *DomainService) CanCreate(ctx context.Context, name IdolName) error {
	exists, err := s.repository.ExistsByName(ctx, name)
	if err != nil {
		return err
	}

	if exists {
		return errors.New("同じ名前のアイドルが既に存在します")
	}

	return nil
}

// IsDuplicateName は名前の重複をチェックする
func (s *DomainService) IsDuplicateName(ctx context.Context, name IdolName, excludeID *IdolID) (bool, error) {
	idols, err := s.repository.FindAll(ctx)
	if err != nil {
		return false, err
	}

	for _, idol := range idols {
		// 除外するIDがある場合はスキップ
		if excludeID != nil && idol.ID().Equals(*excludeID) {
			continue
		}

		if idol.Name().Value() == name.Value() {
			return true, nil
		}
	}

	return false, nil
}

// DuplicateCandidate は重複候補
type DuplicateCandidate struct {
	Idol   *Idol
	Reason string // 重複と判断した理由
	Score  int    // 類似スコア（高いほど類似度が高い）
}

// FindDuplicateCandidates は指定したアイドルに対する重複候補を返す
// 以下の基準で判定:
// 1. 完全一致名（同姓同名、Score: 100）
// 2. 同じ誕生日（Score: 60）
// 3. 名前の部分一致（先頭2文字一致など、Score: 30）
func (s *DomainService) FindDuplicateCandidates(ctx context.Context, target *Idol) ([]*DuplicateCandidate, error) {
	all, err := s.repository.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	var candidates []*DuplicateCandidate
	seen := map[string]bool{}

	for _, other := range all {
		// 自分自身は除外
		if other.ID().Value() == target.ID().Value() {
			continue
		}

		score := 0
		var reasons []string

		// 完全一致名チェック
		if other.Name().Value() == target.Name().Value() {
			score += 100
			reasons = append(reasons, "名前が完全一致")
		}

		// 同じ誕生日チェック
		if target.Birthdate() != nil && other.Birthdate() != nil {
			if target.Birthdate().Value().Equal(other.Birthdate().Value()) {
				score += 60
				reasons = append(reasons, "誕生日が一致")
			}
		}

		// 名前の先頭2文字一致チェック
		if score == 0 { // 完全一致でない場合のみ
			targetName := target.Name().Value()
			otherName := other.Name().Value()
			if utf8.RuneCountInString(targetName) >= 2 && utf8.RuneCountInString(otherName) >= 2 {
				targetPrefix := string([]rune(targetName)[:2])
				otherPrefix := string([]rune(otherName)[:2])
				if targetPrefix == otherPrefix {
					score += 30
					reasons = append(reasons, "名前の先頭2文字が一致")
				}
			}
		}

		// 同じ事務所 + 名前の部分一致
		if target.AgencyID() != nil && other.AgencyID() != nil &&
			*target.AgencyID() == *other.AgencyID() {
			targetName := strings.ToLower(target.Name().Value())
			otherName := strings.ToLower(other.Name().Value())
			if strings.Contains(targetName, otherName) || strings.Contains(otherName, targetName) {
				score += 40
				reasons = append(reasons, "同一事務所かつ名前が部分一致")
			}
		}

		if score > 0 && !seen[other.ID().Value()] {
			seen[other.ID().Value()] = true
			reason := strings.Join(reasons, "、")
			candidates = append(candidates, &DuplicateCandidate{
				Idol:   other,
				Reason: reason,
				Score:  score,
			})
		}
	}

	// スコア降順にソート
	sortCandidatesByScore(candidates)

	return candidates, nil
}

func sortCandidatesByScore(candidates []*DuplicateCandidate) {
	for i := 1; i < len(candidates); i++ {
		for j := i; j > 0 && candidates[j].Score > candidates[j-1].Score; j-- {
			candidates[j], candidates[j-1] = candidates[j-1], candidates[j]
		}
	}
}
