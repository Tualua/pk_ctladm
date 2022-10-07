package main

import (
	"encoding/xml"
	"strings"

	scst "github.com/Tualua/pk_ctladm/pk_scst"
)

type CtldLun struct {
	XMLName      xml.Name `xml:"lun"`
	Id           string   `xml:"id,attr"`
	BackendType  string   `xml:"backend_type"`
	LunType      int      `xml:"lun_type"`
	Size         string   `xml:"size"`
	Blocksize    string   `xml:"blocksize"`
	SerialNumber string   `xml:"serial_number"`
	DeviceId     string   `xml:"device_id"`
	NumThreads   string   `xml:"num_threads"`
	File         string   `xml:"file"`
	CtldName     string   `xml:"ctld_name"`
}

type CtldLunList struct {
	XMLName xml.Name  `xml:"ctllunlist"`
	Luns    []CtldLun `xmlname:"lun"`
}

type CtldPort struct {
	XMLName   xml.Name    `xml:"targ_port"`
	Id        string      `xml:"id,attr"`
	Lun       CtldPortLun `xml:"lun"`
	Target    string      `xml:"target"`
	Initiator string      `xml:"initiator"`
}

type CtldPortLun struct {
	XMLName xml.Name `xml:"lun"`
	Id      int      `xml:"id,attr"`
	Value   string   `xml:",chardata"`
}

type CtldPortList struct {
	XMLName xml.Name   `xml:"ctlportlist"`
	Ports   []CtldPort `xmlname:"targ_port"`
}

func LunFromSlice(device []string) (lun CtldLun) {
	lun.Id = device[0]
	lun.BackendType = device[1]
	lun.LunType = 0
	lun.Size = device[2]
	lun.Blocksize = device[3]
	lun.SerialNumber = device[4]
	lun.DeviceId = device[5]
	lun.NumThreads = device[8]
	lun.File = device[6]
	lun.CtldName = strings.Join([]string{
		device[7],
		"lun",
		"0",
	}, ",")
	return
}

func PortFromSlice(device []string) (port CtldPort) {

	port.Id = device[0]
	port.Lun = CtldPortLun{
		Id:    0,
		Value: device[0],
	}
	port.Target = device[5]
	initiators := scst.ScstGetIscsiTargetSessions(device[5])
	if len(initiators) > 0 {
		port.Initiator = initiators[0]
	}
	return
}
