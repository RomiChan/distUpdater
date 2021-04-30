package main

import (
	"bytes"
	_ "embed"
	"log"
	"os"
	"os/exec"
	"strings"
)

const (
	Unsupport int8 = iota
	CentOS
	Ubuntu
	Debian
)

//go:embed update.sh
var script []byte

func main() {
	log.SetPrefix("distUpdater: ")
	OSType := checkOS()
	OSName := ""

	switch OSType {
	case Unsupport:
		log.Println("不支持的操作系统，请更换为CentOS, Ubuntu, Debian")
		return
	case CentOS:
		OSName = "CentOS"
	case Ubuntu:
		OSName = "Ubuntu"
	case Debian:
		OSName = "Debian"
	}

	log.Println("检测到您的系统为:", OSName)

	err := installDep(OSType)
	if err != nil {
		log.Fatalf("构建工具安装失败: %v，请手动安装仓库构建工具", err)
	}

	err = updateRepo()
	if err != nil {
		log.Fatalf("仓库更新失败: %v", err)
	}
	err = push()
	if err != nil {
		log.Fatalf("仓库推送失败: %v", err)
	}
}

func checkOS() int8 {
	file, err := os.ReadFile("/etc/os-release")
	if err != nil {
		return Unsupport
	}
	if bytes.Contains(file, []byte("CentOS")) {
		return CentOS
	}
	if bytes.Contains(file, []byte("Ubuntu")) {
		return Ubuntu
	}
	if bytes.Contains(file, []byte("Debian")) {
		return Debian
	}
	return Unsupport
}

func installDep(OSType int8) error {
	pm := ""
	switch OSType {
	case Ubuntu, Debian:
		pm = "apt"
	case CentOS:
		pm = "dnf"
	}
	log.Println("正在安装仓库构建工具...")
	log.Printf("Running sudo %s install reprepro createrepo -y\n", pm)
	c := exec.Command("sudo", pm, "install", "reprepro", "createrepo", "-y")
	c.Stderr = os.Stderr
	err := c.Run()
	if err != nil {
		return err
	}
	log.Println("构建工具安装成功")
	return nil
}

func updateRepo() error {
	downDir, err := os.ReadDir("./download")
	if err != nil {
		return err
	}
	repreproArgs := []string{"-C", "main", "-b", "./deb", "includedeb", "nini"}
	for i := range downDir {
		fName := downDir[i].Name()
		if !downDir[i].IsDir() {
			if strings.HasSuffix(fName, ".rpm") {
				err = os.Rename("./download/"+fName, "./rpm/dist/"+fName)
				if err != nil {
					return err
				}
				continue
			}
			if strings.HasSuffix(fName, ".deb") {
				repreproArgs = append(repreproArgs, "./download/"+fName)
				continue
			}
		}
	}

	log.Println("正在更新APT仓库...")
	log.Println("Running reprepro -C main -b ./deb includedeb nini ./download/*.deb")
	c := exec.Command("reprepro", repreproArgs...)
	c.Stderr = os.Stderr
	err = c.Run()
	if err != nil {
		return err
	}
	log.Println("APT仓库更新完成")

	log.Println("正在更新RPM仓库...")

	c = exec.Command("createrepo", "--update", "./rpm/dist")
	c.Stderr = os.Stderr
	err = c.Run()
	if err != nil {
		return err
	}
	log.Println("RPM仓库更新完成")
	return nil
}

func push() error {
	log.Println("正在推送仓库...")
	f, err := os.Create("__tmp.sh")
	defer os.Remove("__tmp.sh")
	if err != nil {
		return err
	}
	_, err = f.Write(script)
	if err != nil {
		return err
	}
	f.Close()

	c := exec.Command("bash", "__tmp.sh")
	c.Stderr = os.Stderr
	err = c.Run()
	if err != nil {
		return err
	}
	return nil
}
