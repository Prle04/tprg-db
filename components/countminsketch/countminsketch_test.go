package countminsketch

import (
	"bytes"
	"testing"
)

func TestSave(t *testing.T) {

	var k uint32 = 100
	var m uint32 = 100

	//t.Fatalf("faild")
	cms := New(k, m)
	buffer := bytes.NewBuffer(nil)
	err := cms.Save(buffer)
	if err != nil {
		t.Fatalf("failed to save data %s", err)
		return
	}

	t.Logf("buffer len %d", buffer.Len())
	totalLen := 1 + 4 + 4 + 4*k + k*m*8
	if totalLen != uint32(buffer.Len()) {
		t.Fatalf("excected %d, got %d", totalLen, buffer.Len())
	}
}
