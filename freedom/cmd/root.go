package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	cfgFile string

	// rootCmd represents the base command when called without any subcommands
	rootCmd = &cobra.Command{
		Use:   "The tool used to generate the code.",
		Short: "The tool used to generate the code.",
	}
)

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {

}

//AddCommand add sub command to root
func AddCommand(cmd *cobra.Command) {
	rootCmd.AddCommand(cmd)
}

//Commands returns all sub command
func Commands() []*cobra.Command {
	return rootCmd.Commands()
}
