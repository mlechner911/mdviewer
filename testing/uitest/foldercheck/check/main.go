// Stand-in folder check for the UI test on win11: mode.txt next to
// .marksafe.yaml picks the answer (ok, findings, exit2).
package main

import (
	"fmt"
	"os"
	"strings"
)

func main() {
	b, _ := os.ReadFile("mode.txt")
	switch strings.TrimSpace(string(b)) {
	case "findings":
		fmt.Println(`{"ok":false,"findings":[{"file":"sop.md","line":1,"col":1,"severity":"error","code":"version-unchanged","message":"Text geändert, Version nicht erhöht"}]}`)
		os.Exit(1)
	case "exit2":
		fmt.Println(`{"ok":false,"error":"Prüfprogramm passt nicht zu den Prüfregeln im Checkout (Testfall) — Prüfprogramm neu bauen"}`)
		os.Exit(2)
	default:
		fmt.Println(`{"ok":true,"findings":[]}`)
	}
}
