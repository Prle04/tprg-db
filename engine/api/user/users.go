package user

import (
	"errors"
	"fmt"
	"tprgdb/components/hash"
	"tprgdb/components/serialization"
)

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func NewUser(username, password string) (*User, error) {
	user := &User{
		Username: username,
		Password: fmt.Sprintf("%x", hash.GetMD5Hash([]byte(password))),
	}
	if err := user.Validate(); err != nil {
		return nil, err
	}
	return user, nil
}

func NewEmptyUser() *User {
	return &User{
		Username: "",
		Password: "",
	}
}

func (u *User) Validate() error {
	if u.Username == "" {
		return errors.New("username cannot be empty")
	}
	if u.Password == "" {
		return errors.New("password cannot be empty")
	}
	return nil
}

func (u *User) SerializeToBytes() []byte {
	data := make([]byte, 0)
	data = append(data, serialization.UInt64ToBytes(uint64(len(u.Username)))...)
	data = append(data, []byte(u.Username)...)
	data = append(data, serialization.UInt64ToBytes(uint64(len(u.Password)))...)
	data = append(data, []byte(u.Password)...)
	return data
}

func (u *User) DeserializeFromBytes(data []byte) {
	usernameLen := serialization.BytesToUInt64(data[:8])
	u.Username = string(data[8 : 8+usernameLen])
	passwordLen := serialization.BytesToUInt64(data[8+usernameLen : 16+usernameLen])
	u.Password = string(data[16+usernameLen : 16+usernameLen+passwordLen])
}

func (u *User) Equals(other *User) bool {
	if u == nil || other == nil {
		return false
	}
	return u.Username == other.Username && u.Password == other.Password
}
