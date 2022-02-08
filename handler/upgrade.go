package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"grandhelmsman/filecoin-agent/infras"
	"grandhelmsman/filecoin-agent/model"
	"grandhelmsman/filecoin-agent/supd"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var (
	Upgrade = &upgradeHandler{}
)

type upgradeHandler struct {
}

func (h *upgradeHandler) Handle(msg *model.CommandRequest) (resp *model.CommandResponse, err error) {
	cmd := &model.UpgradeCommand{}
	if err = json.Unmarshal([]byte(msg.Body), cmd); err != nil {
		return nil, err
	}

	//1.setup
	base := "/root/hlm-miner"
	if len(cmd.TargetPath) > 0 {
		base = cmd.TargetPath
	}
	dir := filepath.Join(base, "upgrade")
	src := filepath.Join(dir, "src")
	dest := filepath.Join(dir, "dest")
	if err = os.RemoveAll(dest); err != nil {
		return nil, err
	}

	//2.download zip and decompress
	if zip, err := infras.DownloadToDir(cmd.SourceUrl, cmd.Username, cmd.Password, src); err != nil {
		return nil, err
	} else if err = infras.Decompress(zip, dest); err != nil {
		return nil, err
	}

	//3.get dest dir name
	if dest, err = h.getDestDirName(dest); err != nil {
		return nil, err
	}

	//4.parse package.json
	pkg := &model.Package{}
	pkgName := "package.json"
	if pkgPath := filepath.Join(dest, pkgName); infras.PathExist(pkgPath) {
		if data, err := ioutil.ReadFile(pkgPath); err == nil {
			if err = json.Unmarshal(data, pkg); err != nil {
				return nil, err
			}
		}
	}

	//5.1.stop and start services
	if err = h.operateServices("stop", cmd.Services); err != nil {
		return nil, err
	}
	defer func() {
		//5.2.start services
		if err == nil {
			err = h.operateServices("start", cmd.Services)
		}
	}()

	//6.copy files
	if err = h.copyFiles(base, dest, pkg); err != nil {
		return nil, err
	}

	return &model.CommandResponse{
		ID:         msg.ID,
		Host:       id,
		Status:     int(model.CommandStatus_Success),
		FinishTime: time.Now().Unix(),
	}, nil
}

func (h *upgradeHandler) copyFiles(base, dest string, pkg *model.Package) (err error) {
	var (
		file  *os.File
		infos []os.FileInfo
		baks  []string
	)
	if file, err = os.Open(dest); err != nil {
		return err
	}
	defer file.Close()

	if infos, err = file.Readdir(-1); err != nil {
		return err
	}
	defer func() {
		if err != nil { //rollback
			for _, bak := range baks {
				src := strings.TrimSuffix(bak, ".bak")
				if infras.PathExist(src) {
					if _, err2 := infras.ExecCommand("rm", "-rf", src); err2 != nil {
						fmt.Println("rollback rm src error:", err2)
					}
				}
				if _, err2 := infras.ExecCommand("mv", "-f", bak, src); err2 != nil {
					fmt.Println("rollback mv bak error:", err2)
				}
			}
		} else { //remove bak
			for _, bak := range baks {
				if _, err2 := infras.ExecCommand("rm", "-rf", bak); err2 != nil {
					fmt.Println("clear bak error:", err2)
				}
			}
		}
	}()

	for _, info := range infos {
		from := filepath.Join(dest, info.Name())
		to := filepath.Join(base, info.Name())
		//1.backup
		if infras.PathExist(to) {
			bak := to + ".bak"
			if _, err = infras.ExecCommand("cp", "-rf", to, bak); err != nil {
				return err
			} else {
				baks = append(baks, bak)
			}

			if pkg.Full { //全量更新
				if _, err = infras.ExecCommand("rm", "-rf", to); err != nil {
					return err
				}
			}
		}

		//2.copy
		if _, err = infras.ExecCommand("cp", "-rf", from, base); err != nil {
			return err
		}
	}

	return nil
}

func (h *upgradeHandler) operateServices(operate string, services []string) error {
	for _, srv := range services {
		if err := supd.Execute([]string{operate, srv}); err != nil {
			return err
		}
	}
	return nil
}

func (h *upgradeHandler) getDestDirName(dest string) (string, error) {
	var (
		err   error
		file  *os.File
		infos []os.FileInfo
	)
	if file, err = os.Open(dest); err != nil {
		return "", err
	}
	defer file.Close()

	if infos, err = file.Readdir(-1); err != nil {
		return "", err
	} else if len(infos) != 1 {
		return "", errors.New("dir count invalid of dest")
	}

	return filepath.Join(dest, infos[0].Name()), nil
}
