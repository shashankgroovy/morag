package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// fetchCmd represents the fetch command
var fetchCmd = &cobra.Command{
	Use:   "fetch",
	Short: "Fetches track information for an artist.",
	Long: `Fetch helps you download the entire catalog/library of an artist (of
your choice) from Spotify and save that in an output.csv file. A valid
artistID or multiple artistIDs can be passed separated by space as
arguments to this command.

USAGE:
$ morag fetch [artistID]

EXAMPLE:
$ morag fetch 0OdUWJ0sBjDrqHygGUXeCF
`,
	Run: fetch,
}

func init() {
	rootCmd.AddCommand(fetchCmd)

	// Add a local flag which will only run when this command
	// is called directly.
	fetchCmd.Flags().StringP("output_csv", "o", "output.csv", "Provide an output file name of your choice")
}

func fetch(cmd *cobra.Command, args []string) {
	if len(args) < 1 {
		fmt.Printf("ERROR: Please provide a Spotify artistID.\nSee the help text below.\n\n")
		cmd.Help()

	}
	fmt.Println("fetch called. ", args)
}
