//go:build test_gitcommand

package gitcommand

import (
	"context"
	"fmt"
	"os"
	"testing"
)

func Test_GitCommandDriver(t *testing.T) {
	driver := NewGitCommandDriver("test", os.Getenv("GITHUB_TOKEN"))

	t.Run(`Clone -> SwitchNewBranch -> CommitAll -> Push`, func(t *testing.T) {
		// Clone
		dir, err := driver.Clone(context.Background(), "ShotaKitazawa", "dotfiles")
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println(dir)

		// SwitchNewBranch
		err = driver.SwitchNewBranch(context.Background(), dir, "demo")
		if err != nil {
			t.Fatal(err)
		}

		// create new empty file
		fp, _ := os.Create("/tmp/dotfiles/.test")
		fp.Close()

		// CommitAll
		err = driver.CommitAll(context.Background(), dir, "for test")
		if err != nil {
			t.Fatal(err)
		}

		// Push
		err = driver.Push(context.Background(), dir)
		if err != nil {
			t.Fatal(err)
		}
	})
}
