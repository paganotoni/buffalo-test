package test

import (
	"bytes"
	"os/exec"

	"github.com/gobuffalo/envy"
)

func removeFlag(flag string, args []string) []string {
	for i, v := range args {
		if v != flag {
			continue
		}

		args = append(args[:i], args[i+1:]...)
		break
	}

	return args
}

func testPackages() ([]string, error) {
	args := []string{}
	out, err := exec.Command(envy.Get("GO_BIN", "go"), "list", "./...").Output()

	if err != nil {
		return args, err
	}

	pkgs := bytes.Split(bytes.TrimSpace(out), []byte("\n"))
	for _, p := range pkgs {
		if !vendorRegex.Match(p) {
			args = append(args, string(p))
		}
	}
	return args, nil
}

func hasTestify(p string) bool {
	cmd := exec.Command("go", "test", "-thisflagdoesntexist")
	b, _ := cmd.Output()
	return bytes.Contains(b, []byte("-testify.m"))
}
