package install

type Step interface {
	Install(i *Installer, args Args) error
	Uninstall(i *Installer, args Args) error
	Upgrade(i *Installer, args Args) error
	Repair(i *Installer, args Args) error
}
