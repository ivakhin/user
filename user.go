package user

type (
	Users []User
	User  struct {
		ID    int    `bson:"_id"`
		Name  string `bson:"name"`
		Sex   Sex    `bson:"sex"`
		Phone int    `bson:"phone"`
	}
)

type Sex string

const (
	SexMale   Sex = "male"
	SexFemale Sex = "female"
)
