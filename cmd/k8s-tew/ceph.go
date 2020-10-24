package main

import (
	"errors"

	"github.com/darxkies/k8s-tew/pkg/ceph"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const SSL_CERTIFICATE = "/etc/ssl/certs/ceph.pem"
const SSL_KEY = "/etc/ssl/certs/ceph-key.pem"

var cephBinariesPath string
var cephConfigPath string
var cephDataPath string
var cephID string
var cephPublicAddress string
var cephDashboardUsername string
var cephDashboardPassword string
var cephRadosgwUsername string
var cephRadosgwPassword string
var cephSSLCertificate string
var cephSSLKey string
var cephProxyPort string
var cephPlacementGroups uint
var cephExpectedNumberOfObjects uint

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

		if _error := ceph.RunMgr(cephID, cephPublicAddress, cephSSLCertificate, cephSSLKey, cephProxyPort); _error != nil {
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

		if _error := ceph.RunMon(cephID, cephPublicAddress); _error != nil {
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

		if _error := ceph.RunMds(cephID, cephPublicAddress); _error != nil {
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

		if _error := ceph.RunOsd(cephID, cephPublicAddress); _error != nil {
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

		if _error := ceph.RunRgw(cephID, cephPublicAddress, cephSSLCertificate, cephSSLKey, cephProxyPort); _error != nil {
			return _error
		}

		return nil
	},
}

var cephSetupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Setup Ceph cluster",
	Long:  "Setup Ceph cluster",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Load config and check the rights
		if error := bootstrap(false); error != nil {
			return error
		}

		ceph := getCeph()

		if _error := ceph.RunSetup(cephDashboardUsername, cephDashboardPassword, cephRadosgwUsername, cephRadosgwPassword, cephSSLCertificate, cephSSLKey, cephPlacementGroups, cephExpectedNumberOfObjects); _error != nil {
			return _error
		}

		return nil
	},
}

func init() {
	cephMgrCmd.Flags().StringVar(&cephID, "id", "", "id")
	cephMgrCmd.Flags().StringVar(&cephPublicAddress, "ip", "", "ip")
	cephMgrCmd.Flags().StringVar(&cephSSLCertificate, "ssl-certificate", SSL_CERTIFICATE, "SSL Certificate")
	cephMgrCmd.Flags().StringVar(&cephSSLKey, "ssl-key", SSL_KEY, "SSL Key")
	cephMgrCmd.Flags().StringVar(&cephProxyPort, "proxy-port", "28715", "Proxy Port")
	cephMonCmd.Flags().StringVar(&cephID, "id", "", "id")
	cephMonCmd.Flags().StringVar(&cephPublicAddress, "ip", "", "ip")
	cephMdsCmd.Flags().StringVar(&cephID, "id", "", "id")
	cephMdsCmd.Flags().StringVar(&cephPublicAddress, "ip", "", "ip")
	cephOsdCmd.Flags().StringVar(&cephID, "id", "0", "id")
	cephOsdCmd.Flags().StringVar(&cephPublicAddress, "ip", "", "ip")
	cephRgwCmd.Flags().StringVar(&cephID, "id", "0", "id")
	cephRgwCmd.Flags().StringVar(&cephPublicAddress, "ip", "", "ip")
	cephRgwCmd.Flags().StringVar(&cephSSLCertificate, "ssl-certificate", SSL_CERTIFICATE, "SSL Certificate")
	cephRgwCmd.Flags().StringVar(&cephSSLKey, "ssl-key", SSL_KEY, "SSL Key")
	cephRgwCmd.Flags().StringVar(&cephProxyPort, "proxy-port", "15728", "Proxy Port")
	cephSetupCmd.Flags().StringVar(&cephDashboardUsername, "dashboard-username", "", "Dashboard username")
	cephSetupCmd.Flags().StringVar(&cephDashboardPassword, "dashboard-password", "", "Dashboard password")
	cephSetupCmd.Flags().StringVar(&cephRadosgwUsername, "radosgw-username", "", "Rados Gateway username")
	cephSetupCmd.Flags().StringVar(&cephRadosgwPassword, "radosgw-password", "", "Rados Gateway password")
	cephSetupCmd.Flags().StringVar(&cephSSLCertificate, "ssl-certificate", SSL_CERTIFICATE, "SSL Certificate")
	cephSetupCmd.Flags().StringVar(&cephSSLKey, "ssl-key", SSL_KEY, "SSL Key")
	cephSetupCmd.Flags().UintVar(&cephPlacementGroups, "placement-groups", 256, "The number of placement groups")
	cephSetupCmd.Flags().UintVar(&cephExpectedNumberOfObjects, "expected-number-of-objects", 1000000, "The number of expected objects")

	cephCmd.AddCommand(cephInitializeCmd)
	cephCmd.AddCommand(cephMgrCmd)
	cephCmd.AddCommand(cephMonCmd)
	cephCmd.AddCommand(cephMdsCmd)
	cephCmd.AddCommand(cephOsdCmd)
	cephCmd.AddCommand(cephRgwCmd)
	cephCmd.AddCommand(cephSetupCmd)

	cephCmd.Flags().StringVar(&cephBinariesPath, "ceph-binaries-path", ceph.CephBinariesPath, "Location of Ceph binaries")
	cephCmd.Flags().StringVar(&cephConfigPath, "ceph-config-path", ceph.CephConfigPath, "Location of Ceph config")
	cephCmd.Flags().StringVar(&cephDataPath, "ceph-data-path", ceph.CephDataPath, "Location of Ceph data")

	RootCmd.AddCommand(cephCmd)
}
