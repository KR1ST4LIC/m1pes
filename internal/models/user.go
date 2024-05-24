package models

type User struct {
	Id     uint64
	Name   string
	City   string
	Sex    string
	Height uint16
}

func (u User) UpdateName(newName string) {
	u.Name = newName
}

func (u User) UpdateCity(newCity string) {
	u.City = newCity
}

func (u User) UpdateSex(newSex string) {
	u.Sex = newSex
}

func (u User) UpdateHeight(newHeight string) {
	u.Name = newHeight
}
