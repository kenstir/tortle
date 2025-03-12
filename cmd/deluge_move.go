/*
Copyright © 2025 Kenneth H. Cox
*/
package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/autobrr/go-deluge"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	delugeCmd.AddCommand(delugeMoveCmd)

	delugeMoveCmd.Flags().BoolP("force", "f", false, "Force move")
	viper.BindPFlag("deluge.move.force", delugeMoveCmd.Flags().Lookup("force"))
}

var delugeMoveCmd = &cobra.Command{
	Use:     "move hash path",
	Aliases: []string{"mv", "m"},
	Short:   "Move torrent",
	Args:    cobra.ExactArgs(2),
	Run:     delugeMoveCmdRun,
}

func delugeMoveCmdRun(cmd *cobra.Command, args []string) {
	hash := args[0]
	path := args[1]

	// get the flags
	force := viper.GetBool("deluge.move.force")

	// OMG, msys does path conversion, turning "/a" into "c:/Program Files/Git/a".
	// Do not allow this.
	if os.Getenv("MSYSTEM") != "" {
		if os.Getenv("MSYS_NO_PATHCONV") != "1" {
			stdoutLogger.Printf("Warning: MSYSTEM=%s, msys path conversion is in effect\n", os.Getenv("MSYSTEM"))
			if !force {
				fmt.Fprintf(os.Stderr, "Error: msys path conversion in effect, rerun with MSYS_NO_PATHCONV=1 or --force\n")
				os.Exit(1)
			}
		}
	}

	// create a deluge client
	client := delugeCreateV2Client()

	// move
	err := delugeMove(context.Background(), client, hash, path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func delugeMove(ctx context.Context, client deluge.DelugeClient, hash string, path string) error {
	// connect
	err := client.Connect(ctx)
	if err != nil {
		return err
	}
	defer client.Close()

	// move
	stdoutLogger.Printf("%s: requesting move to \"%s\"\n", hash, path)
	err = client.MoveStorage(ctx, []string{hash}, path)
	if err != nil {
		return err
	}

	return nil
}
