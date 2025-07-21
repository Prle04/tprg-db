package main

import (
	"fmt"
	"tprgdb/engine"
)

func main() {
	engine, err := engine.NewKVDatabase("tprgdb", "admin", "123")
	if err != nil {
		panic(err)
	}
	defer engine.Shutdown()

	for i := 0; i < 73; i++ {
		engine.Set(fmt.Sprintf("key-%03d", i), []byte(fmt.Sprintf("key-%03d", i)))
	}

	fmt.Println("RANGE SCAN")
	cursor, err := engine.RangeScan("key-040", "key-065", 10)
	if err != nil {
		panic(err)
	}
	pageCount := 0
	for {
		pageCount++
		page, err := cursor.NextPage()
		if err != nil {
			panic(err)
		}
		if len(page) == 0 {
			break
		}
		fmt.Printf("Page %d:\n", pageCount)
		for _, item := range page {
			fmt.Printf("key: %s, value: %s\n", item.Key, item.Value)
		}
	}
	fmt.Println("PREFIX SCAN")
	cursor, err = engine.PrefixScan("key-", 10)
	if err != nil {
		panic(err)
	}
	pageCount = 0
	for {
		pageCount++
		page, err := cursor.NextPage()
		if err != nil {
			panic(err)
		}
		if len(page) == 0 {
			break
		}
		fmt.Printf("Page %d:\n", pageCount)
		for _, item := range page {
			fmt.Printf("key: %s, value: %s\n", item.Key, item.Value)
		}
	}
	fmt.Println("RANGE ITERATE")
	it, err := engine.RangeIterate("key-040", "key-065")
	if err != nil {
		panic(err)
	}
	for {
		item, err := it.Next()
		if err != nil {
			break
		}
		if item == nil {
			break
		}
		fmt.Printf("key: %s, value: %s\n", item.GetKey(), item.GetValue())
	}
	fmt.Println("PREFIX ITERATE")
	it, err = engine.PrefixIterate("key-02")
	if err != nil {
		panic(err)
	}
	for {
		item, err := it.Next()
		if err != nil {
			break
		}
		if item == nil {
			break
		}
		fmt.Printf("key: %s, value: %s\n", item.GetKey(), item.GetValue())
	}
}
