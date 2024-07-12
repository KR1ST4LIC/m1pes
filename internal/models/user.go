package models

type User struct {
	Id               int64
	USDTBalance      float64
	Capital          float64
	Percent          float64
	Income           float64
	ApiKey           string
	SecretKey        string
	Status           string
	TradingActivated bool
	Buy              bool
}

func NewUser(userId int64) User {
	return User{Id: userId}
}

func (u User) UpdateUserId(userId int64) {
	u.Id = userId
}

func (u User) UpdateBalance(newBalance float64) {
	u.USDTBalance = newBalance
}
