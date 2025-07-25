package happle_models

import "time"

func (a *User) GetID() int                      { return a.ID }
func (a *User) SetID(id int)                    { a.ID = id }
func (a *User) SetCreationDate(t time.Time)     { a.CreationDate = t }
func (a *User) SetModificationDate(t time.Time) { a.ModificationDate = t }
func (a *User) GetCreationDate() time.Time      { return a.CreationDate }
func (a *User) GetModificationDate() time.Time  { return a.ModificationDate }

type User struct {
	ID               int       `json:"id"`
	Username         string    `json:"username"`
	PhoneNumber      string    `json:"phoneNumber"`
	Email            string    `json:"email"`
	FirstName        string    `json:"firstName"`
	LastName         string    `json:"lastName"`
	Bio              string    `json:"bio"`
	AvatarURL        string    `json:"avatarURL"`
	IsOnline         bool      `json:"isOnline"`
	LastSeen         time.Time `json:"lastSeen"`
	CreationDate     time.Time `json:"creationDate"`
	ModificationDate time.Time `json:"modificationDate"`
}

type UserHandler struct {
	ID          int    `json:"id"`
	Username    string `json:"username,omitempty"`
	PhoneNumber string `json:"phoneNumber,omitempty"`
	FirstName   string `json:"firstName,omitempty"`
	LastName    string `json:"lastName,omitempty"`
	Email       string `json:"email,omitempty"`
	AvatarURL   string `json:"avatarURL,omitempty"`
	IsOnline    *bool  `json:"isOnline,omitempty"`
}

func UpdateUser(user *User, handler UserHandler) *User {

	if handler.Username != "" {
		user.Username = handler.Username
	}
	if handler.PhoneNumber != "" {
		user.PhoneNumber = handler.PhoneNumber
	}
	if handler.FirstName != "" {
		user.FirstName = handler.FirstName
	}
	if handler.LastName != "" {
		user.LastName = handler.LastName
	}
	if handler.Email != "" {
		user.Email = handler.Email
	}
	if handler.AvatarURL != "" {
		user.AvatarURL = handler.AvatarURL
	}
	if handler.IsOnline != nil {
		user.IsOnline = *handler.IsOnline
	}

	return user
}
