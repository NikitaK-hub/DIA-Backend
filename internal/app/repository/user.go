package repository

import (
	"DIA_Backend/internal/app/ds"
	"errors"

	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

func (r *UserRepository) CreateUser(user *ds.User) error {
	var existing ds.User
	err := r.db.Where("username = ?", user.Username).First(&existing).Error
	if err == nil {
		return errors.New("username already exists")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	return r.db.Create(user).Error
}

func (r *UserRepository) GetUserByID(id uint64) (*ds.User, error) {
	var user ds.User
	err := r.db.Select("id, username, is_mod").First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetUserByUsername(username string) (*ds.User, error) {
	var user ds.User
	err := r.db.Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) UpdateUser(id uint64, username *string, password *string) error {
	updates := make(map[string]interface{})

	if username != nil {
		var existing ds.User
		err := r.db.Where("username = ? AND id != ?", *username, id).First(&existing).Error
		if err == nil {
			return errors.New("username already taken")
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		updates["username"] = *username
	}

	if password != nil {
		updates["password"] = *password
	}

	if len(updates) == 0 {
		return nil
	}

	return r.db.Model(&ds.User{}).Where("id = ?", id).Updates(updates).Error
}

func (r *UserRepository) AuthenticateUser(username, password string) (*ds.User, error) {
	var user ds.User
	err := r.db.Where("username = ? AND password = ?", username, password).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid credentials")
		}
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) UserExists(id uint64) bool {
	var count int64
	r.db.Model(&ds.User{}).Where("id = ?", id).Count(&count)
	return count > 0
}
