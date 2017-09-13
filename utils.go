package main

import (
	"github.com/go-resty/resty"
	log "github.com/sirupsen/logrus"
)

func LookupCurrentIP(ipv6 bool) (*RecordData, error) {
	address := new(RecordData)
	var url string
	if ipv6 {
		url = "https://ipv6bot.whatismyipaddress.com"
	} else {
		url = "https://ipv4bot.whatismyipaddress.com"
	}
	resp, err := resty.R().SetHeader("Content-Type", "text/plain").Get(url)
	if err != nil {
		log.WithField("resp", resp).WithError(err).Error("Unable to determine current V4 IP address")
		return nil, err
	}
	address.Value = resp.String()
	log.WithField("dynamicIP", *address).Debug("Dynamic IP deteremined.")
	return address, nil
}
