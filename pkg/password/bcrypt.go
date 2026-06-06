package password

import "golang.org/x/crypto/bcrypt"

func Hash(value string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(value), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}

func Compare(hashedValue, plainValue string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedValue), []byte(plainValue))
}
