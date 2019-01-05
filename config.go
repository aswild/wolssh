/*******************************************************************************
* config.go: WolSSH config handling
*
* Copyright 2018 Allen Wild <allenwild93@gmail.com>
* SPDX-License-Identifier: MIT
*******************************************************************************/

package main

import (
    "fmt"
    "strings"

    "gopkg.in/ini.v1"
)

type LogConfig struct {
    Level       int
    File        string
    Stderr      bool
    Syslog      bool
    Facility    int
    Tag         string
}

type UserConfig struct {
    Name    string
    Keys    []string `ini:"pubkey,omitempty,allowshadow"`
}

type Config struct {
    Listen      string
    SshDir      string
    Log         LogConfig
    BcastStrs   []string            `ini:"broadcast,omitempty,allowshadow"`
    bcastAddrs  []BroadcastAddr     `ini:"-"`
    Hosts       map[string]string   `ini:"-"`
    Users       []UserConfig        `ini:"-"`
}

func DefaultConfig() (*Config) {
    return &Config{
        Listen:     ":2222",
        SshDir:     "ssh",
        Log: LogConfig{
            Level:      int(LOG_LEVEL_INFO),
            File:       "",
            Stderr:     false,
            Syslog:     false,
            Facility:   18,
            Tag:        "wolssh",
        },
        BcastStrs:  []string{"255.255.255.255"},
    }
}

func LoadConfig(filename string) (*Config, error) {
    // workaround to set options/mapper since we can't do both using
    // ShadowLoad/StrictMapToWithMapper
    iconf, _ := ini.LoadSources(ini.LoadOptions{AllowShadows:true}, []byte(""))
    iconf.NameMapper = ini.TitleUnderscore

    // read the file
    if err := iconf.Append(filename); err != nil {
        return nil, err
    }

    // map everything that go-ini can
    if err := iconf.Section("wolssh").StrictMapTo(conf); err != nil {
        return nil, err
    }

    // set up users
    for _, s := range iconf.Section("user").ChildSections() {
        u := UserConfig{Name: strings.TrimPrefix(s.Name(), "user.")}
        if err := s.StrictMapTo(&u); err != nil {
            return nil, fmt.Errorf("failed to map user %s: %v\n", u.Name, err)
        }
        conf.Users = append(conf.Users, u)
    }

    // set up hosts mapping
    conf.Hosts = iconf.Section("hosts").KeysHash()

    return conf, nil
}
