package storage

import (
	"fmt"
	"github.com/mahdi-cpp/photocloud_v2/internal/domain/model"
)

type UserManager struct {
	dir      string
	users    map[int]model.User
	metadata *MetadataControl[model.UserCollection]
}

func NewUserManager(dir string) *UserManager {

	manager := &UserManager{
		dir:      dir,
		users:    make(map[int]model.User),
		metadata: NewMetadataControl[model.UserCollection](dir),
	}

	users, err := manager.load()
	if err != nil {
		fmt.Println("fail user manager load users")
	}

	for _, user := range users {
		manager.users[user.ID] = user
		//fmt.Println(user.Username)
	}

	return manager
}

func (manager *UserManager) GetById(id int) model.User {
	return manager.users[id]
}

func (manager *UserManager) GetUsername(id int) string {
	return manager.users[id].Username
}

func (manager *UserManager) load() ([]model.User, error) {

	userCollection, err := manager.metadata.Read()
	if err != nil {
		return nil, err
	}

	var result []model.User
	for _, user := range userCollection.Users {
		result = append(result, user)
	}

	return result, nil
}
