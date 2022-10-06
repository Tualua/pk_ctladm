package pk_scst

import (
	"errors"
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strings"
)

const SCST_ROOT_PATH string = "/sys/kernel/scst_tgt"
const SCST_DEVICES string = SCST_ROOT_PATH + "/devices"
const SCST_ISCSI_TARGETS string = SCST_ROOT_PATH + "/targets/iscsi"
const SYSFS_SCST_DEV_MGMT string = SCST_ROOT_PATH + "/handlers/vdisk_blockio/mgmt"
const SYSFS_SCST_LUNS_MGMT string = "/ini_groups/allowed_ini/luns/mgmt"
const SYSFS_SCST_LUN0_DEV string = "ini_groups/allowed_ini/luns/0/device"

func readParamsFromDir(dirpath string) (map[string]string, error) {
	var (
		res map[string]string
		err error
		val []byte
	)
	res = make(map[string]string)
	if dataDir, err := os.Open(dirpath); err != nil {
		return res, err
	} else {
		if files, err := dataDir.ReadDir(0); err != nil {
			return res, err
		} else {
			for _, v := range files {
				if f, err := os.Open(path.Join(dirpath, v.Name())); err != nil {
					res[v.Name()] = err.Error()
				} else {
					if val, err = ioutil.ReadAll(f); err != nil {
						res[v.Name()] = err.Error()
					}
					res[v.Name()] = strings.TrimSuffix(string(val), "\n")
					res[v.Name()] = strings.TrimSuffix(res[v.Name()], "\n[key]")
					res[v.Name()] = strings.TrimSuffix(res[v.Name()], "\n[key]\n")
					res[v.Name()] = strings.TrimSuffix(res[v.Name()], "\n")
				}
			}
		}
	}
	return res, err
}

func readFromDir(dirpath string) ([]string, error) {
	var (
		res []string
		err error
	)

	if dataDir, err := os.Open(dirpath); err != nil {
		return res, err
	} else {
		if files, err := dataDir.ReadDir(0); err != nil {
			return res, err
		} else {
			for _, v := range files {
				res = append(res, v.Name())
			}
		}
	}
	return res, err
}

func listSubDirs(dirPath string) (res []string, err error) {
	var (
		dir        *os.File
		dirContent []fs.DirEntry
	)
	if dir, err = os.Open(dirPath); err != nil {
		log.Println(err.Error())
	} else {
		if dirContent, err = dir.ReadDir(0); err != nil {
			log.Println(err.Error())
		} else {
			for _, v := range dirContent {
				if v.IsDir() {
					res = append(res, v.Name())
				}
			}
		}
	}
	return
}

func ScstGetDevices() ([]string, error) {
	var (
		res         []string
		scstDevices []fs.DirEntry
		err         error
	)
	if scstDevicesDir, err := os.Open(SCST_DEVICES); err != nil {
		log.Println(err.Error())
	} else {
		if scstDevices, err = scstDevicesDir.ReadDir(0); err != nil {
			log.Println(err.Error())
		} else {
			for _, v := range scstDevices {
				res = append(res, v.Name())
			}
		}
	}
	return res, err
}

func ScstGetIscsiTargets() (res []string, err error) {
	if res, err = listSubDirs(SCST_ISCSI_TARGETS); err != nil {
		log.Println(err.Error())
	}
	return res, err
}

func ScstFindWwn(tgtId string) (res string, err error) {
	var (
		targets []string
	)
	if targets, err = ScstGetIscsiTargets(); err != nil {
		log.Println(err.Error())
	} else {
		for _, v := range targets {
			tgtId2 := strings.Split(v, ":")
			if tgtId == tgtId2[len(tgtId2)-1] {
				res = v
				break
			}
		}
		if res == "" {
			err = fmt.Errorf("target with id %s not found", tgtId)
		}
	}
	return
}

func ScstGetDeviceParam(device string, param string) string {
	var (
		paramPath string
		res       string
	)

	paramPath = SCST_DEVICES + "/" + device + "/" + param
	deviceParam, err := os.Open(paramPath)
	if err != nil {
		res = ""
	} else {
		deviceParamData, err := ioutil.ReadAll(deviceParam)
		if err != nil {
			res = ""
		}
		res = strings.Split(string(deviceParamData), "\n")[0]
	}
	return res

}

func ScstGetIscsiTargetParam(wwn string, param string) (res string, err error) {
	var (
		paramPath string = SCST_ISCSI_TARGETS + "/" + wwn + "/" + param
	)

	if deviceParam, err := os.Open(paramPath); err != nil {
		res = ""
	} else {
		if deviceParamData, err := ioutil.ReadAll(deviceParam); err != nil {
			res = ""
		} else {
			res = strings.Split(string(deviceParamData), "\n")[0]
		}
	}
	return
}

func scstSetDeviceParam(device string, param string, val string) (err error) {
	var (
		paramPath string = SCST_DEVICES + "/" + device + "/" + param
	)

	if deviceParam, err := os.OpenFile(paramPath, os.O_WRONLY, 0644); err != nil {
		log.Println(err)
	} else {
		defer deviceParam.Close()
		deviceParamW, err := deviceParam.Write([]byte(val))
		if err != nil {
			log.Println(err)
		}
		log.Printf("Wrote %d bytes.\n", deviceParamW)
	}
	return
}

func ScstGetDeviceParams(device string) (res map[string]string, err error) {

	if res, err = readParamsFromDir(path.Join(SCST_DEVICES, device)); err != nil {
		log.Println(err.Error())
	}
	return
}

func ScstGetIscsiTargetParams(target string) (res map[string]string, err error) {
	var (
		wwn string
	)
	res = make(map[string]string)
	if wwn, err = ScstFindWwn(target); err != nil {
		log.Println(err.Error())
	} else {
		if res, err = readParamsFromDir(path.Join(SCST_ISCSI_TARGETS, wwn)); err != nil {
			log.Println(err.Error())
		}
		res["wwn"] = wwn
	}
	return
}

func ScstDeleteDevice(device string) (err error) {

	scstCmd := []byte("del_device " + device)
	if mgmt, err := os.OpenFile(SYSFS_SCST_DEV_MGMT, os.O_WRONLY, 0644); err != nil {
		log.Println(err.Error())
	} else {
		defer mgmt.Close()
		if _, err := mgmt.Write(scstCmd); err != nil {
			log.Println(err.Error())
		} else {
			log.Printf("Device %s deleted \n", device)
		}
	}
	return
}

func ScstDeactivateDevice(device string) (err error) {
	var (
		mgmt *os.File
	)
	scstCmd := []byte("0")
	scstCmdPath := path.Join(SCST_DEVICES, device, "active")
	if mgmt, err = os.OpenFile(scstCmdPath, os.O_WRONLY, 0644); err != nil {
		log.Println(err.Error())
	} else {
		defer mgmt.Close()
		if _, err = mgmt.Write(scstCmd); err != nil {
			log.Println(err.Error())
		} else {
			log.Printf("Device %s deactivated \n", device)
		}
	}
	return
}

func ScstActivateDevice(device string) (err error) {
	var (
		mgmt *os.File
	)
	scstCmd := []byte("1")
	scstCmdPath := path.Join(SCST_DEVICES, device, "active")
	if mgmt, err = os.OpenFile(scstCmdPath, os.O_WRONLY, 0644); err != nil {
		log.Println(err.Error())
	} else {
		defer mgmt.Close()
		if _, err = mgmt.Write(scstCmd); err != nil {
			log.Println(err.Error())
		} else {
			log.Printf("Device %s activated \n", device)
		}
	}
	return
}

func ScstCreateLun(devId string, fileName string) (err error) {
	var (
		lunPathMgmt string
		wwn         string
		targets     []string
	)
	if targets, err = ScstGetIscsiTargets(); err != nil {
		log.Println(err.Error())
	} else {
		for _, v := range targets {
			if strings.Contains(v, devId) {
				wwn = v
				break
			}
		}
		if wwn != "" {
			if mgmt, err := os.OpenFile(SYSFS_SCST_DEV_MGMT, os.O_WRONLY, 0644); err != nil {
				err = errors.New("cannot open SCST device management interface")
				log.Println(err.Error())
			} else {
				defer mgmt.Close()
				scstCmd := "add_device " + devId + " filename=" + fileName + "; nv_cache=1; rotational=0"
				if _, err := mgmt.Write([]byte(scstCmd)); err != nil {
					log.Println(err.Error())
				} else {
					log.Printf("Device %s backed by %s added\n", devId, fileName)
					lunPathMgmt = SCST_ISCSI_TARGETS + "/" + wwn + SYSFS_SCST_LUNS_MGMT
					if mgmt, err = os.OpenFile(lunPathMgmt, os.O_WRONLY, 0644); err != nil {
						log.Println(err.Error())
					} else {
						defer mgmt.Close()
						scstCmd = "add " + devId + " 0"
						if _, err := mgmt.Write([]byte(scstCmd)); err != nil {
							log.Println(err.Error())
						} else {
							log.Printf("LUN 0 with device %s exported via target %s\n", devId, wwn)
							if err = scstSetDeviceParam(devId, "t10_vend_id", "FREE_TT"); err != nil {
								log.Println(err.Error())
							}
						}
					}
				}
			}
		} else {
			err = fmt.Errorf("target %s not found", devId)
			log.Println(err.Error())
		}
	}
	return err
}

func ScstListIscsiSessions(target string) (res []string, err error) {
	var (
		sessionsPath string
		wwn          string
	)
	if wwn, err = ScstFindWwn(target); err != nil {
		log.Println(err.Error())
	} else {
		sessionsPath = path.Join(SCST_ISCSI_TARGETS, wwn, "sessions")
		if res, err = readFromDir(sessionsPath); err != nil {
			log.Println(err.Error())
		}
	}

	return
}
