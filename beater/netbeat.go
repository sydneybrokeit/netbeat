package beater

import (
	"fmt"
	"github.com/soniah/gosnmp"
	"time"

	"github.com/elastic/beats/libbeat/beat"
	"github.com/elastic/beats/libbeat/common"
	"github.com/elastic/beats/libbeat/logp"

	"github.com/hmschreck/netbeat/config"
)

// Netbeat configuration.
type Netbeat struct {
	done   chan struct{}
	config config.Config
	client beat.Client
}

// New creates an instance of netbeat.
func New(b *beat.Beat, cfg *common.Config) (beat.Beater, error) {
	c := config.DefaultConfig
	if err := cfg.Unpack(&c); err != nil {
		return nil, fmt.Errorf("Error reading config file: %v", err)
	}

	bt := &Netbeat{
		done:   make(chan struct{}),
		config: c,
	}
	return bt, nil
}

// Run starts netbeat.
func (bt *Netbeat) Run(b *beat.Beat) error {
	logp.Info("netbeat is running! Hit CTRL-C to stop it.")

	var err error
	bt.client, err = b.Publisher.Connect()
	if err != nil {
		return err
	}

	ticker := time.NewTicker(bt.config.Period)
	for {
		select {
		case <-bt.done:
			return nil
		case <-ticker.C:
			for _, configuration := range bt.config.Sets {
				for _, host := range configuration.Hosts {
					event := beat.Event{
						Timestamp: time.Now(),
						Fields: common.MapStr{"snmp.host": host, "type": b.Info.Name },
					}
					gosnmp.Default.Target = host
					gosnmp.Default.Port = configuration.Port
					gosnmp.Default.Community = configuration.Community
					version := gosnmp.Version2c
					switch configuration.Version {
					case "1":
						version = gosnmp.Version1
					case "2c":
						version = gosnmp.Version2c
					case "3":
						version = gosnmp.Version3
					default:
						logp.Err("Wrong SNMP version %s, defaulting to 2c", configuration.Version)
					}
					gosnmp.Default.Version = version
					if version == gosnmp.Version3 {
						gosnmp.Default.SecurityModel = gosnmp.UserSecurityModel
						gosnmp.Default.SecurityParameters = &gosnmp.UsmSecurityParameters{
							UserName:                 configuration.User,
							AuthenticationPassphrase: configuration.AuthPassword,
							PrivacyPassphrase:        configuration.PrivPassword,
							AuthenticationProtocol:   gosnmp.SHA,
							PrivacyProtocol:          gosnmp.DES,
						}
					}
					m := make(map[string]string)
					var oids []string
					for _, v := range configuration.OIDs {
						m[v["oid"]] = v["name"]
						oids = append(oids, v["oid"])
					}
					err := gosnmp.Default.Connect()
					if err != nil {
						logp.Critical("Can't connect to %s: %v", host, err.Error())
						return fmt.Errorf("Can't connec to %s", host)
					}
					defer gosnmp.Default.Conn.Close()
					r, err := gosnmp.Default.Get(oids)
					if err != nil {
						logp.Err("Can't get OIDs for %v: %v", host, err.Error())
					} else {
						for _, v := range r.Variables {
							var value interface{}
							k := m[v.Name]
							if k == "" {
								k = v.Name
							}
							switch v.Type {
							case gosnmp.OctetString:
								value = string(v.Value.([]byte))
							default:
								value = gosnmp.ToBigInt(v.Value)
							}
							event.Fields.Put(k, value)
						}
					}
					bt.client.Publish(event)
				}
			}
		}
	}
}

// Stop stops netbeat.
func (bt *Netbeat) Stop() {
	bt.client.Close()
	close(bt.done)
}
