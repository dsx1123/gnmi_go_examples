package cmd

import (
	"github.com/dsx1123/gnmi_go/pkg/app"
	"github.com/spf13/cobra"
)

func capCmd(a *app.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "cap",
		Short:        "Get the gNMI capabilites from the target",
		PreRunE:      a.PreRunE,
		RunE:         a.RunE,
		SilenceUsage: true,
	}
	return cmd

}

func getCmd(a *app.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "get",
		Short:        "Get the configuration or operational state from the target",
		PreRunE:      a.PreRunE,
		RunE:         a.RunE,
		SilenceUsage: true,
	}
	return cmd

}

func setMergeCmd(a *app.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "merge",
		Short:        "Merge the candidate configuration wtih the configuration on the target",
		PreRunE:      a.PreRunE,
		RunE:         a.RunE,
		SilenceUsage: true,
	}
	return cmd

}

func setReplaceCmd(a *app.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "replace",
		Short:        "Replace the configurations on the target",
		PreRunE:      a.PreRunE,
		RunE:         a.RunE,
		SilenceUsage: true,
	}
	return cmd

}

func deleteCmd(a *app.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "delete",
		Short:        "Delete a conatiner or leaf on the target",
		PreRunE:      a.PreRunE,
		RunE:         a.RunE,
		SilenceUsage: true,
	}
	return cmd

}

func subCmd(a *app.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "subscribe",
		Short:        "run gnmi subscribe on the target",
		PreRunE:      a.PreRunE,
		RunE:         a.RunE,
		SilenceUsage: true,
	}
	return cmd
}

func edaCmd(a *app.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "eda",
		Short:        "Run EDA demo on the target",
		PreRunE:      a.PreRunE,
		RunE:         a.RunE,
		SilenceUsage: true,
	}
	return cmd
}
