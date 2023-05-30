// SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
// SPDX-License-Identifier: Apache-2.0

package main

import "encoding/xml"

type Response struct {
	XMLName xml.Name `xml:"List"`
	Text    string   `xml:",chardata"`
	Items   []struct {
		Text                          string `xml:",chardata"`
		Index                         string `xml:"Index"`
		IPAddress                     string `xml:"IPAddress"`
		MACAddress                    string `xml:"MACAddress"`
		Active                        string `xml:"Active"`
		HostName                      string `xml:"HostName"`
		InterfaceType                 string `xml:"InterfaceType"`
		XAVMDEPort                    string `xml:"X_AVM-DE_Port"`
		XAVMDESpeed                   string `xml:"X_AVM-DE_Speed"`
		XAVMDEUpdateAvailable         string `xml:"X_AVM-DE_UpdateAvailable"`
		XAVMDEUpdateSuccessful        string `xml:"X_AVM-DE_UpdateSuccessful"`
		XAVMDEInfoURL                 string `xml:"X_AVM-DE_InfoURL"`
		XAVMDEMACAddressList          string `xml:"X_AVM-DE_MACAddressList"`
		XAVMDEModel                   string `xml:"X_AVM-DE_Model"`
		XAVMDEURL                     string `xml:"X_AVM-DE_URL"`
		XAVMDEGuest                   string `xml:"X_AVM-DE_Guest"`
		XAVMDERequestClient           string `xml:"X_AVM-DE_RequestClient"`
		XAVMDEVPN                     string `xml:"X_AVM-DE_VPN"`
		XAVMDEWANAccess               string `xml:"X_AVM-DE_WANAccess"`
		XAVMDEDisallow                string `xml:"X_AVM-DE_Disallow"`
		XAVMDEIsMeshable              string `xml:"X_AVM-DE_IsMeshable"`
		XAVMDEPriority                string `xml:"X_AVM-DE_Priority"`
		XAVMDEFriendlyName            string `xml:"X_AVM-DE_FriendlyName"`
		XAVMDEFriendlyNameIsWriteable string `xml:"X_AVM-DE_FriendlyNameIsWriteable"`
	} `xml:"Item"`
}
