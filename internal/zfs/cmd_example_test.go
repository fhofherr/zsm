package zfs_test

import (
	"bytes"
	"fmt"
	"log"
	"strings"

	"github.com/fhofherr/zsm/internal/zfs"
)

func ExampleNewCmdFunc() {
	var out bytes.Buffer

	cmdFunc := zfs.NewCmdFunc("tr", "a-z")
	cmd := cmdFunc("A-Z")
	cmd.Stdin = strings.NewReader("some input")
	cmd.Stdout = &out

	// Execution equivalent to
	//
	//     echo "some input" | tr "a-z" "A-Z"
	if err := cmd.Run(); err != nil {
		log.Fatalf("run cmd: %v", err)
	}
	fmt.Printf("output: %q", out.String())
}
