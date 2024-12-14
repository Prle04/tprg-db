package main

import (
	"fmt"
	"tprgdb/components/btree"
)

func main() {
	fmt.Printf("\n\n")
	tr := btree.NewBTree(4)
	i := 0
	for i < 20 {
		str := fmt.Sprintf("%3d", i)
		tr.Push(str, []byte(str))
		i += 3

	}

	tr.Print()

	fmt.Printf("\n\n")
	i = 2
	for i < 20 {
		str := fmt.Sprintf("%3d", i)
		tr.Push(str, []byte(str))

		i += 3

	}

	//Stavi dodatno funkciju ako vec postoji da samo izmeni
	str := fmt.Sprintf("%3d", 19)
	tr.Push(str, []byte(str))

	str = fmt.Sprintf("%3d", 4)
	tr.Push(str, []byte(str))

	str = fmt.Sprintf("%3d", 20)
	tr.Push(str, []byte(str))
	str = fmt.Sprintf("%3d", 21)
	tr.Push(str, []byte(str))
	str = fmt.Sprintf("%3d", 22)
	tr.Push(str, []byte(str))

	e := tr.Find(str)
	fmt.Println(e)

	str = fmt.Sprintf("%3d", 22)
	tr.Push(str, []byte("Lukaa"))

	e = tr.Find(str)
	fmt.Println(e)

	tr.Print()

}
