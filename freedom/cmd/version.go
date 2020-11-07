package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

const (
	versionNum = "v1.8.5"
)

var (
	// VersionCmd .
	VersionCmd = &cobra.Command{
		Use:   "version",
		Short: "Output current version number",
		Long:  `Output current version number`,
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			fmt.Println("freedom " + versionNum)
			return
		},
	}
)

func init() {
	AddCommand(VersionCmd)
}
