package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/ka2n/simple-xbrl/xbrl"
)

func main() {
	if len(os.Args) < 2 {
		return
	}
	f, err := os.Open(os.Args[1])
	if err != nil {
		panic(err)
	}
	defer f.Close()

	// Unmarshal
	var x xbrl.XBRL
	if err := xbrl.UnmarshalXBRL(&x, f); err != nil {
		fmt.Fprintf(os.Stderr, "unmarshal error: %v", err.Error())
	}

	// Example: Marshal to JSON
	mapping := make(map[string][]xbrl.Fact)
	for _, ctx := range x.Contexts {
		mapping[ctx.ID] = make([]xbrl.Fact, 0)
	}
	for _, fact := range x.Facts {
		if facts, ok := mapping[fact.ContextRef]; ok {
			mapping[fact.ContextRef] = append(facts, fact)
		} else {
			mapping[fact.ContextRef] = []xbrl.Fact{fact}
		}
	}
	jb, err := json.MarshalIndent(mapping, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Fprintf(os.Stdout, string(jb))
}
