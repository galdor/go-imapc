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
	cmdline.AddOption("m", "mailbox", "name", "select a mailbox")
	cmdline.AddOption("p", "password", "password", "set the password")

	cmdline.AddCommand("connect", "connect to a server")
	cmdline.AddCommand("list", "list mailboxes")
	cmdline.AddCommand("lsub", "list subscribed mailboxes")
	cmdline.AddCommand("create", "create a mailbox")
	cmdline.AddCommand("delete", "delete a mailbox")
	cmdline.AddCommand("rename", "rename a mailbox")
	cmdline.AddCommand("subscribe", "subscribe to a mailbox")
	cmdline.AddCommand("unsubscribe", "unsubscribe from a mailbox")
	cmdline.AddCommand("examine", "examine a mailbox")

	cmdline.Parse(os.Args)

	login := cmdline.OptionValue("login")
	mailbox := cmdline.OptionValue("mailbox")
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
	case "lsub":
		cmdFn = CmdLSub
	case "create":
		cmdFn = CmdCreate
	case "delete":
		cmdFn = CmdDelete
	case "rename":
		cmdFn = CmdRename
	case "subscribe":
		cmdFn = CmdSubscribe
	case "unsubscribe":
		cmdFn = CmdUnsubscribe
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

	if mailbox != "" {
		if _, err := client.SendCommandSelect(mailbox); err != nil {
			Die("cannot select mailbox: %v", err)
		}
	}

	cmdFn(client, append([]string{cmd}, cmdArgs...))

	if mailbox != "" {
		if err := client.SendCommandClose(); err != nil {
			Die("cannot close mailbox: %v", err)
		}
	}

	if err := client.SendCommandLogout(); err != nil {
		Die("cannot logout: %v", err)
	}
}

func CmdConnect(client *imapc.Client, args []string) {
}

func CmdList(client *imapc.Client, args []string) {
	rs, err := client.SendCommandList("", "*")
	if err != nil {
		Die("%v", err)
	}

	width := 0
	for _, mbox := range rs.Mailboxes {
		if len(mbox.Name) > width {
			width = len(mbox.Name)
		}
	}

	for _, mbox := range rs.Mailboxes {
		fmt.Printf("%-*s ", width, mbox.Name)

		for _, flag := range mbox.Flags {
			fmt.Printf(" \\%s", flag)
		}

		fmt.Printf("\n")
	}
}

func CmdLSub(client *imapc.Client, args []string) {
	rs, err := client.SendCommandLSub("", "*")
	if err != nil {
		Die("%v", err)
	}

	width := 0
	for _, mbox := range rs.Mailboxes {
		if len(mbox.Name) > width {
			width = len(mbox.Name)
		}
	}

	for _, mbox := range rs.Mailboxes {
		fmt.Printf("%-*s ", width, mbox.Name)

		for _, flag := range mbox.Flags {
			fmt.Printf(" \\%s", flag)
		}

		fmt.Printf("\n")
	}
}

func CmdCreate(client *imapc.Client, args []string) {
	cmdline := cmdline.New()
	cmdline.AddArgument("mailbox", "the name of the mailbox")
	cmdline.Parse(args)

	mailboxName := cmdline.ArgumentValue("mailbox")

	if err := client.SendCommandCreate(mailboxName); err != nil {
		Die("%v", err)
	}
}

func CmdDelete(client *imapc.Client, args []string) {
	cmdline := cmdline.New()
	cmdline.AddArgument("mailbox", "the name of the mailbox")
	cmdline.Parse(args)

	mailboxName := cmdline.ArgumentValue("mailbox")

	if err := client.SendCommandDelete(mailboxName); err != nil {
		Die("%v", err)
	}
}

func CmdRename(client *imapc.Client, args []string) {
	cmdline := cmdline.New()
	cmdline.AddArgument("mailbox", "the name of the mailbox")
	cmdline.AddArgument("new-name", "the new name of the mailbox")
	cmdline.Parse(args)

	mailboxName := cmdline.ArgumentValue("mailbox")
	mailboxNewName := cmdline.ArgumentValue("new-name")

	err := client.SendCommandRename(mailboxName, mailboxNewName)
	if err != nil {
		Die("%v", err)
	}
}

func CmdSubscribe(client *imapc.Client, args []string) {
	cmdline := cmdline.New()
	cmdline.AddArgument("mailbox", "the name of the mailbox")
	cmdline.Parse(args)

	mailboxName := cmdline.ArgumentValue("mailbox")

	if err := client.SendCommandSubscribe(mailboxName); err != nil {
		Die("%v", err)
	}
}

func CmdUnsubscribe(client *imapc.Client, args []string) {
	cmdline := cmdline.New()
	cmdline.AddArgument("mailbox", "the name of the mailbox")
	cmdline.Parse(args)

	mailboxName := cmdline.ArgumentValue("mailbox")

	if err := client.SendCommandUnsubscribe(mailboxName); err != nil {
		Die("%v", err)
	}
}

func CmdExamine(client *imapc.Client, args []string) {
	cmdline := cmdline.New()
	cmdline.AddArgument("mailbox", "the name of the mailbox")
	cmdline.Parse(args)

	mailboxName := cmdline.ArgumentValue("mailbox")

	rs, err := client.SendCommandExamine(mailboxName)
	if err != nil {
		Die("%v", err)
	}

	fmt.Printf("Flags           ")
	for _, flag := range rs.Flags {
		fmt.Printf(" \\%s", flag)
	}
	fmt.Printf("\n")

	fmt.Printf("Permanent flags ")
	for _, flag := range rs.PermanentFlags {
		fmt.Printf(" \\%s", flag)
	}
	fmt.Printf("\n")

	fmt.Printf("Messages         %d\n", rs.Exists)
	fmt.Printf("Unseen messages  %d\n", rs.Unseen)
	fmt.Printf("Recent messages  %d\n", rs.Recent)
}

func Error(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(os.Stderr, "%s\n", msg)
}

func Die(format string, args ...interface{}) {
	Error(format, args...)
	os.Exit(1)
}
