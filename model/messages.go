package model

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"strings"

	"github.com/dgraph-io/badger/v3"
	"github.com/pkg/errors"
)

// Message represents a message sent by a user
type Message struct {
	Sender  string `json:"sender"`
	Message string `json:"message"`
}

// AddMessage adds a message to the database
func AddMessage(db *DB, message Message) error {
	buf := new(bytes.Buffer)
	err := gob.NewEncoder(buf).Encode(message)
	if err != nil {
		return err
	}
	err = db.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(message.Sender), buf.Bytes())
	})
	if err != nil {
		return err
	}
	return nil
}

// GetMessagesBySender retrieves all messages sent by a specific sender
func GetMessagesBySender(db *DB, sender string) ([]Message, error) {
	messages := []Message{}
	err := db.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		defer it.Close()
		prefix := []byte(sender)
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			value, err := item.ValueCopy(nil)
			if err != nil {
				return err
			}
			var message Message
			err = gob.NewDecoder(bytes.NewReader(value)).Decode(&message)
			if err != nil {
				return err
			}
			messages = append(messages, message)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return messages, nil
}

// Parse Message
func ParseMessage(tx []byte) (*Message, error) {
	fmt.Println("Inside Parse func")
	msg := &Message{}

	// Parse the message into key-value pairs
	pairs := strings.Split(string(tx), ",")

	if len(pairs) != 2 {
		return nil, errors.New("invalid number of key-value pairs in message")
	}

	for _, pair := range pairs {
		kv := strings.Split(pair, ":")

		if len(kv) != 2 {
			return nil, fmt.Errorf("invalid key-value pair in message: %s", pair)
		}

		key := kv[0]
		value := kv[1]

		switch key {
		case "sender":
			msg.Sender = value
		case "message":
			msg.Message = value
		default:
			return nil, fmt.Errorf("unknown key in message: %s", key)
		}
	}

	// Check if the message contains a sender and message
	if msg.Sender == "" {
		return nil, errors.New("message is missing sender")
	}
	if msg.Message == "" {
		return nil, errors.New("message is missing message")
	}
	fmt.Println("Exiting Parse func")

	return msg, nil
}
