// bin2go: program to create go byte arrays from files
//
// Copyright 2018 Allen Wild <allenwild93@gmail.com>
// SPDX-License-Identifier: MIT

package main

import (
    "bytes"
    "flag"
    "fmt"
    "io"
    "os"
    "strings"

    "github.com/aswild/bin2go"
)

var opts struct {
    help    bool
    outfile string
    pkg     string
}

const usageText =
`Usage: %s [-h] [-p NAME] [-o FILE] FILENAME[:VARNAME] [FILENAME2[:VARNAME2]...]
  Options:
    -h         Show this help and exit
    -p PKG     Package name in the generated file
    -o FILE    Output file (default stdout)

  Input Files:
    Each input file is turned into a Go byte array (not slice) containing the
    full contents of the file. To specify the generated variable's name, append
    it to the filename, separated by a colon (there is no support for filenames
    with colons in them).

  Variable names must match the regex '^[A-Za-z_][A-Za-z0-9_]*$' but can't be
  a single underscore.
  If a variable name isn't specified, it will be created from the filename by:
    1. replace all invalid characters with underscores
    2. squash all adjacent underscores
    3. remove trailing underscores
    4. if the first character is a number, prepend an underscore
  If a valid variable name can't be created from the filename, that's an error.
`

func usage() {
    fmt.Fprintf(os.Stderr, usageText, os.Args[0])
}

func logErr(format string, args ...interface{}) {
    fmt.Fprintf(os.Stderr, format, args...)
    os.Stderr.WriteString("\n")
}

func realMain() (ret int) {
    flag.Usage = usage
    flag.BoolVar(&opts.help, "h", false, "show help and exit")
    flag.StringVar(&opts.pkg, "p", "main", "package name (default main)")
    flag.StringVar(&opts.outfile, "o", "", "Output file (default stdout)")
    flag.Parse()

    if opts.help {
        usage()
        return 0
    }

    if len(flag.Args()) < 1 {
        logErr("No filename specified")
        usage()
        return 2
    }

    gen, err := bin2go.New(opts.pkg)
    if err != nil {
        logErr("Failed to initialize generator: %v", err)
        return 1
    }

    for _, f := range flag.Args() {
        sp := strings.SplitN(f, ":", 2)
        if len(sp) == 1 {
            if err := gen.AddFile(f); err != nil {
                logErr("Failed to add file %q: %v", f, err)
                return 1
            }
        } else {
            if err := gen.AddFileVar(sp[0], sp[1]); err != nil {
                logErr("Failed to add file/var %q: %v", f, err)
                return 1
            }
        }
    }

    // input flags have been handled and validated, time to start the real work
    var out io.Writer
    useStdout := opts.outfile == ""
    ret = 1

    if useStdout {
        // if stdout is the destination, write to an internal buffer.
        // The output will be printed at the end only if there were
        // no errors.
        out = new(bytes.Buffer)
    } else {
        // open (create/truncate) output file
        out, err = os.Create(opts.outfile)
        if err != nil {
            logErr("Failed to open output file: %v", err)
            return
        }
    }

    // deferred cleanup handler
    defer func() {
        if !useStdout {
            // we wrote to a file, close it
            out.(*os.File).Close()
            if ret != 0 {
                // if we failed, delete the partially-written output file
                os.Remove(opts.outfile)
            }
        } else if ret == 0 {
            // if we succeeded and are printing to stdout, actually
            // print the output. (if we failed it will be discared)
            os.Stdout.Write(out.(*bytes.Buffer).Bytes())
        }
    }()

    // generate the output. It's not until here that the input files
    // will actually be opened/read
    if err := gen.Output(out); err != nil {
        logErr("Failed to generate data: %v", err)
        return
    }

    ret = 0 // success!
    return
}

func main() {
    // wrap main since os.Exit doesn't call deferred functions
    os.Exit(realMain())
}
