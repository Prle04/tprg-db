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
		tr.Set(str, []byte(str))
		i += 3

	}

	tr.Print()

	fmt.Printf("\n\n")
	i = 2
	for i < 20 {
		str := fmt.Sprintf("%3d", i)
		tr.Set(str, []byte(str))

		i += 3

	}

	//Stavi dodatno funkciju ako vec postoji da samo izmeni
	str := fmt.Sprintf("%3d", 19)
	tr.Set(str, []byte(str))

	str = fmt.Sprintf("%3d", 4)
	tr.Set(str, []byte(str))

	str = fmt.Sprintf("%3d", 20)
	tr.Set(str, []byte(str))
	str = fmt.Sprintf("%3d", 21)
	tr.Set(str, []byte(str))
	str = fmt.Sprintf("%3d", 22)
	tr.Set(str, []byte(str))

	e := tr.Get(str)
	fmt.Println(e)

	str = fmt.Sprintf("%3d", 22)
	tr.Set(str, []byte("Lukaa"))

	e = tr.Get(str)
	fmt.Println(e)

	tr.Print()

	tr.Delete(str)

	fmt.Print("\n\n\n")
	tr.Print()

}
