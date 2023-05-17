package model

import "strings"

func IsCurseWord(word string) bool {
	// Define your list of curse words here
	// For example:
	curseWords := []string{"fuck", "shit", "asshole", "bitch"}

	for _, w := range curseWords {
		if strings.ToLower(word) == w {
			return true
		}
	}

	return false
}
