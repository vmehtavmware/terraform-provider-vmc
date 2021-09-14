/* Copyright 2020 VMware, Inc.
   SPDX-License-Identifier: MPL-2.0 */

package vmc

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	uuid "github.com/satori/go.uuid"
	"github.com/vmware/vsphere-automation-sdk-go/runtime/protocol/client"
	"github.com/vmware/vsphere-automation-sdk-go/services/vmc/model"
	"github.com/vmware/vsphere-automation-sdk-go/services/vmc/orgs"
)

var storageCapacityMap = map[string]int64{
	"15TB": 15003,
	"20TB": 20004,
	"25TB": 25005,
	"30TB": 30006,
	"35TB": 35007,
}

func GetSDDC(connector client.Connector, orgID string, sddcID string) (model.Sddc, error) {
	sddcClient := orgs.NewDefaultSddcsClient(connector)
	sddc, err := sddcClient.Get(orgID, sddcID)
	return sddc, err
}

func ConvertStorageCapacitytoInt(s string) int64 {
	storageCapacity := storageCapacityMap[s]
	return storageCapacity
}

// Mapping for deployment_type field
// During refresh/import state, return value of VMC API should be converted to uppercamel case in terraform
// to maintain consistency
func ConvertDeployType(s string) string {
	if s == "SINGLE_AZ" {
		return "SingleAZ"
	} else if s == "MULTI_AZ" {
		return "MultiAZ"
	} else {
		return ""
	}
}

func IsValidUUID(u string) error {
	_, err := uuid.FromString(u)
	if err != nil {
		return err
	}
	return nil
}

func IsValidURL(s string) error {
	_, err := url.ParseRequestURI(s)
	if err != nil {
		return err
	}
	return nil
}
func expandMsftLicenseConfig(config []interface{}) *model.MsftLicensingConfig {
	if len(config) == 0 {
		return nil
	}
	var licenseConfig model.MsftLicensingConfig
	licenseConfigMap := config[0].(map[string]interface{})
	mssqlLicensing := strings.ToUpper(licenseConfigMap["mssql_licensing"].(string))
	windowsLicensing := strings.ToUpper(licenseConfigMap["windows_licensing"].(string))
	licenseConfig = model.MsftLicensingConfig{MssqlLicensing: &mssqlLicensing, WindowsLicensing: &windowsLicensing}
	return &licenseConfig
}

func getNSXTReverseProxyURLConnector(nsxtReverseProxyUrl string) (client.Connector, error) {
	apiToken := os.Getenv(APIToken)
	if len(nsxtReverseProxyUrl) == 0 {
		return nil, fmt.Errorf("NSX reverse proxy url is required for public IP resource creation")
	}
	nsxtReverseProxyUrl = strings.Replace(nsxtReverseProxyUrl, SksNSXTManager, "", -1)
	httpClient := http.Client{}
	connector, err := NewClientConnectorByRefreshToken(apiToken, nsxtReverseProxyUrl, DefaultCSPUrl, httpClient)
	if err != nil {
		return nil, HandleCreateError("NSXT reverse proxy URL connector", err)
	}
	return connector, nil
}

func getTotalSddcHosts(sddc *model.Sddc) int {
	totalHosts := 0
	if sddc != nil && sddc.ResourceConfig.Clusters != nil {
		for _, cluster := range sddc.ResourceConfig.Clusters {
			totalHosts += len(cluster.EsxHostList)
		}
	}
	return totalHosts
}
