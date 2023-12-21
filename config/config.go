package config

import (
	"strings"
)

type iconfig interface {
	Get(args ...string) aReader
}

var iconf iconfig

func Init(f string) {
	if strings.HasSuffix(f, ".xml") {
		iconf = GetXml(f)
	}
}

func Get(args ...string) aReader {
	if nil == iconf {
		return aReader{}
	}

	arg := make([]string, 0)
	arg = append(arg, "gfz")
	arg = append(arg, args...)
	return iconf.Get(arg...)
}
