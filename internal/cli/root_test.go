package cli

import (
	"testing"

	"github.com/spf13/cobra"
)

func findExactChild(cmdName string, parent interface{ Commands() []*cobra.Command }) *cobra.Command {
	for _, child := range parent.Commands() {
		if child.Name() == cmdName {
			return child
		}
	}
	return nil
}

func TestRootCommandRemovesGlobalNSFlag(t *testing.T) {
	root := NewRootCommand()
	if root.PersistentFlags().Lookup("ns") != nil {
		t.Fatalf("expected root command to not expose a global --ns flag")
	}
}

func TestAuthCommandUsesWhoAmIAndScopedLoginNS(t *testing.T) {
	root := NewRootCommand()
	authCmd := findExactChild("auth", root)
	if authCmd == nil {
		t.Fatalf("expected auth command")
	}

	if findExactChild("whoami", authCmd) == nil {
		t.Fatalf("expected auth whoami command")
	}
	if findExactChild("status", authCmd) != nil {
		t.Fatalf("expected auth status command to be absent")
	}

	loginCmd := findExactChild("login", authCmd)
	if loginCmd == nil {
		t.Fatalf("expected auth login command")
	}
	if loginCmd.Flags().Lookup("ns") == nil {
		t.Fatalf("expected auth login to expose --ns")
	}
}

func TestNSInviteCommandsAreNested(t *testing.T) {
	root := NewRootCommand()
	nsCmd := findExactChild("ns", root)
	if nsCmd == nil {
		t.Fatalf("expected ns command")
	}
	inviteCmd := findExactChild("invite", nsCmd)
	if inviteCmd == nil {
		t.Fatalf("expected ns invite command")
	}

	if findExactChild("list", inviteCmd) == nil {
		t.Fatalf("expected ns invite list")
	}
	if findExactChild("create", inviteCmd) == nil {
		t.Fatalf("expected ns invite create")
	}
	if findExactChild("accept", inviteCmd) == nil {
		t.Fatalf("expected ns invite accept")
	}
	if findExactChild("invite-list", nsCmd) != nil {
		t.Fatalf("expected ns invite-list to be absent")
	}
	if findExactChild("accept-invite", nsCmd) != nil {
		t.Fatalf("expected ns accept-invite to be absent")
	}
}

func TestDocumentCreateCommandsExposeSourceSpecificSubcommands(t *testing.T) {
	root := NewRootCommand()
	documentCmd := findExactChild("document", root)
	if documentCmd == nil {
		t.Fatalf("expected document command")
	}
	createCmd := findExactChild("create", documentCmd)
	if createCmd == nil {
		t.Fatalf("expected document create command")
	}

	for _, name := range []string{"text", "file", "url", "x", "twitter", "zhihu", "wechat", "youtube", "xiaohongshu"} {
		if findExactChild(name, createCmd) == nil {
			t.Fatalf("expected document create %s subcommand", name)
		}
	}
}
