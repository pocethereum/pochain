// Copyright 2015 Satoshi Konno. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package upnp

import (
	"encoding/xml"
	"errors"
)

const (
	errorDeviceDescriptionNullDevice = "device is null"
)

// A DeviceDescriptionRoot represents a root UPnP device description.
type DeviceDescriptionRoot struct {
	XMLName     xml.Name    `xml:"root"`
	SpecVersion SpecVersion `xml:"specVersion"`
	URLBase     string      `xml:"URLBase"`
	Device      Device      `xml:"device"`
}

// A DeviceDescription represents a UPnP device description.
type DeviceDescription struct {
	XMLName          xml.Name    `xml:"device"`
	DeviceType       string      `xml:"deviceType"`
	FriendlyName     string      `xml:"friendlyName"`
	Manufacturer     string      `xml:"manufacturer"`
	ManufacturerURL  string      `xml:"manufacturerURL"`
	ModelDescription string      `xml:"modelDescription"`
	ModelName        string      `xml:"modelName"`
	ModelNumber      string      `xml:"modelNumber"`
	ModelURL         string      `xml:"modelURL"`
	SerialNumber     string      `xml:"serialNumber"`
	BindUser         string      `xml:"bindUser"`
	BindUserHash     string      `xml:"bindUserHash"`
	DiskInfo         DiskInfo    `xml:"diskInfo"`
	UDN              string      `xml:"UDN"`
	UPC              string      `xml:"UPC"`
	PresentationURL  string      `xml:"presentationURL"`
	IconList         IconList    `xml:"iconList"`
	ServiceList      ServiceList `xml:"serviceList"`
	DeviceList       DeviceList  `xml:"deviceList"`
}

type DiskInfo struct {
	All      uint64 `json:"all" xml:"all"`
	Used     uint64 `json:"used" xml:"used"`
	Free     uint64 `json:"free" xml:"free"`
	Count    int    `json:"count" xml:"count"`
}

// A DeviceList represents a ServiceList.
type DeviceList struct {
	XMLName xml.Name `xml:"deviceList"`
	Devices []Device `xml:"device"`
}

func NewDeviceDescriptionRoot() *DeviceDescriptionRoot {
	root := &DeviceDescriptionRoot{}
	specVer := NewSpecVersion()
	root.SpecVersion = (*specVer)
	return root
}

func NewDeviceDescriptionRootFromDevice(dev *Device) (*DeviceDescriptionRoot, error) {
	if dev == nil {
		return nil, errors.New(errorDeviceDescriptionNullDevice)
	}
	root := NewDeviceDescriptionRoot()
	root.Device = (*dev)
	root.URLBase = dev.URLBase
	return root, nil
}

func NewDeviceDescriptionRootFromString(descStr string) (*DeviceDescriptionRoot, error) {
	root := NewDeviceDescriptionRoot()
	err := xml.Unmarshal([]byte(descStr), root)
	if err != nil {
		return nil, err
	}

	err = root.Device.reviseParentObject()
	if err != nil {
		return nil, err
	}

	return root, err
}
