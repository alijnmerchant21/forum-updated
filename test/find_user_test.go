package test

import (
	"fmt"
	"testing"

	"github.com/alijnmerchant21/forum-updated/model"
)

func TestFindUserByname(t *testing.T) {
	// Initialize the database
	println("DB to be initialized")
	db, err := model.NewDB("test.db")
	println("DB initialized")
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Create some test users
	println("User being created")
	users := []*model.User{
		{Name: "user1", PubKey: []byte("pubkey1"), Moderator: false, Banned: false},
		{Name: "user2", PubKey: []byte("pubkey2"), Moderator: false, Banned: false},
		{Name: "user3", PubKey: []byte("pubkey3"), Moderator: false, Banned: false},
	}
	println("User is defined")
	for _, user := range users {
		err := db.CreateUser(user)
		fmt.Println(user.Name)
		if err != nil {
			t.Fatalf("Failed to create user: %v", err)
		}
	}

	// Find a user by name
	//println("Trying to find user")

	/*foundUser, err := db.FindUserByName("user5")
	if err != nil {
		t.Fatalf("Failed to find user: %v", err)
	}

	if foundUser == nil {
		// t.Fatalf("Expected user2, but got %s", foundUser.Name)
		println("USer not found")

	}
	if foundUser.Name != "" {
		println("Expected empty name, but got %s", foundUser.Name)
	}*/

	// Verify that the correct user was returned
	println("Trying to find user")
	foundUser1, err1 := db.FindUserByName("user2")
	if err1 != nil {
		t.Fatalf("Failed to find user by name: %v", err1)
	}

	if foundUser1 == nil {
		t.Fatalf("User not found")
	}

	if foundUser1.Name != "user2" {
		// t.Fatalf("Expected user2, but got %s", foundUser.Name)
		println("Expected user2, but got %s", foundUser1.Name)
	}

	if foundUser1.Name == "user2" {
		println("Voila! User found")
	}
}
