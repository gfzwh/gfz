package config

import (
	"encoding/xml"
	"io"
	"os"
)

type Gfz struct {
	XMLName xml.Name `xml:"gfz"`
	Log     log      `xml:"log"`
	Client  client   `xml:"client"`
	Server  server   `xml:"server"`
}

type log struct {
	XMLName xml.Name `xml:"log"`
	Level   string   `xml:"level"`
	LogFile string   `xml:"log_file"`
}

type client struct {
	XMLName xml.Name `xml:"client"`
	S2sName string   `xml:"s2sname"`
	S2sKey  string   `xml:"s2skey"`
	Url     string   `xml:"url"`
	Zone    string   `xml:"zone"`
	Env     string   `xml:"env"`
	Host    string   `xml:"host"`
}

type server struct {
	XMLName xml.Name `xml:"server"`
	S2sName string   `xml:"s2sname"`
	S2sKey  string   `xml:"s2skey"`
	Region  string   `xml:"region"`
	Zone    string   `xml:"zone"`
	Env     string   `xml:"env"`
	Host    string   `xml:"host"`
	Nodes   nodes    `xml:"nodes"`
	Url     string   `xml:"url"`
}

type nodes struct {
	XMLName xml.Name `xml:"nodes"`
	Node    []string `xml:"node"`
}

func XmlParse(f string) (gfzf *Gfz, err error) {
	xmlFile, err := os.Open(f)
	if err != nil {
		return
	}
	defer xmlFile.Close()

	byteValue, err := io.ReadAll(xmlFile)
	if nil != err {
		return
	}

	var gfz Gfz
	err = xml.Unmarshal(byteValue, &gfz)
	if nil != err {
		return
	}

	gfzf = &gfz
	return
}
