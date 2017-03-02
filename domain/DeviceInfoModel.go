package domain

type DeviceInfoModel struct {
	ID           string
	Hostname     string
	Architecture string
	Platform     string
	IsConnected  bool
	Tags         []string
}
