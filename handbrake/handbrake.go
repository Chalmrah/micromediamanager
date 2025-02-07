package handbrake

import (
	"os/exec"
)

func Reencode() {

}

func verifyHandbrakeInstalled() bool {
	_, err := exec.LookPath("handbrakecli")
	return err == nil
}
