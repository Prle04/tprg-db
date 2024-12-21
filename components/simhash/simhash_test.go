package simhash

import (
	"bufio"
	"fmt"
	"os"
	"testing"
)

func readFile(path string) string {
	file, err := os.Open(path)

	if err != nil {
		panic(err)
	}

	scanner := bufio.NewScanner(file)

	scanner.Split(bufio.ScanLines)

	content := ""

	for scanner.Scan() {
		content += scanner.Text()
	}

	file.Close()
	return content
}

func TestSimhash(t *testing.T) {
	sh1 := Create(readFile("./text1.txt"))
	sh2 := Create(readFile("./text2.txt"))
	sh3 := Create(readFile("./text3.txt"))

	fp1 := sh1.fingerprint
	fp2 := sh2.fingerprint
	fp3 := sh3.fingerprint

	t.Log("sh1 fingerprint: ", fp1)
	t.Log("sh2 fingerprint: ", fp2)
	t.Log("sh3 fingerprint: ", fp3)

	if fp1 == 0 {
		t.Errorf("Got 0 for fp1")
	}
	if fp2 == 0 {
		t.Errorf("Got 0 for fp2")
	}
	if fp3 == 0 {
		t.Errorf("Got 0 for fp3")
	}

	dist12 := HammingDistance(fp1, fp2)
	dist13 := HammingDistance(fp1, fp3)

	fmt.Println(dist12)
	fmt.Println(dist13)

}
