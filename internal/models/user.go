package models

type User struct {
	Id      int64
	Balance float64
}

func NewUser(userId int64) User {
	return User{Id: userId}
}

func (u User) UpdateBalance(newBalance float64) {
	u.Balance = newBalance
}

//func (u User) UpdateCity(newCity string) {
//	u.City = newCity
//}
//
//func (u User) UpdateSex(newSex string) {
//	u.Sex = newSex
//}
//
//func (u User) UpdateHeight(newHeight string) {
//	u.Name = newHeight
//}
