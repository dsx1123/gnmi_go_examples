package cmd

import (
	"github.com/dsx1123/gnmi_go/pkg/app"
	"github.com/spf13/cobra"
)

func getCmd(a *app.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "get",
		Short:        "get from target",
		PreRunE:      a.PreRunE,
		RunE:         a.RunE,
		SilenceUsage: true,
	}
	return cmd

}

func capCmd(a *app.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "cap",
		Short:        "get gnmi capabilites from target",
		PreRunE:      a.PreRunE,
		RunE:         a.RunE,
		SilenceUsage: true,
	}
	return cmd

}

func setMergeCmd(a *app.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "merge",
		Short:        "merge the configurations on target",
		PreRunE:      a.PreRunE,
		RunE:         a.RunE,
		SilenceUsage: true,
	}
	return cmd

}

func setReplaceCmd(a *app.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "replace",
		Short:        "replace the configurations on target",
		PreRunE:      a.PreRunE,
		RunE:         a.RunE,
		SilenceUsage: true,
	}
	return cmd

}

func subCmd(a *app.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "subscribe",
		Short:        "run gnmi subscribe on target",
		PreRunE:      a.PreRunE,
		RunE:         a.RunE,
		SilenceUsage: true,
	}
	return cmd
}

func edaCmd(a *app.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:          "eda",
		Short:        "run EDA demo on target",
		PreRunE:      a.PreRunE,
		RunE:         a.RunE,
		SilenceUsage: true,
	}
	return cmd
}
