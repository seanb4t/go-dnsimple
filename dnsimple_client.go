package main

import (
	"fmt"
	"github.com/dnsimple/dnsimple-go/dnsimple"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"strconv"
)

type DNSimpleAPI struct {
	Client    *dnsimple.Client
	AccountID string
}

func NewDNSimpleAPI(oauthToken string) (*DNSimpleAPI, error) {
	api := DNSimpleAPI{
		Client: dnsimple.NewClient(dnsimple.NewOauthTokenCredentials(oauthToken)),
	}
	api.Client.UserAgent = "go-dnsimple"
	//api.Client.BaseURL = "https://api.sandbox.dnsimple.com"

	accountList, err := api.Client.Accounts.ListAccounts(&dnsimple.ListOptions{})
	if err != nil {
		log.WithError(err).Fatal("Cannot get DNSimple account info")
		return nil, errors.Wrap(err, "Unable to authenticate and retrieve DNSimple account info")
	}
	log.WithField("accountList", accountList.Data).Debug("Retrieved account list")

	whoami, err := api.Client.Identity.Whoami()
	if err != nil {
		log.WithError(err).Fatal("Cannot get DNSimple whoami info")
		return nil, errors.Wrap(err, "Unable to authenticate and retrieve DNSimple whoami info")
	} else {

		if whoami.Data.Account != nil {
			log.WithField("accountId", fmt.Sprint(whoami.Data.Account.ID)).Info("Account Id")
			api.AccountID = strconv.FormatInt(whoami.Data.Account.ID, 10)
		} else {
			log.WithField("userId", fmt.Sprint(whoami.Data.User.ID)).Info("User Id")
			api.AccountID = strconv.FormatInt(whoami.Data.User.ID, 10)
		}
	}

	log.WithField("api.AccountID", api.AccountID).Info("API Account Id chosen")
	log.WithFields(log.Fields{
		"AccountID": api.AccountID,
	}).Info("Initialized Client.")
	return &api, nil
}

func (dns *DNSimpleAPI) GetZone(name string) (zone *dnsimple.Zone, exists bool, err error) {
	resp, err := dns.Client.Zones.GetZone(dns.AccountID, name)
	if err != nil {
		return nil, false, err
	}
	if resp.Data != nil {
		return resp.Data, true, nil
	}
	return nil, false, nil
}

func (dns *DNSimpleAPI) GetZoneRecord(name string, domain string, recordType string) (zoneRecord *dnsimple.ZoneRecord, exists bool, err error) {
	resp, err := dns.Client.Zones.ListRecords(dns.AccountID, domain, &dnsimple.ZoneRecordListOptions{
		Name: name,
	})
	if err != nil {
		return nil, false, err
	}
	if len(resp.Data) > 0 {
		record := resp.Data[0]
		if record.Type == recordType {
			return &record, true, nil
		}
	}
	return nil, false, nil
}

func (dns *DNSimpleAPI) CreateRecord(name string, domain string, addrType string, content string) (zoneRecord *dnsimple.ZoneRecord, err error) {
	resp, err := dns.Client.Zones.CreateRecord(dns.AccountID, domain, dnsimple.ZoneRecord{
		Name:    name,
		Type:    addrType,
		Content: content,
		TTL:     300, //TODO: make option
	})
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}

func (dns *DNSimpleAPI) UpsertRecord(name string, domain string, addrType string, content string) (zoneRecord *dnsimple.ZoneRecord, err error) {

	record := dnsimple.ZoneRecord{
		Name:    name,
		Type:    addrType,
		Content: content,
		TTL:     300, //TODO: make option
	}

	zoneRecord, exists, _ := dns.GetZoneRecord(name, domain, addrType)

	if exists && record.Type == "AAAA" {
		resp, err := dns.Client.Zones.DeleteRecord(dns.AccountID, domain, zoneRecord.ID)
		if err != nil {
			return nil, err
		}
		return resp.Data, nil
	}

	if exists {
		resp, err := dns.Client.Zones.UpdateRecord(dns.AccountID, domain, zoneRecord.ID, record)
		if err != nil {
			return nil, err
		}
		return resp.Data, nil
	} else {
		resp, err := dns.Client.Zones.CreateRecord(dns.AccountID, domain, record)
		if err != nil {
			return nil, err
		}
		return resp.Data, nil
	}
}

func (dns *DNSimpleAPI) DeleteRecord(name string, domain string, addrType string, content string) (zoneRecord *dnsimple.ZoneRecord, err error) {
	zoneRecord, exists, _ := dns.GetZoneRecord(name, domain, addrType)

	if exists {
		deleteResponse, err := dns.Client.Zones.DeleteRecord(dns.AccountID, domain, zoneRecord.ID)
		if err != nil {
			return nil, err
		}
		return deleteResponse.Data, nil
	}
	return nil, nil
}
