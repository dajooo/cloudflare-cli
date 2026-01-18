package r2

import (
	"github.com/spf13/cobra"
)

var R2Cmd = &cobra.Command{
	Use:   "r2",
	Short: "Manage R2 buckets and objects",
}

var bucketCmd = &cobra.Command{
	Use:   "bucket",
	Short: "Manage R2 buckets",
}

func init() {
	R2Cmd.AddCommand(bucketCmd)
}
