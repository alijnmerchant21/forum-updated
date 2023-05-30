package test

import (
	"testing"

	"github.com/alijnmerchant21/forum-updated/model"
	"github.com/dgraph-io/badger/v3"
)

func TestAddAndGetMessages(t *testing.T) {
	// Open a Badger database
	opts := badger.DefaultOptions("").WithInMemory(true)
	db, err := badger.Open(opts)
	//require.NoError(t, err)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	// Create a model.DB instance using the Badger database
	modelDB := &model.DB{}
	modelDB.Init(db)

	// Add a message to the database
	message := &model.Message{
		Sender:  "alice",
		Message: "hello",
	}
	err = model.AddMessage(modelDB, *message)
	if err != nil {
		t.Fatalf("failed to add message: %v", err)
	}

	println("Message Added to DB")

	// Retrieve the message from the database
	messages, err := model.GetMessagesBySender(modelDB, "alice")
	if err != nil {
		t.Fatalf("failed to get messages: %v", err)
	}
	if len(messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(messages))
	}
	if messages[0].Sender != "alice" {
		t.Fatalf("expected sender to be alice, got %s", messages[0].Sender)
	}
	if messages[0].Message != "hello" {
		t.Fatalf("expected message to be hello, got %s", messages[0].Message)
	}

	println("Message is: ", message.Message)

	// Same process again

	// Add a message to the database
	message1 := &model.Message{
		Sender:  "alice",
		Message: "world",
	}
	err = model.AddMessage(modelDB, *message1)
	if err != nil {
		t.Fatalf("failed to add message: %v", err)
	}

	println("Message1 Added to DB")

	// Retrieve the message from the database
	messages1, err := model.GetMessagesBySender(modelDB, "alice")
	if err != nil {
		t.Fatalf("failed to get messages: %v", err)
	}
	if len(messages1) != 1 {
		t.Fatalf("expected 1 message, got %d", len(messages1))
	}
	if messages1[0].Sender != "alice" {
		t.Fatalf("expected sender to be alice, got %s", messages1[0].Sender)
	}
	if messages1[0].Message != "world" {
		t.Fatalf("expected message to be world, got %s", messages1[0].Message)
	}

	println("Message1 is: ", message1.Message)
}

func TestParseMessage(t *testing.T) {

	// Define a test message
	testMessage := []byte("sender:alice,message:hello")

	// Parse the test message
	message, err := model.ParseMessage(testMessage)
	if err != nil {
		t.Fatalf("failed to parse message: %v", err)
	}

	println("Parsed message is: ", message.Message)

}

/*func TestAddMessageEncoding(t *testing.T) {
	// Open a Badger database
	opts := badger.DefaultOptions("").WithInMemory(true)
	db, err := badger.Open(opts)
	//require.NoError(t, err)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	// Create a model.DB instance using the Badger database
	modelDB := &model.DB{}
	modelDB.Init(db)

	// Add a message to the database
	message := model.Message{
		Sender:  "alice",
		Message: "hello",
	}
	err = model.AddMessage(modelDB, message)
	if err != nil {
		t.Fatalf("failed to add message: %v", err)
	}

	// Retrieve the message from the database and decode it
	err = modelDB.GetDB().View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("alice"))
		if err != nil {
			return err
		}
		var decodedMessage model.Message
		err = item.Value(func(val []byte) error {
			return gob.NewDecoder(bytes.NewReader(val)).Decode(&decodedMessage)
		})
		if err != nil {
			return err
		}
		if decodedMessage.Sender != "alice" {
			t.Fatalf("expected sender to be alice, got %s", decodedMessage.Sender)
		}
		if decodedMessage.Message != "hello" {
			t.Fatalf("expected message to be hello, got %s", decodedMessage.Message)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("failed to retrieve message from database: %v", err)
	}
}*/
