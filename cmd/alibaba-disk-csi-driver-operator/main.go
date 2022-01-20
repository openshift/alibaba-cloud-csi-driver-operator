package main

import (
	"os"

	"github.com/spf13/cobra"

	"k8s.io/component-base/cli"

	"github.com/openshift/alibaba-disk-csi-driver-operator/pkg/operator"
	"github.com/openshift/alibaba-disk-csi-driver-operator/pkg/version"
	"github.com/openshift/library-go/pkg/controller/controllercmd"
)

func main() {
	command := NewOperatorCommand()
	code := cli.Run(command)
	os.Exit(code)
}

func NewOperatorCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "alibaba-disk-csi-driver-operator",
		Short: "OpenShift Alibaba Disk CSI Driver Operator",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
			os.Exit(1)
		},
	}

	ctrlCmd := controllercmd.NewControllerCommandConfig(
		"alibaba-disk-csi-driver-operator",
		version.Get(),
		operator.RunOperator,
	).NewCommand()
	ctrlCmd.Use = "start"
	ctrlCmd.Short = "Start the Alibaba Disk CSI Driver Operator"

	cmd.AddCommand(ctrlCmd)

	return cmd
}
