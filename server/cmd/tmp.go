package main

import (
	"fmt"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/unicode/norm"
)

var folder = cases.Fold()

func Display(s string) string {
	return norm.NFC.String(strings.TrimSpace(s))
}

func Key(s string) string {
	s = strings.TrimSpace(s)
	s = norm.NFKC.String(s)
	s = folder.String(s)
	return s
}

func main() {
	// Two visually similar usernames:
	// "êric" (e + combining circumflex)
	// "êric"  (precomposed character)

	u1 := "e\u0302riC"
	u2 := "êriС"

	fmt.Println("RAW INPUTS:")
	fmt.Printf("u1: %q\n", u1)
	fmt.Printf("u2: %q\n\n", u2)

	fmt.Println("DISPLAY (NFC):")
	d1 := Display(u1)
	d2 := Display(u2)
	fmt.Printf("d1: %q\n", d1)
	fmt.Printf("d2: %q\n\n", d2)

	fmt.Println("BYTES OF DISPLAY:")
	fmt.Printf("d1 bytes: %v\n", []byte(d1))
	fmt.Printf("d2 bytes: %v\n\n", []byte(d2))

	fmt.Println("IDENTITY KEY (NFKC + CASEFOLD):")
	k1 := Key(u1)
	k2 := Key(u2)
	fmt.Printf("k1: %q\n", k1)
	fmt.Printf("k2: %q\n\n", k2)

	fmt.Println("BYTES OF KEY:")
	fmt.Printf("k1 bytes: %v\n", []byte(k1))
	fmt.Printf("k2 bytes: %v\n\n", []byte(k2))

	fmt.Println("ARE KEYS EQUAL?")
	fmt.Println(k1 == k2)
	fmt.Println(d1 == d2)
}
