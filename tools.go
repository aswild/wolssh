// +build linux
// +build windows

// I have to import bin2go/cmd to get the source for the tool (main package)
// but if go tries to compile this file it fails because bin2go/cmd isn't
// actually importable in code.
// As a workaround, define an impossible set of build conditions (linux and windows)
// so that this file never actually gets compiled.

package main

import (
    _ "github.com/aswild/bin2go/cmd"
)
