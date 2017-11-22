package util

import (
	"os"
	"os/exec"
)

func WorkDir() string {
	WorkDir, err := os.Getwd()
	if err != nil {
		return ""
	}
	return WorkDir
}

func Command(name string, args ...string) *exec.Cmd {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd
}

func Symlink(oldname, newname string) error {
	_, err := os.Stat(oldname)
	if err != nil {
		return err
	}
	return os.Symlink(oldname, newname)
}
