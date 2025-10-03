// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	"os"

	"golang.org/x/sync/errgroup"
)

func main() {
	eg := new(errgroup.Group)

	eg.Go(newGenerator().generateLogs)

	if err := eg.Wait(); err != nil {
		fmt.Println("Error:", err)

		os.Exit(1)
	}

	fmt.Println("Test data generated successfully")

	os.Exit(0)
}
