package model

import (
	"bytes"
	"compress/flate"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/cometbft/cometbft/crypto/ed25519"
)

type Messages struct {
	db *DB
}

type Message struct {
	Sender  string `json:"sender"`
	Message string `json:"message"`
}

/*func UnmarshalMessage(tx []byte) (*Message, error) {
	if len(tx) == 0 {
		return nil, fmt.Errorf("tx failes")
	}
	var msg Message
	err := json.Unmarshal(tx, &msg)
	if err != nil {
		return nil, err
	}
	return &msg, nil
}*/

func UnmarshalMessage(tx []byte) (*Message, error) {
	// Log the transaction message
	log.Printf("Unmarshaling transaction message: %v\n", string(tx))

	if len(tx) == 0 {
		return nil, fmt.Errorf("transaction message is empty")
	}
	var msg Message
	err := json.Unmarshal(tx, &msg)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal transaction message: %v", err)
	}
	return &msg, nil
}

func ParseTransactionMessage(tx string) (*Message, error) {
	// Split the transaction message string using the separator character $
	txArray := strings.Split(tx, ",")

	// Initialize a new Message struct to hold the extracted fields
	msg := &Message{}

	// Loop through each field in the transaction message
	for _, item := range txArray {
		// Trim leading and trailing whitespaces.
		item = strings.TrimSpace(item)

		// Split the field into a key-value pair using the equals character =
		itemSplit := strings.Split(item, "=")
		if len(itemSplit) != 2 {
			//fmt.Printf("invalid transaction message field: %s\n", item)
			return nil, fmt.Errorf("invalid transaction message field: %s", item)
		}

		// Extract the key and value from the key-value pair
		key := itemSplit[0]
		value := itemSplit[1]

		// Assign the value to the appropriate field in the Message struct
		switch key {
		case "sender":
			msg.Sender = key
		case "message":
			msg.Message = value
		default:
			///fmt.Printf("unknown transaction message field: %s\n", key)
			return nil, fmt.Errorf("unknown transaction message field: %s", key)
		}
	}

	// Return the parsed Message struct
	return msg, nil
}

/*func ParseTransactionMessage(b []byte) (*Message, error) {
	// Convert the byte slice to a string
	tx := string(b)

	// Split the transaction message string using the separator character $
	txArray := strings.Split(tx, ",")

	// Initialize a new Message struct to hold the extracted fields
	msg := &Message{}

	// Loop through each field in the transaction message
	for _, item := range txArray {
		// Trim leading and trailing whitespaces.
		item = strings.TrimSpace(item)

		// Split the field into a key-value pair using either the equals character = or the colon character :
		var itemSplit []string
		if strings.Contains(item, "=") {
			itemSplit = strings.Split(item, "=")
		} else if strings.Contains(item, ":") {
			itemSplit = strings.Split(item, ":")
		} else {
			return nil, fmt.Errorf("invalid transaction message field: %s", item)
		}

		// Extract the key and value from the key-value pair
		key := itemSplit[0]
		value := itemSplit[1]

		// Assign the value to the appropriate field in the Message struct
		switch key {
		case "sender":
			msg.Sender = value
		case "message":
			msg.Message = value
		default:
			fmt.Printf("unknown transaction message field: %s\n", key)
			return nil, fmt.Errorf("unknown transaction message field: %s", key)
		}
	}

	// Return the parsed Message struct
	return msg, nil
}*/

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

func (m *Messages) GetMessage(pubkey ed25519.PubKey) (*Message, error) {
	var msg Message
	err := m.db.GetByPubKey(pubkey, &msg)
	if err != nil {
		return nil, err
	}
	return &msg, nil
}

func (db *DB) AddMessage(msg *Message) error {
	msgBytes, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %v", err)
	}

	// Compress the message data using flate
	var buf bytes.Buffer
	gz, err := flate.NewWriter(&buf, flate.BestCompression)
	if err != nil {
		return fmt.Errorf("failed to create gzip writer: %v", err)
	}
	if _, err := gz.Write(msgBytes); err != nil {
		return fmt.Errorf("failed to compress message: %v", err)
	}
	if err := gz.Close(); err != nil {
		return fmt.Errorf("failed to close gzip writer: %v", err)
	}

	// Store the compressed message data in the database
	key := msg.Sender
	if err := db.store.Insert([]byte(key), buf.Bytes()); err != nil {
		return fmt.Errorf("failed to store message: %v", err)
	}

	return nil
}
