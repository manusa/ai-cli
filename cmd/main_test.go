package main

import "os"

func Example_version() {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"ai-cli", "version"}
	main()
	// Output: 0.0.0
}

func Example_version_flag() {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"ai-cli", "--version"}
	main()
	// Output: 0.0.0
}
