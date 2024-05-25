package models

type User struct {
	Id      int64
	Balance float64 `db:"bal"`
	Capital float64 `db:"capital"`
	Percent float64 `db:"percent"`
	Income  float64 `db:"income"`
}

func NewUser(userId int64) User {
	return User{Id: userId}
}

func (u User) UpdateUserId(userId int64) {
	u.Id = userId
}

func (u User) UpdateBalance(newBalance float64) {
	u.Balance = newBalance
}
