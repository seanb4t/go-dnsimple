package main

import (
	"github.com/dghubble/sling"
	log "github.com/sirupsen/logrus"
)

func LookupCurrentIP(ipv6 bool) (*ipAddress, error) {
	address := new(ipAddress)
	var url string
	if ipv6 {
		url = "https://v6.ident.me"
	} else {
		url = "https://v4.ident.me"
	}
	req, err := sling.New().Base(url).Path(".json").ReceiveSuccess(address)
	if err != nil {
		log.WithField("request", req).WithError(err).Error("Unable to determine current V4 IP address")
		return nil, err
	}
	log.WithField("dynamicIP", *address).Debug("Dynamic IP deteremined.")
	return address, nil
}
