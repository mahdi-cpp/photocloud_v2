package user

import (
	"fmt"
	"os"
)

type UserService struct {
	Users map[int]User
}

func NewUserService(appDir string) *UserService {

	service := &UserService{
		Users: make(map[int]User),
	}

	err := service.loadUsers(appDir)
	if err != nil {
		fmt.Println("fail user service load users")
	}

	return service
}

func (service *UserService) GetDirectory(id int) string {
	return service.Users[id].FirstName
}

func (service *UserService) loadUsers(appDir string) error {

	// Scan app directory
	files, err := os.ReadDir(appDir)
	if err != nil {
		return err
	}

	i := 1 //create fake user id

	for _, file := range files {
		if !file.IsDir() {
			continue
		}

		fmt.Println("user service: ", i, file.Name())
		user := User{ID: i, FirstName: file.Name()}
		service.Users[user.ID] = user
		i++
	}

	return nil
}
