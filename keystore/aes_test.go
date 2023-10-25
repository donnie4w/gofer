package keystore

import (
	"fmt"
	"testing"
)

func TestAES(t *testing.T) {
	a := NewAes("ajfijeifanb", "iejfefvvvvvvvvefa")
	s := "1234567aaaafffaaaafffaaaaaa1aa"
	bs, _ := a.EncrypterCbc([]byte(s))
	b, _ := a.DecrypterCbc(bs)

	fmt.Println(b)
}

func BenchmarkAES(b *testing.B) {
	a := NewAes("ajfijeifanb", "iejfefvvvvvvvvefa")
	for i := 0; i < b.N; i++ {
		s := "1234567"
		s = s +  fmt.Sprint(i)
		bs, _ := a.EncrypterCbc([]byte(s))
		b, _ := a.DecrypterCbc(bs)
		if string(b) != s {
			fmt.Println(len(b), " >>", len([]byte(s)))
			panic(">>>" + s + " " + string(b))
		}
	}
}
