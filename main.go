package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	var rootCmd = newRootCmd()

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}

func newRootCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "hello",
		Short: "This is hello command",
		Long:  "This is hello command long long description",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("world")
			return nil
		},
	}
}
