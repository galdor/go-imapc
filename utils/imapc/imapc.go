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

	"github.com/galdor/go-cmdline"
	"github.com/galdor/go-imapc"
)

func main() {
	// Command line
	cmdline := cmdline.New()

	cmdline.AddOption("l", "login", "login", "set the login")
	cmdline.AddOption("p", "password", "password", "set the password")

	cmdline.AddCommand("connect", "connect to a server")
	cmdline.AddCommand("list", "list mailboxes")
	cmdline.AddCommand("examine", "examine a mailbox")

	cmdline.Parse(os.Args)

	login := cmdline.OptionValue("login")
	password := cmdline.OptionValue("password")

	cmd := cmdline.CommandName()
	cmdArgs := cmdline.CommandArgumentsValues()

	// Command
	var cmdFn func(*imapc.Client, []string)

	switch cmd {
	case "connect":
		cmdFn = CmdConnect
	case "list":
		cmdFn = CmdList
	case "examine":
		cmdFn = CmdExamine
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

	cmdFn(client, append([]string{cmd}, cmdArgs...))
}

func CmdConnect(client *imapc.Client, args []string) {
}

func CmdList(client *imapc.Client, args []string) {
	resps, err := client.SendCommandList("", "*")
	if err != nil {
		Die("%v", err)
	}

	width := 0
	for _, resp := range resps {
		if len(resp.MailboxName) > width {
			width = len(resp.MailboxName)
		}
	}

	for _, resp := range resps {
		fmt.Printf("%-*s ", width, resp.MailboxName)

		for _, flag := range resp.MailboxFlags {
			fmt.Printf(" \\%s", flag)
		}

		fmt.Printf("\n")
	}
}

func CmdExamine(client *imapc.Client, args []string) {
	cmdline := cmdline.New()
	cmdline.AddArgument("mailbox", "the name of the mailbox")
	cmdline.Parse(args)

	mailboxName := cmdline.ArgumentValue("mailbox")

	err := client.SendCommandExamine(mailboxName)
	if err != nil {
		Die("%v", err)
	}

	// TODO
}

func Error(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(os.Stderr, "%s\n", msg)
}

func Die(format string, args ...interface{}) {
	Error(format, args...)
	os.Exit(1)
}
