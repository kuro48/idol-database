package repository

import "github.com/kuro48/idol-api/internal/domain/models"

// IdolRepository はアイドルリポジトリのインターフェース
type IdolRepository interface {
	Create(idol *models.Idol) error
	FindAll() ([]models.Idol, error)
	FindByID(id string) (*models.Idol, error)
	Update(id string, update *models.UpdateIdolRequest) error
	Delete(id string) error
}
