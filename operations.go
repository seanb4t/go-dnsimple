package main

import (
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type RecordData struct {
	Value string
}

type CrudOperationData struct {
	dns *DNSimpleAPI
	RecordData
	Domain       string
	RecordName   string
	RecordType   string
	DomainExists bool
	RecordExists bool
	DeleteRecord bool
}

func NewOperationData(dns *DNSimpleAPI, recordData string, recordName string, domain string, recordType string, ipv6 bool) (*CrudOperationData, error) {
	data := CrudOperationData{
		dns:        dns,
		RecordName: recordName,
		Domain:     domain,
		RecordType: recordType,
	}

	if len(recordData) == 0 {
		log.Debug("no record data provided, dynamically looking up current IP")
		resp, err := LookupCurrentIP(ipv6)
		if err != nil {
			return nil, errors.Wrap(err, "Unable to lookup current IP")
		}
		data.Value = resp.Value
		log.Infof("Current dynamic ip is: %s", data.Value)
	} else {
		data.Value = recordData
	}

	resp, _, err := dns.GetZone(domain)
	if err != nil {
		data.DomainExists = false
		return &data, errors.Wrapf(err, "Unable to find domain: %s", domain)
	}
	domain = resp.Name
	data.DomainExists = true

	_, exists, err := dns.GetZoneRecord(recordName, domain, recordType)
	if err != nil {
		data.RecordExists = exists
		return &data, errors.Wrapf(err, "Unable to find record: %s in domain %s of type %s", recordName, domain, recordType)
	}

	return &data, nil
}

func (data *CrudOperationData) Create() (err error) {
	_, err = data.dns.CreateRecord(data.RecordName, data.Domain, data.RecordType, data.Value)
	if err != nil {
		log.WithFields(log.Fields{
			"accountID":  data.dns.AccountID,
			"domain":     data.Domain,
			"recordName": data.RecordName,
			"type":       data.RecordType,
			"recordData": data.Value,
		}).WithError(err).Error("Cannot create record")
		return errors.Wrapf(err, "Cannot create record %s with data: %s in domain: %s", data.RecordName, data.Value, data.Domain)
	}
	return nil
}

func (data *CrudOperationData) Delete() (err error) {
	_, err = data.dns.DeleteRecord(data.RecordName, data.Domain, data.RecordType, data.Value)
	if err != nil {
		log.WithFields(log.Fields{
			"accountID":  data.dns.AccountID,
			"domain":     data.Domain,
			"recordName": data.RecordName,
			"type":       data.RecordType,
			"recordData": data.Value,
		}).WithError(err).Error("Cannot create record")
		return errors.Wrapf(err, "Cannot create record %s with data: %s in domain: %s", data.RecordName, data.Value, data.Domain)
	}
	return nil
}

func (data *CrudOperationData) Upsert() (err error) {
	resp, err := data.dns.UpsertRecord(data.RecordName, data.Domain, data.RecordType, data.Value)
	if err != nil {
		log.WithFields(log.Fields{
			"accountID":  data.dns.AccountID,
			"domain":     data.Domain,
			"recordName": data.RecordName,
			"type":       data.RecordType,
			"recordData": data.Value,
		}).WithError(err).Error("Cannot update/create record")
		return errors.Wrapf(err, "Cannot update/create record %s with data: %s in domain: %s", data.RecordName, data.Value, data.Domain)
	}
	log.WithField("response", resp).Debug("Record upsert complete.")
	log.WithFields(log.Fields{
		"recordName":   resp.Name,
		"recordDomain": resp.ZoneID,
		"recordData":   resp.Content,
		"recordType":   resp.Type,
		"recordTTL":    resp.TTL,
	}).Info("Upsert record complete")
	return nil
}
