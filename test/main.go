package main

import (
	"fmt"
	"os"

	"github.com/leftslash/jdb"
)

const file = "items.db"

type Item struct {
	Id    int
	Value string
}

func (i *Item) New() jdb.Item { return new(Item) }
func (i *Item) GetId() int    { return i.Id }
func (i *Item) SetId(id int)  { i.Id = id }

func main() {
	data := `a:{"Id":1,"Value":"a"}
a:{"Id":2,"Value":"b"}
a:{"Id":3,"Value":"c"}
d:{"Id":1,"Value":"a"}
u:{"Id":2,"Value":"B"}
`
	os.WriteFile(file, []byte(data), 0664)
	db := jdb.Open(file, &Item{})
	defer db.Close()
	db.ForEach(func(i jdb.Item) {
		item := i.(*Item)
		fmt.Printf("%#v\n", item)
	})
	item := Item{Value: "d"}
	db.Add(&item)
	db.Add(&Item{Value: "e"})
	db.Delete(&item)
	item = Item{Value: "f"}
	db.Add(&item)
	item.Value = "F"
	db.Update(&item)
	db.Add(&Item{Value: "g"})
	fmt.Println()
	db.ForEach(func(i jdb.Item) {
		item := i.(*Item)
		fmt.Printf("%#v\n", item)
	})
	fmt.Println()
	fmt.Printf("%#v\n", db.Get(7))
	fmt.Printf("%#v\n", db.Get(8))

}
