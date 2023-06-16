package main

import (
	"fmt"
	"github.com/setherplatform/sether-node/version"
	"os"

	"github.com/setherplatform/sether-node/cmd/sether/launcher"
)

func main() {
	fmt.Printf("Sether runtime v%s\n", version.AsString())
	if err := launcher.Launch(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
