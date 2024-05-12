/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>

*/
package cmd

import (
	"os"

	"github.com/dsx1123/gnmi_go/pkg/app"
	"github.com/spf13/cobra"
)

var a = app.New()

func newRootCmd() *cobra.Command {
	a.RootCmd = &cobra.Command{
		Use:   "gnmi_go",
		Short: "A demo application of gNMI",
		Long: `A demo application of gNMI:
	Demonstrate gNMI  CAPABILITES/GET/SET/SUBSCRIBE.`,
		SilenceUsage: false,
	}
	a.InitFlags()
	a.RootCmd.AddCommand(capCmd(a))
	a.RootCmd.AddCommand(getCmd(a))
	a.RootCmd.AddCommand(setMergeCmd(a))
	a.RootCmd.AddCommand(setReplaceCmd(a))
	a.RootCmd.AddCommand(subCmd(a))
	a.RootCmd.AddCommand(edaCmd(a))
	return a.RootCmd
}

func Execute() {
	err := newRootCmd().Execute()
	if err != nil {
		os.Exit(1)
	}
}
