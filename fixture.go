package user

import (
	"math/rand"
)

func randUser(id int) User {
	s := randSex()

	return User{
		ID:    id,
		Name:  randName(s),
		Sex:   s,
		Phone: randPhone(),
	}
}

func randSex() Sex {
	if rand.Intn(2) == 0 { //nolint:gomnd
		return SexMale
	}

	return SexFemale
}

func randName(s Sex) string {
	if s == SexMale {
		return male[rand.Intn(len(male))]
	}

	return female[rand.Intn(len(female))]
}

func randPhone() int {
	return 900_000_00_00 + rand.Intn(100_000_00_00) //nolint:gomnd
}
