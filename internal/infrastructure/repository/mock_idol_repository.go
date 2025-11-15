package repository

import (
	"github.com/kuro48/idol-api/internal/domain/models"
	domainRepo "github.com/kuro48/idol-api/internal/domain/repository"
	"github.com/stretchr/testify/mock"
)

// MockIdolRepository はテスト用のモックリポジトリ
type MockIdolRepository struct {
	mock.Mock
}

// インターフェース実装の確認
var _ domainRepo.IdolRepository = (*MockIdolRepository)(nil)

func (m *MockIdolRepository) Create(idol *models.Idol) error {
	args := m.Called(idol)
	return args.Error(0)
}

func (m *MockIdolRepository) FindAll() ([]models.Idol, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Idol), args.Error(1)
}

func (m *MockIdolRepository) FindByID(id string) (*models.Idol, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Idol), args.Error(1)
}

func (m *MockIdolRepository) Update(id string, update *models.UpdateIdolRequest) error {
	args := m.Called(id, update)
	return args.Error(0)
}

func (m *MockIdolRepository) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}
