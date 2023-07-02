package models

func ListModels() []interface{} {
	return []interface{}{
		&User{},
		&Member{},
		&Band{},
		&OTP{},
	}
}
