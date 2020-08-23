package main

import (
	"errors"

	"github.com/darxkies/k8s-tew/pkg/ceph"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var cephBinariesPath string
var cephConfigPath string
var cephDataPath string
var cephID string

func getCeph() *ceph.Ceph {
	return ceph.NewCeph(_config, cephBinariesPath, cephConfigPath, cephDataPath)
}

var cephCmd = &cobra.Command{
	Use:   "ceph",
	Short: "Setup and run Ceph cluster",
	Long:  "Setup and run Ceph cluster",
	RunE: func(cmd *cobra.Command, args []string) error {
		return errors.New("Missing sub-command")
	},
}

var cephInitializeCmd = &cobra.Command{
	Use:   "initialize",
	Short: "Initialize the Ceph cluster",
	Long:  "Initialize the Ceph cluster",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load config and check the rights
		if error := bootstrap(false); error != nil {
			return error
		}

		ceph := getCeph()

		if _, _error := ceph.Setup(); _error != nil {
			return _error
		}

		log.Info("Initialized")

		return nil
	},
}

var cephMgrCmd = &cobra.Command{
	Use:   "mgr",
	Short: "Run mgr",
	Long:  "Run mgr",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load config and check the rights
		if error := bootstrap(false); error != nil {
			return error
		}

		ceph := getCeph()

		if _error := ceph.RunMgr(cephID); _error != nil {
			return _error
		}

		return nil
	},
}

var cephMonCmd = &cobra.Command{
	Use:   "mon",
	Short: "Run mon",
	Long:  "Run mon",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load config and check the rights
		if error := bootstrap(false); error != nil {
			return error
		}

		ceph := getCeph()

		if _error := ceph.RunMon(cephID); _error != nil {
			return _error
		}

		return nil
	},
}

var cephMdsCmd = &cobra.Command{
	Use:   "mds",
	Short: "Run mds",
	Long:  "Run mds",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load config and check the rights
		if error := bootstrap(false); error != nil {
			return error
		}

		ceph := getCeph()

		if _error := ceph.RunMds(cephID); _error != nil {
			return _error
		}

		return nil
	},
}

var cephOsdCmd = &cobra.Command{
	Use:   "osd",
	Short: "Run osd",
	Long:  "Run osd",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load config and check the rights
		if error := bootstrap(false); error != nil {
			return error
		}

		ceph := getCeph()

		if _error := ceph.RunOsd(cephID); _error != nil {
			return _error
		}

		return nil
	},
}

var cephRgwCmd = &cobra.Command{
	Use:   "rgw",
	Short: "Run rgw",
	Long:  "Run rgw",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load config and check the rights
		if error := bootstrap(false); error != nil {
			return error
		}

		ceph := getCeph()

		if _error := ceph.RunRgw(cephID); _error != nil {
			return _error
		}

		return nil
	},
}

func init() {
	cephMgrCmd.Flags().StringVar(&cephID, "id", "", "id")
	cephMonCmd.Flags().StringVar(&cephID, "id", "", "id")
	cephMdsCmd.Flags().StringVar(&cephID, "id", "", "id")
	cephOsdCmd.Flags().StringVar(&cephID, "id", "0", "id")
	cephRgwCmd.Flags().StringVar(&cephID, "id", "0", "id")

	cephCmd.AddCommand(cephInitializeCmd)
	cephCmd.AddCommand(cephMgrCmd)
	cephCmd.AddCommand(cephMonCmd)
	cephCmd.AddCommand(cephMdsCmd)
	cephCmd.AddCommand(cephOsdCmd)
	cephCmd.AddCommand(cephRgwCmd)

	cephCmd.Flags().StringVar(&cephBinariesPath, "ceph-binaries-path", ceph.CephBinariesPath, "Location of Ceph binaries")
	cephCmd.Flags().StringVar(&cephConfigPath, "ceph-config-path", ceph.CephConfigPath, "Location of Ceph config")
	cephCmd.Flags().StringVar(&cephDataPath, "ceph-data-path", ceph.CephDataPath, "Location of Ceph data")

	RootCmd.AddCommand(cephCmd)
}
