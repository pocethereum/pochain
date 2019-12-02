// Copyright 2015 Satoshi Konno. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package minedev

import (
	"crypto/md5"
	"github.com/pocethereum/pochain/common/hexutil"
	"github.com/pocethereum/pochain/log"
	"github.com/cybergarage/go-net-upnp/net/upnp"
	"io/ioutil"
	"os"
	"strings"
	"time"
)

const (
	DefaultTarget = "Living"
	DefaultStatus = true
)

type UPnPDevice struct {
	*upnp.Device
	Target string
	Status bool
}

var Instance *UPnPDevice

func NewLightDevice() (*UPnPDevice, error) {

	devBytes, err := ioutil.ReadFile(DefaultDescription)
	if err != nil {
		log.Info("NewLightDevice failed, read description file error", "error", err.Error())
		return nil, err
	}

	dev, err := upnp.NewDeviceFromDescription(string(devBytes))
	if err != nil {
		log.Info("NewDeviceFromDescription failed", "error", err.Error())
		return nil, err
	}

	lightDev := &UPnPDevice{
		Device: dev,
		Target: DefaultTarget,
		Status: DefaultStatus,
	}
	lightDev.RefreshDeviceData()

	log.Info("init device success", "dev", *dev.DeviceDescription)
	return lightDev, nil
}

func getEncodedBindUserName(username string) (encUser string) {
	switch len(username) {
	case 0:
		return "****"
	case 1:
		return "***" + username
	case 2, 3, 4, 5:
		return "**" + username
	default:
		return username[0:5] + "**" + username[len(username)-2:]
	}
	return
}

func (self *UPnPDevice) RefreshDeviceData() {
	dev := self.Device
	mac, ip, err := getMacAndIp()
	if err != nil {
		log.Info("getMacAndIp failed", "error", err.Error())
		return
	}
	dev.SerialNumber = mac
	dev.URLBase = "http://" + ip + ":8545"

	h := Host{}
	h.QueryHostInfo()
	uo := User{}
	binduser, err := uo.QueryBindUser()
	if err != nil {
		log.Info("QueryBindUser failed", "error", err.Error())
		return
	}

	hash := md5.Sum([]byte(strings.ToLower(binduser.Username)))
	dev.BindUser = getEncodedBindUserName(binduser.Username)
	dev.BindUserHash = hexutil.Encode(hash[:])

	dev.FriendlyName = h.Hostname
	fillDiskInfo(&dev.DiskInfo)
}

func (self *UPnPDevice) ActionRequestReceived(action *upnp.Action) upnp.Error {
	log.Info("UPnPDevice ActionRequestReceived", "action", *action)
	return upnp.NewErrorFromCode(upnp.ErrorOptionalActionNotImplemented)
}

func GetInstance() *UPnPDevice {
	if Instance == nil {
		var err error
		if Instance, err = NewLightDevice(); err != nil {
			panic("New UPnPDevice failed, error:" + err.Error())
		}
	}
	return Instance
}

func Start() {
	log.Info("MineUpnpDevice start ...")

	dev := GetInstance()

	if dev == nil {
		log.Info("Get MineUpnpDevice error")
		os.Exit(1)
	}

	try := 0

	for try < 5 {
		try++
		err := dev.Start()
		if err == nil {
			log.Info("MineUpnpDevice start success", "try", try)
			break
		}
		log.Info("MineUpnpDevice start err error, will retry", "err", err.Error(), "try", try)

		if try == 5 {
			os.Exit(1)
		}
		time.Sleep(time.Second * 3)
	}
}
