package synchronization

import "strconv"

type Target struct {
	Port     uint16
	Hostname string
	Folder   string
	Username string
}

func (o *Target) buildUrl() string {
	return o.Username + "@" + o.Hostname + ":" + strconv.Itoa(int(o.Port)) + ":" + o.Folder
}
