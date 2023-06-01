package model

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cometbft/cometbft/abci/types"
	"github.com/dgraph-io/badger/v3"
	"github.com/pkg/errors"
)

type DB struct {
	db *badger.DB
}

func (db *DB) Init(database *badger.DB) {
	db.db = database
}

func (db *DB) Commit() error {
	return db.db.Update(func(txn *badger.Txn) error {
		return txn.Commit()
	})
}

func NewDB(dbPath string) (*DB, error) {
	fmt.Println("New DB")
	// Open badger DB
	opts := badger.DefaultOptions(dbPath)
	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}

	// Create a new DB instance and initialize with badger DB
	dbInstance := &DB{}
	dbInstance.Init(db)

	return dbInstance, nil

}

func (db *DB) GetDB() *badger.DB {
	return db.db
}

func (db *DB) Size() int64 {
	lsm, vlog := db.GetDB().Size()
	return lsm + vlog
}

func (db *DB) CreateUser(user *User) error {
	// Check if the user already exists
	err := db.db.View(func(txn *badger.Txn) error {
		_, err := txn.Get([]byte(user.Name))
		return err
	})
	if err == nil {
		return errors.New("user already exists")
	}

	// Save the user to the database
	err = db.db.Update(func(txn *badger.Txn) error {
		userBytes, err := json.Marshal(user)
		if err != nil {
			return errors.Wrap(err, "failed to marshal user to JSON")
		}
		err = txn.Set([]byte(user.Name), userBytes)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (db *DB) FindUserByName(name string) (*User, error) {
	// Read the user from the database
	var user *User
	err := db.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(name))
		if err != nil {
			return err
		}
		err = item.Value(func(val []byte) error {
			return json.Unmarshal(val, &user)
		})
		return err
	})
	if err != nil {
		fmt.Println("Error in retrieveing user: ", err)
		return nil, err
	}

	return user, nil
}

func (db *DB) Set(key, value []byte) error {
	return db.db.Update(func(txn *badger.Txn) error {
		return txn.Set(key, value)
	})
}

func ViewDB(db *badger.DB, key []byte) ([]byte, error) {
	var value []byte
	err := db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			if err != badger.ErrKeyNotFound {
				return err
			}
			return nil
		}
		value, err = item.ValueCopy(nil)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return value, nil
}

func (db *DB) Close() error {
	return db.db.Close()
}

func (db *DB) Get(key []byte) ([]byte, error) {

	return ViewDB(db.db, key)

}
func (db *DB) GetValidators() (validators []types.ValidatorUpdate, err error) {
	err = db.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			var err error
			item := it.Item()
			k := item.Key()
			if isValidatorTx(k) {
				err := item.Value(func(v []byte) error {
					validator := new(types.ValidatorUpdate)
					err = types.ReadMessage(bytes.NewBuffer(v), validator)
					validators = append(validators, *validator)
					return err
				})
				if err != nil {
					return err
				}
			}
		}
		return nil
	})
	return
}
func isValidatorTx(tx []byte) bool {
	return strings.HasPrefix(string(tx), "val")
}
