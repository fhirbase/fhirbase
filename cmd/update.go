package cmd

import (
	"context"
	"fmt"
	"log"

	"github.com/creativeprojects/go-selfupdate"
	"github.com/spf13/cobra"
)

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Updates Fhirbase to most recent version",
	Long:  `Updates Fhirbase to most recent version.`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		UpdateCommand(ctx)
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// updateCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// updateCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// UpdateCommand updates the binary to the most recent version
func UpdateCommand(ctx context.Context) {
	latest, found, err := selfupdate.DetectLatest(ctx, selfupdate.ParseSlug("fhirbase/fhirbase"))

	if err != nil {
		log.Fatalf("Error occurred while detecting version: %v", err)
	}

	if !found || latest.LessOrEqual(Version) {
		log.Println("Current version is the latest")
		return
	}

	fmt.Printf("Do you want to update to version %s? (y/n): ", latest.Version())
	var response string
	fmt.Scanln(&response)

	if response != "y" {
		return
	}

	exe, err := selfupdate.ExecutablePath()
	if err != nil {
		log.Fatalf("Error occurred while getting executable path: %v", err)
	}
	if err := selfupdate.UpdateTo(context.Background(), latest.AssetURL, latest.AssetName, exe); err != nil {
		log.Fatalf("Error occurred while updating binary: %v", err)
	}
	log.Printf("Successfully updated to version %s", latest.Version())
	return

}
