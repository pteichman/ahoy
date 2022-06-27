package ahoy

import (
	"errors"
	"io/ioutil"
	"os"
	"os/exec"
)

func Edit(data []byte) ([]byte, error) {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		return nil, errors.New("no EDITOR environment variable")
	}

	file, err := ioutil.TempFile("", "ahoy_*")
	if err != nil {
		return nil, err
	}
	file.Close()
	defer os.Remove(file.Name())

	if err := ioutil.WriteFile(file.Name(), data, 0600); err != nil {
		return nil, err
	}

	cmd := exec.Command(editor, file.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	if err := cmd.Wait(); err != nil {
		return nil, err
	}

	ret, err := ioutil.ReadFile(file.Name())
	if err != nil {
		return nil, err
	}

	return ret, err
}
