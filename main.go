package main

import (
	"os"

	"strconv"

	"github.com/dghubble/sling"
	"github.com/dnsimple/dnsimple-go/dnsimple"
	log "github.com/sirupsen/logrus"
	"gopkg.in/urfave/cli.v2"
)

func init() {
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)
}

type ipAddress struct {
	Address string `json:"address,omitempty"`
}

type DNSimpleAccess struct {
	Client    *dnsimple.Client
	AccountID string
	IPAddress *ipAddress
}

func (dns DNSimpleAccess) Init(oauthToken string) (DNSimpleAccess, error) {
	dns.Client = dnsimple.NewClient(dnsimple.NewOauthTokenCredentials(oauthToken))
	dns.Client.UserAgent = "go-dnsimple"
	dns.Client.BaseURL = "https://api.sandbox.dnsimple.com"

	whoami, err := dns.Client.Identity.Whoami()
	if err != nil {
		return dns, err
	}

	if whoami.Data.Account != nil {
		dns.AccountID = strconv.Itoa(whoami.Data.Account.ID)
	} else {
		dns.AccountID = strconv.Itoa(whoami.Data.User.ID)
	}

	log.WithFields(log.Fields{
		"AccountID": dns.AccountID,
	}).Info("Initialized Client.")
	return dns, nil
}

func (dns DNSimpleAccess) ListDomains() ([]dnsimple.Domain, error) {

	domains, err := dns.Client.Domains.ListDomains(dns.AccountID, nil)

	if err != nil {
		return nil, err
	}

	log.WithFields(log.Fields{
		"AccountID": dns.AccountID,
		"domains":   domains,
	}).Debug("Domains loaded.")

	return domains.Data, nil
}

func (dns DNSimpleAccess) GetDomain(name string) (*dnsimple.Domain, error) {
	domain, err := dns.Client.Domains.GetDomain(dns.AccountID, name)
	return domain.Data, err
}

func (dns DNSimpleAccess) CreateRecord(name string, domain string, addrType string, content string) (*dnsimple.ZoneRecord, error) {
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
		log.WithField("request", req).WithError(err).Error("Unable to determine current V4 IP Address")
		return nil, err
	}
	log.WithField("dynamicIP", *address).Debug("Dynamic IP deteremined.")
	return address, nil
}

func main() {

	app := &cli.App{
		EnableShellCompletion: true,
		Name:  "go-dnsimple",
		Usage: "manipulate DNSimple domains!",
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "debug",
				Aliases: []string{"D"},
				Usage:   "Enable debug output",
			},
			&cli.BoolFlag{
				Name:    "dynamicIP",
				Aliases: []string{"d"},
				Usage:   "Lookup current IP dynamically",
			},
			&cli.StringFlag{
				Name:    "token",
				Aliases: []string{"t"},
				Usage:   "DNSimple API OAuth Token",
				EnvVars: []string{"DNSIMPLE_AUTH_TOKEN"},
			},
			&cli.BoolFlag{
				Name:        "ipv4",
				Aliases:     []string{"4"},
				Usage:       "Dynamic lookup with IP V4 address",
				Value:       true,
				DefaultText: "true",
			},
			&cli.BoolFlag{
				Name:        "ipv6",
				Aliases:     []string{"6"},
				Usage:       "Dynamic lookup with IP V6 address",
				Value:       false,
				DefaultText: "false",
			},
		},
		Commands: []*cli.Command{
			{
				Name:  "record",
				Usage: "domain record operations",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "domain",
						Aliases: []string{"d"},
						Usage:   "domain name to operate on",
					},
					&cli.BoolFlag{
						Name:    "dynamicLookup",
						Usage:   "lookup current Internet IP dynamically",
						Aliases: []string{"dynamic", "dyn"},
					},
				},
				Subcommands: []*cli.Command{
					{
						Name:    "list",
						Aliases: []string{"ls"},
					},
					{
						Name:    "create",
						Aliases: []string{"add", "cr"},
						Usage:   "create a new record",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:        "type",
								Aliases:     []string{"t"},
								DefaultText: "A",
								Usage:       "Record type to create: A, AAAA",
							},
						},
						Action: func(context *cli.Context) error {
							dns, err := DNSimpleAccess{}.Init(context.String("token"))
							if err != nil {
								return cli.Exit("Cannot communicate with DNSimple API", 1)
							}
							if context.Bool("dynamicIP") {
								dns.IPAddress, err = LookupCurrentIP(context.Bool("ipv6"))
								if err != nil {
									return cli.Exit("Cannot determine external IP", 1)
								}
							}
							domain, err := dns.GetDomain(context.String("domain"))
							if err != nil {
								log.WithField("accountID", dns.AccountID).WithError(err).Error("Domain not found.")
								return cli.Exit("Domain not found at DNSimple for current account.", 1)
							}

							resp, err := dns.CreateRecord(context.Args().First(), domain.Name, context.String("type"), dns.IPAddress.Address)
							if err != nil {
								log.WithFields(log.Fields{
									"accountID": dns.AccountID,
									"domain":    domain.Name,
									"name":      context.Args().First(),
									"type":      context.String("type"),
								}).WithError(err).Error("Cannot create record")
								return cli.Exit("Domain not found at DNSimple for current account.", 1)
							}

							log.WithField("record", resp)
							return nil
						},
					},
					{
						Name:    "delete",
						Aliases: []string{"del"},
					},
					{
						Name: "update",
					},
					{
						Name: "upsert",
					},
				},
			},
		},
	}

	app.Run(os.Args)
}
