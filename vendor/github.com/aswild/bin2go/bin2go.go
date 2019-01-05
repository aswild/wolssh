// bin2go: program to create go byte arrays from files
//
// Copyright 2018 Allen Wild <allenwild93@gmail.com>
// SPDX-License-Identifier: MIT

package bin2go

import (
    "fmt"
    "io"
    "io/ioutil"
)

type Generator struct{
    vars    map[string]string   // map of variablename:filename
    pkg     string              // package name
}

// Create a new bin2go generator with the given package name and input files.
// Variable names for input files are autogenerated with FilenameToVarname
// For more granular control, use AddFileVar
func New(pkg string, files ...string) (*Generator, error) {
    g := Generator{pkg: pkg}
    g.vars = make(map[string]string)
    for _, f := range files {
        v, err := FilenameToVarname(f)
        if err != nil {
            return nil, err
        }
        if err := g.addFileVar(f, v); err != nil {
            return nil, err
        }
    }
    return &g, nil
}

// Add a file with an autogenerated variable name.
// The variable name must not already be present.
func (g *Generator) AddFile(f string) error {
    v, err := FilenameToVarname(f)
    if err != nil {
        return err
    }
    return g.addFileVar(f, v)
}

// Add a file with the given variable name.
// The variable name must be valid and not already present.
func (g *Generator) AddFileVar(f, v string) error {
    if !CheckVarname(v) {
        return fmt.Errorf("Invalid variable name %q", v)
    }
    return g.addFileVar(f, v)
}

func (g *Generator) addFileVar(f, v string) error {
    // make sure a variable v doesn't already exist
    if _, ok := g.vars[v]; ok {
        return fmt.Errorf("Duplicate variable %q", v)
    }
    g.vars[v] = f
    return nil
}

// Write the full output to the given Writer, this includes the header and all
// files' data.
func (g *Generator) Output(w io.Writer) error {
    if _, err := fmt.Fprintf(w, "// generated by bin2go\n\npackage %s\n", g.pkg); err != nil {
        return err
    }
    for v, f := range g.vars {
        if err := outputFileVar(w, f, v); err != nil {
            return err
        }
    }
    return nil
}

// Emit one file's data array
func outputFileVar(w io.Writer, f, v string) error {
    if _, err := fmt.Fprintf(w, "\nvar %s = [...]byte{", v); err != nil {
        return err
    }

    data, err := ioutil.ReadFile(f)
    if err != nil {
        return fmt.Errorf("Failed to read file: %v", err)
    }

    for i, b := range data {
        if (i % 16) == 0 {
            if _, err := io.WriteString(w, "\n    "); err != nil {
                return err
            }
        } else {
            if _, err := io.WriteString(w, " "); err != nil {
                return err
            }
        }
        if _, err := fmt.Fprintf(w, "0x%02x,", b); err != nil {
            return err
        }
    }
    if _, err := io.WriteString(w, "\n}\n"); err != nil {
        return err
    }
    return nil
}