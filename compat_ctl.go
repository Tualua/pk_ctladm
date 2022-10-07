package main

import (
	"fmt"
	"path/filepath"
	"strings"

	scst "github.com/Tualua/pk_ctladm/pk_scst"
)

func GetLunIds() (res map[string]string, err error) {
	res = map[string]string{}
	if targets, err := scst.ScstGetIscsiTargets(); err != nil {
		log.Errorf("GetLunIds: cannot get iscsi targets: %w", err)
	} else {
		for _, target := range targets {
			lun0Device, _ := scst.ScstGetLunDevice(target, 0)
			lun0RelId, _ := scst.ScstGetIscsiTargetParam(target, "rel_tgt_id")
			res[lun0Device.Filename] = lun0RelId
		}
	}
	return
}

func FindLunDevice(lun string) (device string, err error) {
	wwns := make(map[string]string)
	if targets, err := scst.ScstGetIscsiTargets(); err != nil {
		log.Errorf("FindLunDevice: cannot get iscsi targets. %w", err)
	} else {
		for _, target := range targets {
			if relId, err := scst.ScstGetIscsiTargetParam(target, "rel_tgt_id"); err != nil {
				log.Errorf("FindLunDevice: cannot get relative id for target %s: %w", target, err)
			} else {
				wwns[relId] = target
			}
		}

		if target, ok := wwns[lun]; ok {
			if blockDevice, err := scst.ScstGetLunDevice(target, 0); err != nil {
				log.Errorf("FindLunDevice: cannot get LUN device filename: %w", err)
			} else {
				device = blockDevice.Name
			}
		} else {
			log.Errorf("FindLunDevice: LUN %s not found", lun)
		}

	}
	return
}

func FindDeviceWwn(device string) (wwn string) {
	wwns := []string{}
	pathExports := fmt.Sprintf("%s/%s/exported", scst.SCST_DEVICES, device)
	if exports, err := scst.ReadFromDir(pathExports); err != nil {
		log.Errorf("FindDeviceWwn: error getting exports for device %s: %w", device, err)
	} else {
		if len(exports) > 0 {
			for _, export := range exports {
				pathExport := fmt.Sprintf("%s/%s/exported/%s", scst.SCST_DEVICES, device, export)
				absPathExport, _ := filepath.EvalSymlinks(pathExport)
				if strings.Contains(absPathExport, "iscsi") {
					wwns = append(wwns, strings.Split(absPathExport, "/")[6])
				}
			}
			if len(wwns) > 0 {
				wwn = wwns[0]
			}
		}
	}
	return
}
