package cmd

import "github.com/spf13/cobra"

import "fmt"

const (
	versionNum = "v1.3.4"
)

var (
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
