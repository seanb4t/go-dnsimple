package main

import (
	"os"

	log "github.com/sirupsen/logrus"
	"gopkg.in/urfave/cli.v2"
)

func init() {
	log.SetOutput(os.Stdout)
	//log.SetLevel(log.DebugLevel)
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
				Name:    "jsonOutput",
				Aliases: []string{"J"},
				Usage:   "Enable json output",
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
		Before: func(c *cli.Context) error {
			if c.Bool("debug") {
				log.SetLevel(log.DebugLevel)
			}

			if c.Bool("jsonOutput") {
				log.SetFormatter(&log.JSONFormatter{})
			}
			return nil
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
							&cli.StringFlag{
								Name:  "data",
								Usage: "Record data",
							},
						},
						Action: func(context *cli.Context) error {
							dns, err := NewDNSimpleAPI(context.String("token"))
							if err != nil {
								log.WithError(err).Fatal("Cannot communicate with DNSimple API")
								return cli.Exit("Cannot communicate with DNSimple API", 1)
							}

							domainName := context.String("domain")
							recordName := context.Args().First()
							recordType := context.String("type")
							recordData := context.String("data")

							crudOp, err := NewOperationData(dns, recordData, recordName, domainName, recordType, context.Bool("ipv6"))
							if err != nil {
								log.WithError(err).Fatal("Could not create command operation data.")
								return cli.Exit("Cannot create record", 1)
							}

							err = crudOp.Create()
							if err != nil {
								log.WithError(err).Fatal("Could not create record")
								return cli.Exit("Cannot create record", 1)
							}
							return nil
						},
					},
					{
						Name:        "delete",
						Aliases:     []string{"del"},
						Description: "delete record data",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:        "type",
								Aliases:     []string{"t"},
								DefaultText: "A",
								Usage:       "Record type to delete: A, AAAA",
							},
							&cli.StringFlag{
								Name:  "data",
								Usage: "Record data",
							},
						},
						Action: func(context *cli.Context) error {
							dns, err := NewDNSimpleAPI(context.String("token"))
							if err != nil {
								return cli.Exit("Cannot communicate with DNSimple API", 1)
							}

							domainName := context.String("domain")
							recordName := context.Args().First()
							recordType := context.String("type")
							recordData := context.String("data")

							crudOp, err := NewOperationData(dns, recordData, recordName, domainName, recordType, context.Bool("ipv6"))
							if err != nil {
								log.WithError(err).Fatal("Could not create command operation data.")
								return cli.Exit("Cannot delete record", 1)
							}

							err = crudOp.Delete()
							if err != nil {
								return cli.Exit("Cannot delete record", 1)
							}
							return nil
						},
					},
					{
						Name:        "upsert",
						Aliases:     []string{"up", "update"},
						Description: "upsert record data",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:        "type",
								Aliases:     []string{"t"},
								DefaultText: "A",
								Usage:       "Record type to upsert: A, AAAA",
							},
							&cli.StringFlag{
								Name:  "data",
								Usage: "Record data",
							},
						},
						Action: func(context *cli.Context) error {
							dns, err := NewDNSimpleAPI(context.String("token"))
							if err != nil {
								return cli.Exit("Cannot communicate with DNSimple API", 1)
							}

							domainName := context.String("domain")
							recordName := context.Args().First()
							recordType := context.String("type")
							recordData := context.String("data")

							crudOp, err := NewOperationData(dns, recordData, recordName, domainName, recordType, context.Bool("ipv6"))
							if err != nil {
								log.WithError(err).Fatal("Could not create command operation data.")
								return cli.Exit("Cannot upsert record", 1)
							}

							err = crudOp.Upsert()
							if err != nil {
								return cli.Exit("Cannot upsert record", 1)
							}
							return nil
						},
					},
				},
			},
		},
	}

	app.Run(os.Args)
}
