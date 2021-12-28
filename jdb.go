package jdb

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
)

const (
	ActionAdd       = "a"
	ActionUpdate    = "u"
	ActionDelete    = "d"
	ActionSeparator = ":"
)

type Item interface {
	New() Item
	GetId() int
	SetId(int)
}

type Handler func(Item)

type Database struct {
	nextId   int
	items    map[int]Item
	proto    Item
	journal  *os.File
	isClosed bool
}

func Open(file string, proto Item) (db *Database) {
	db = &Database{
		nextId: 1,
		items:  map[int]Item{},
		proto:  proto,
	}
	var err error
	db.journal, err = os.OpenFile(file, os.O_RDWR|os.O_CREATE, 0664)
	if err != nil {
		log.Fatal(err)
	}
	db.readJournal()
	return
}

func (db *Database) Close() {
	db.isClosed = true
	_, err := db.journal.Seek(0, 0)
	if err != nil {
		log.Print(err)
		return
	}
	err = db.journal.Truncate(0)
	if err != nil {
		log.Print(err)
		return
	}
	for _, item := range db.items {
		db.writeJournal(ActionAdd, item)
	}
}

func (db *Database) Get(id int) Item {
	if db.isClosed {
		return nil
	}
	return db.items[id]
}

func (db *Database) Add(item Item) {
	if item == nil || db.isClosed {
		return
	}
	item.SetId(db.nextId)
	db.writeJournal(ActionAdd, item)
	db.items[db.nextId] = item
	db.nextId++
}

func (db *Database) Update(item Item) {
	if item == nil || db.isClosed {
		return
	}
	db.writeJournal(ActionUpdate, item)
}

func (db *Database) Delete(item Item) {
	if item == nil || db.isClosed {
		return
	}
	db.writeJournal(ActionDelete, item)
	delete(db.items, item.GetId())
}

func (db *Database) ForEach(handler Handler) {
	if db.isClosed {
		return
	}
	for _, item := range db.items {
		handler(item)
	}
}

func (db *Database) readJournal() {
	scanner := bufio.NewScanner(db.journal)
	for scanner.Scan() {
		text := scanner.Text()
		fields := strings.SplitN(text, ActionSeparator, 2)
		item := db.proto.New()
		err := json.Unmarshal([]byte(fields[1]), item)
		if err != nil {
			log.Print(err)
			return
		}
		id := item.GetId()
		switch fields[0] {
		case ActionAdd, ActionUpdate:
			if id == 0 {
				id = db.nextId
			}
			if id >= db.nextId {
				db.nextId = id + 1
			}
			db.items[id] = item
		case ActionDelete:
			delete(db.items, id)
		}
	}
	if err := scanner.Err(); err != nil {
		log.Print(err)
		return
	}
}

func (db *Database) writeJournal(action string, item Item) {
	bytes, err := json.Marshal(item)
	if err != nil {
		log.Print(err)
		return
	}
	_, err = fmt.Fprintf(db.journal, "%s%s%s\n", action, ActionSeparator, string(bytes))
	if err != nil {
		log.Print(err)
		return
	}
}
