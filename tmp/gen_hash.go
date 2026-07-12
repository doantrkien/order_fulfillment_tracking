package main

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	// The old hash from seed
	oldHash := "$2a$10$w/lW8qT/51/AwaF877oIL.k9wykL.PMMDe1kFQpfmpQeW0I2URgIi"
	// The new hash we just set
	newHash := "$2a$10$jIvB4qNgi.fWEdjIq0kRSe9u/DBgFnCHz3X7uYqyZflty8HKCs6o6"

	passwords := []string{"password", "123456", "12345", "admin", "password123"}
	for _, pw := range passwords {
		err := bcrypt.CompareHashAndPassword([]byte(oldHash), []byte(pw))
		if err == nil {
			fmt.Printf("OLD HASH matches: %s\n", pw)
		}
	}
	
	err := bcrypt.CompareHashAndPassword([]byte(newHash), []byte("password"))
	if err == nil {
		fmt.Println("NEW HASH: 'password' matches correctly!")
	} else {
		fmt.Printf("NEW HASH error: %v\n", err)
	}
}
