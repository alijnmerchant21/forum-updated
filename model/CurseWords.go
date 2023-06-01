package model

import "strings"

func IsCurseWord(word string, curseWords string) bool {
	// Define your list of curse words here
	// For example:
	return strings.Contains(curseWords, word)
}
