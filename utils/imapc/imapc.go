//
// Copyright (c) 2016 Nicolas Martyanoff <khaelin@gmail.com>
//
// Permission to use, copy, modify, and distribute this software for any
// purpose with or without fee is hereby granted, provided that the above
// copyright notice and this permission notice appear in all copies.
//
// THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
// WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
// MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
// ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
// WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
// ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
// OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.

package main

import (
	"fmt"
	"os"

	"github.com/galdor/go-imapc"
	"github.com/pborman/getopt"
)

func main() {
	// Command line
	var printHelp bool
	var login, password string

	getopt.BoolVarLong(&printHelp, "help", 'h', "print help and exit")
	getopt.StringVarLong(&login, "login", 'l', "login")
	getopt.StringVarLong(&password, "password", 'p', "password")

	getopt.CommandLine.SetParameters("<command> <args...>")
	getopt.CommandLine.Parse(os.Args)

	if printHelp {
		getopt.CommandLine.PrintUsage(os.Stdout)
		os.Exit(0)
	}

	args := getopt.CommandLine.Args()

	if len(args) < 1 {
		Die("missing argument")
	}

	cmd := args[0]
	cmdArgs := args[1:]

	// Command
	var cmdFn func(*imapc.Client, []string)

	switch cmd {
	case "help":
		getopt.CommandLine.PrintUsage(os.Stdout)
		return
	case "connect":
		cmdFn = CmdConnect
	default:
		Die("unknown command")
	}

	// Client
	client := imapc.NewClient()
	client.Login = login
	client.Password = password

	if err := client.Connect(); err != nil {
		Die("%v", err)
	}

	cmdFn(client, cmdArgs)
}

func CmdConnect(client *imapc.Client, args []string) {
}

func Error(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(os.Stderr, "%s\n", msg)
}

func Die(format string, args ...interface{}) {
	Error(format, args...)
	os.Exit(1)
}
