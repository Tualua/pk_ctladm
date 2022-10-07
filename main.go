package main

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	scst "github.com/Tualua/pk_ctladm/pk_scst"
	"github.com/akamensky/argparse"
	"github.com/sirupsen/logrus"
)

var log = logrus.New()

func GetDevList(xFlag bool) {
	DevList := [][]string{}
	if LunIds, err := GetLunIds(); err != nil {
		err = fmt.Errorf("cannot get LUNs IDs: %w", err)
		log.Errorf("GetDevList: %w", err)
		fmt.Println(err)
	} else {
		if Devices, err := scst.ScstGetDevices(); err != nil {
			err = fmt.Errorf("cannot get devices: %w", err)
			log.Errorf("GetDevList: %w", err)
			fmt.Println(err)
		} else {
			DevicesFiltered := []string{}
			for _, dev := range Devices {
				if !strings.Contains(dev, ":") {
					DevicesFiltered = append(DevicesFiltered, dev)
				}
			}
			for _, dev := range DevicesFiltered {
				if params, err := scst.ScstGetDeviceParams(dev); err != nil {
					err = fmt.Errorf("cannot get device %s parameters: %w", dev, err)
					log.Errorf("GetDevList: %w", err)
					fmt.Println(err)
				} else {
					if relId, ok := LunIds[params["filename"]]; ok {
						device := []string{
							relId,
							"block",
							params["size"],
							params["blocksize"],
							params["usn"],
							filepath.Base(params["filename"]),
							params["filename"],
							FindDeviceWwn(dev),
							params["threads_num"],
						}
						DevList = append(DevList, device)
						log.Trace(device)
					}
				}
			}
			if xFlag {
				XmlDevList := new(CtldLunList)
				for _, device := range DevList {
					XmlDevList.Luns = append(XmlDevList.Luns, LunFromSlice(device))
				}
				if outXml, err := xml.MarshalIndent(XmlDevList, "", "        "); err != nil {
					err = fmt.Errorf("error marshalling to XML. %s", err)
					log.Errorf("GetDevList: %w", err)
					fmt.Println(err)
				} else {
					log.Trace("XML Output:")
					log.Trace(string(outXml))
					fmt.Println(string(outXml))
				}
			} else {
				for _, device := range DevList {
					fmt.Println(strings.Join(device, "\t"))
				}
			}
		}
	}
}

func GetPortList(xFlag bool) {
	PortList := [][]string{}
	if LunIds, err := GetLunIds(); err != nil {
		err = fmt.Errorf("error getting LUN IDs: %w", err)
		log.Errorf("GetPortList: %w", err)
		fmt.Println(err)
	} else {
		if Devices, err := scst.ScstGetDevices(); err != nil {
			err = fmt.Errorf("error getting devices %w", err)
			log.Errorf("GetPortList: %w", err)
			fmt.Println(err)
		} else {
			DevicesFiltered := []string{}
			for _, dev := range Devices {
				if !strings.Contains(dev, ":") {
					DevicesFiltered = append(DevicesFiltered, dev)
				}
			}
			for _, dev := range DevicesFiltered {
				if params, err := scst.ScstGetDeviceParams(dev); err != nil {
					err = fmt.Errorf("error getting device %s parameters %s", dev, err)
					log.Errorf("GetPortList: %w", err)
					log.Println(err)
				} else {
					if relId, ok := LunIds[params["filename"]]; ok {
						portActive := "NO"
						if params["active"] == "1" {
							portActive = "YES"
						}
						wwn := FindDeviceWwn(dev)
						port := []string{
							relId,
							portActive,
							"iscsi",
							"iscsi",
							fmt.Sprintf("%s,t,0x0101", wwn),
							wwn,
						}
						log.Trace(port)
						PortList = append(PortList, port)
					}
				}
			}
			if xFlag {
				XmlPortList := new(CtldPortList)
				for _, device := range PortList {
					XmlPortList.Ports = append(XmlPortList.Ports, PortFromSlice(device))
				}
				if outXml, err := xml.MarshalIndent(XmlPortList, "", "        "); err != nil {
					err = fmt.Errorf("error marshalling to XML. %s", err)
					log.Errorf("GetPortList: %w", err)
					log.Println(err)
				} else {
					log.Trace("XML Output:")
					log.Trace(string(outXml))
					fmt.Println(string(outXml))
				}
			} else {
				for _, device := range PortList {
					fmt.Println(strings.Join(device, "\t"))
				}
			}
		}
	}
}

func RemoveLun(lun string) {
	if device, err := FindLunDevice(lun); err != nil {
		err = fmt.Errorf("cannot find corresponding device for LUN %s: %w", lun, err)
		log.Errorf("RemoveLun: %w", err)
		fmt.Println(err)
	} else {
		if err := scst.ScstDeactivateDevice(device); err != nil {
			err = fmt.Errorf("failed to deactivate device %s", lun)
			log.Errorf("RemoveLun: %w", err)
			fmt.Println(err)
		} else {
			msgInfo := fmt.Sprintf("LUN %s (%s) deactivated", lun, device)
			log.Info(msgInfo)
			fmt.Println(msgInfo)
		}

	}
}

func CreateLun(dev string) {
	if err := scst.ScstActivateDevice(dev); err != nil {
		err = fmt.Errorf("failed to activate device %s: %w", dev, err)
		log.Errorf("CreateLun: %w", err)
	} else {
		msgInfo := fmt.Sprintf("Device %s activated", dev)
		log.Info(msgInfo)
		fmt.Println(msgInfo)
	}
}

func init() {
	var (
		logFilePath string
	)
	log.SetFormatter(&logrus.TextFormatter{})
	log.SetLevel(logrus.DebugLevel)
	if os.Getenv("CTLADM_DEBUG") == "true" {
		logFilePath = "ctladm.log"
		fmt.Println("WARNING! Running in development environment")
	} else {
		logFilePath = "/var/log/ctladm.log"
	}

	if logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666); err != nil {
		err = fmt.Errorf("failed to log to file, using stderr: %w", err)
		log.Infof("init: %w", err)
		fmt.Println(err)

	} else {
		log.SetOutput(logFile)
	}
}

func main() {
	var (
		err error
	)
	parser := argparse.NewParser("ctladm", "Replacement for ctladm for Linux Playkey SDS")

	parserDevlist := parser.NewCommand("devlist", "List devices")
	argDevListXml := parserDevlist.Flag("x", "xml", &argparse.Options{Help: "Enable XML Output"})

	parserPortlist := parser.NewCommand("portlist", "List ports")
	argPortListXml := parserPortlist.Flag("x", "xml", &argparse.Options{Help: "Enable XML Output"})

	parserRemove := parser.NewCommand("remove", "Remove port")
	argRemoveB := parserRemove.String("b", "b", &argparse.Options{Help: "Accepts only \"block\""})
	argRemoveLun := parserRemove.String("l", "lun", &argparse.Options{Help: "LUN ID"})

	parserCreate := parser.NewCommand("create", "Create port")
	argCreateB := parserCreate.String("b", "b", &argparse.Options{Help: "Accepts only \"block\""})
	argCreateOptions := parserCreate.StringList("o", "options", &argparse.Options{Help: "Options"})
	argCreateDevice := parserCreate.String("d", "device", &argparse.Options{Help: "Device ID"})
	argCreateLun := parserCreate.String("l", "lun", &argparse.Options{Help: "LUN ID"})

	if err = parser.Parse(os.Args); err != nil {
		fmt.Println(parser.Usage(err))
	} else {
		if parserDevlist.Happened() {
			log.Debug("Command: devlist")
			log.Debug("Arguments:")
			log.Debug("-x:", *argDevListXml)
			GetDevList(*argDevListXml)
		} else if parserPortlist.Happened() {
			log.Debug("Command: portlist")
			log.Debug("Arguments:")
			log.Debug("-x:", *argPortListXml)
			GetPortList(*argPortListXml)
		} else if parserRemove.Happened() {
			log.Debug("Command: remove")
			log.Debug("Arguments:")
			log.Debug("-b:", *argRemoveB)
			log.Debug("-l:", *argRemoveLun)
			RemoveLun(*argRemoveLun)
		} else if parserCreate.Happened() {
			log.Debug("Command: create")
			log.Debug("Arguments:")
			log.Debug("-b:", *argCreateB)
			log.Debug("-o:", *argCreateOptions)
			log.Debug("-d:", *argCreateDevice)
			log.Debug("-l:", *argCreateLun)
			CreateLun(*argCreateDevice)
		}
	}

}
