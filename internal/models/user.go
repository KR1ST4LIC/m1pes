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
