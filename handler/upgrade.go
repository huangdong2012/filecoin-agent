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
		err = h.operateServices("start", cmd.Services)
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

func (h *upgradeHandler) copyFiles(base, dest string, pkg *model.Package) error {
	var (
		err   error
		file  *os.File
		infos []os.FileInfo
	)
	if file, err = os.Open(dest); err != nil {
		return err
	}
	defer file.Close()

	if infos, err = file.Readdir(-1); err != nil {
		return err
	}
	defer func() {
		if err == nil {
			_, err = infras.ExecCommand("rm", "-rf", filepath.Join(base, "*.bak"))
		}
	}()

	for _, info := range infos {
		from := filepath.Join(dest, info.Name())
		to := filepath.Join(base, info.Name())
		//1.backup and remove
		if infras.PathExist(to) {
			if _, err = infras.ExecCommand("cp", "-rf", to, to+".bak"); err != nil {
				return err
			}
			if err = os.RemoveAll(to); err != nil {
				return err
			}
		}

		//2.copy or rollback
		if _, err = infras.ExecCommand("cp", "-rf", from, to); err != nil {
			if _, err2 := infras.ExecCommand("mv", to+".bak", to); err2 != nil {
				return fmt.Errorf("mutil errors: %v\n%v", err, err2)
			}
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

	return infos[0].Name(), nil
}
