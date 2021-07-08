package main

import (
	"flag"
	"fmt"
	"github.com/JiaoDean/alibaba-cloud-csi-driver-operator/pkg/operator"
	"github.com/JiaoDean/alibaba-cloud-csi-driver-operator/pkg/version"
	"github.com/openshift/library-go/pkg/controller/controllercmd"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"io"
	"os"
	"strings"
	"time"
)

const (
	logfilePrefix = "/var/log/alicloud/"
	mbsize        = 1024 * 1024
)

var (
	logLevel = flag.String("log-level", "Info", "Set Log Level")
)

func setLogAttribute(logName string) {
	logType := os.Getenv("LOG_TYPE")
	logType = strings.ToLower(logType)
	if logType != "stdout" && logType != "host" {
		logType = "both"
	}
	if logType == "stdout" {
		return
	}

	os.MkdirAll(logfilePrefix, os.FileMode(0755))
	logFile := logfilePrefix + logName + ".log"
	f, err := os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		os.Exit(1)
	}

	// rotate the log file if too large
	if fi, err := f.Stat(); err == nil && fi.Size() > 2*mbsize {
		f.Close()
		timeStr := time.Now().Format("-2006-01-02-15:04:05")
		timedLogfile := logfilePrefix + logName + timeStr + ".log"
		os.Rename(logFile, timedLogfile)
		f, err = os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			os.Exit(1)
		}
	}
	if logType == "both" {
		mw := io.MultiWriter(os.Stdout, f)
		log.SetOutput(mw)
	} else {
		log.SetOutput(f)
	}

	logLevelLow := strings.ToLower(*logLevel)
	if logLevelLow == "debug" {
		log.SetLevel(log.DebugLevel)
	} else if logLevelLow == "warning" {
		log.SetLevel(log.WarnLevel)
	}
	log.Infof("Set Log level to %s...", logLevelLow)
}

func NewOperatorCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "alibabacloud-csi-driver-operator",
		Short: "OpenShift Alibaba Cloud CSI Driver Operator",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
			os.Exit(1)
		},
	}

	ctrlCmd := controllercmd.NewControllerCommandConfig(
		"alibabacloud-csi-driver-operator",
		version.Get(),
		operator.RunOperator,
	).NewCommand()
	ctrlCmd.Use = "start"
	ctrlCmd.Short = "Start the AlibabaCloud CSI Driver Operator"

	cmd.AddCommand(ctrlCmd)

	return cmd
}

func main() {
	flag.Parse()
	setLogAttribute("alibabacloud.csi.driver.com")
	command := NewOperatorCommand()
	if err := command.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
