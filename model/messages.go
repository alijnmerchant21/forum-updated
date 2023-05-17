package model

import (
	"fmt"
	"strings"

	"github.com/cometbft/cometbft/crypto/ed25519"
)

type Messages struct {
	db *DB
}

type MsgSendMessage struct {
	Text string
	From ed25519.PubKey
}

type MsgSetBan struct {
	User  ed25519.PubKey
	State bool
}

type MsgSetModerator struct {
	User  ed25519.PubKey
	State bool
}

type MsgCreateUser struct {
	User User
}

func NewMessages(db *DB) *Messages {
	return &Messages{db: db}
}

func (m *Messages) SendMessage(text string, from ed25519.PubKey) error {
	msg := MsgSendMessage{
		Text: text,
		From: from,
	}
	return m.Process(msg)
}

func (m *Messages) SetBan(user ed25519.PubKey, state bool) error {
	msg := MsgSetBan{
		User:  user,
		State: state,
	}
	return m.Process(msg)
}

func (m *Messages) SetModerator(user ed25519.PubKey, state bool) error {
	msg := MsgSetModerator{
		User:  user,
		State: state,
	}
	return m.Process(msg)
}

func (m *Messages) CreateUser(user User) error {
	msg := MsgCreateUser{
		User: user,
	}
	return m.Process(msg)
}

// Define the Process method to implement the MessageProcessor interface
func (m *Messages) Process(message interface{}) error {
	switch msg := message.(type) {

	case MsgSendMessage:
		u, err := m.db.FindUser(msg.From)
		if err != nil {
			return err
		}

		words := strings.Split(msg.Text, " ")
		for _, word := range words {
			if IsCurseWord(word) {
				// User sent a curse word, ban the user
				err = m.SetBan(msg.From, true)
				if err != nil {
					return err
				}
				return fmt.Errorf("user banned for using curse word: %s", word)
			}
		}

		// No curse words found, process the message normally
		u.Version++
		u.NumMessages++
		return m.db.SaveUser(u)

	case MsgSetBan:
		u, err := m.db.FindUser(msg.User)
		if err != nil {
			return err
		}

		u.Version++
		u.Banned = msg.State
		return m.db.SaveUser(u)

	case MsgSetModerator:
		u, err := m.db.FindUser(msg.User)
		if err != nil {
			return err
		}

		u.Version++
		u.Moderator = msg.State
		return m.db.SaveUser(u)

	case MsgCreateUser:
		msg.User.Version = 0
		return m.db.CreateUser(&msg.User)

	default:
		return fmt.Errorf("message type %T not supported", message)
	}
}
