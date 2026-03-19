package app

import (
	"context"
	"fmt"
	"io"
)

func Run(ctx context.Context, args []string, in io.Reader, out io.Writer, errOut io.Writer) int {
	_ = ctx
	_ = in

	if len(args) > 0 && args[0] == "--help" {
		fmt.Fprintln(out, "Usage: wt [--fzf] [index]")
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Print the selected git worktree path.")
		return 0
	}

	fmt.Fprintln(errOut, "not implemented")
	return 1
}
