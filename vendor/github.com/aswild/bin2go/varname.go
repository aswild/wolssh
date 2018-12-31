// bin2go: program to create go byte arrays from files
//
// Copyright 2018 Allen Wild <allenwild93@gmail.com>
// SPDX-License-Identifier: MIT

package bin2go

import (
    "fmt"
    "regexp"
)

var (
    validVarnameRe   = regexp.MustCompile("^[A-Za-z_][A-Za-z0-9_]*$")
    validStartCharRe = regexp.MustCompile("^[A-Za-z_]")
    invalidVarCharRe = regexp.MustCompile("[^A-Za-z0-9_]")
)

func CheckVarname(f string) bool {
    return f != "_" && validVarnameRe.MatchString(f)
}

// filename to variable name conversion:
//   1. replace all invalid characters with underscores
//   2. squash adjacent underscores
//   3. remove trailing underscores
//   4. if first character is a number, prepend an underscore
// Returns an error if the filename couldn't be converted (i.e. if it
// contains only underscores or special characters)
func FilenameToVarname(filename string) (string, error) {
    if CheckVarname(filename) {
        return filename, nil
    }

    f := invalidVarCharRe.ReplaceAllString(filename, "_")
    f = regexp.MustCompile("_+").ReplaceAllString(f, "_")
    f = regexp.MustCompile("_+$").ReplaceAllString(f, "")
    if !validStartCharRe.MatchString(f) {
        f = "_" + f
    }

    if !CheckVarname(f) {
        return "", fmt.Errorf("couldn't convert filename %q to variable name", filename)
    }
    return f, nil
}
