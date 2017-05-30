package install

import (
	"io"
	"os"
)

type Installer struct {
	Title   string
	Version string
	Steps   []Step
}

func (t *Installer) AddStep(s Step) {
	t.Steps = append(t.Steps, s)
}

func (t *Installer) RunInstall(args Args) error {
	for _, step := range t.Steps {
		err := step.Install(t, args)

		if err != nil {
			return err
		}
	}

	return nil
}

func (t *Installer) RunUninstall(args Args) error {
	for _, step := range t.Steps {
		err := step.Uninstall(t, args)

		if err != nil {
			return err
		}
	}

	return nil
}

func (t *Installer) RunUpgrade(args Args) error {
	for _, step := range t.Steps {
		err := step.Upgrade(t, args)

		if err != nil {
			return err
		}
	}

	return nil
}

func (t *Installer) RunRepair(args Args) error {
	for _, step := range t.Steps {
		err := step.Repair(t, args)

		if err != nil {
			return err
		}
	}

	return nil
}

func (t *Installer) Copy(dst, src string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	cerr := out.Close()
	if err != nil {
		return err
	}
	return cerr
}

func (t *Installer) Exists(path string) bool {
	file, err := os.Open(path)
	defer file.Close()

	if os.IsNotExist(err) {
		return false
	}

	return true
}

func (t *Installer) MakePaths(paths []string) error {
	for _, path := range paths {
		if err := os.MkdirAll(path, 0700); err != nil {
			return err
		}
	}

	return nil
}
