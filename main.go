/*
Copyright © 2025 Conner Ohnesorge <connerohnesorge@outlook.com>
*/
package main

import (
	"os"

	"github.com/connerohnesorge/catls/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
