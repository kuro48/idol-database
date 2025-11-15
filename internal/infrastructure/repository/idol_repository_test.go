package repository

import (
	"testing"

	"github.com/kuro48/idol-api/internal/domain/models"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// 注意: これらのテストは統合テストとして別途実装することを推奨
// ここではデータ構造とロジックの基本検証のみ行う

func TestIdolRepositoryStructure(t *testing.T) {
	t.Run("NewIdolRepository creates repository", func(t *testing.T) {
		// この test は統合テストで MongoDB 接続が必要
		// 現在はスキップ
		t.Skip("Integration test - requires MongoDB connection")
	})
}

func TestIdolDataValidation(t *testing.T) {
	t.Run("valid idol data", func(t *testing.T) {
		idol := &models.Idol{
			ID:          bson.NewObjectID(),
			Name:        "テストアイドル",
			Group:       "テストグループ",
			Birthdate:   "2000-01-01",
			Nationality: "日本",
			ImageURL:    "https://example.com/image.jpg",
		}

		assert.NotEmpty(t, idol.ID)
		assert.Equal(t, "テストアイドル", idol.Name)
		assert.Equal(t, "テストグループ", idol.Group)
	})

	t.Run("update request structure", func(t *testing.T) {
		updateReq := &models.UpdateIdolRequest{
			Name:  "更新後の名前",
			Group: "更新後のグループ",
		}

		assert.Equal(t, "更新後の名前", updateReq.Name)
		assert.Equal(t, "更新後のグループ", updateReq.Group)
	})
}
