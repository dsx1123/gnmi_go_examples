package config

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	"github.com/mcuadros/go-defaults"
)

type Config struct {
	CfgFile            string        `yaml:"-"`
	Address            string        `yaml:"address"`
	Username           string        `yaml:"username"             default:"admin"`
	Password           string        `yaml:"password"`
	InsecureSkipVerify bool          `yaml:"insecure_skip_verify" default:"true"`
	TLSCA              string        `yaml:"tls_ca"`
	TLSCert            string        `yaml:"tls_cert"`
	TLSKey             string        `yaml:"tls_key"`
	Encoding           string        `yaml:"encoding"`
	GetPath            string        `yaml:"get"`
	SetMerge           Update        `yaml:"set"`
	SetReplace         Update        `yaml:"replace"`
	DeletePath         string        `yaml:"delete"`
	Subscriptions      Subscriptions `yaml:"subscriptions"`
}

type Subscription struct {
	Origin   string `yaml:"origin"`
	Path     string `yaml:"path"`
	Mode     string `yaml:"mode"     default:"sample"`
	Interval int    `yaml:"interval"                  defualt:"30"`
}

type Subscriptions []Subscription

type Update struct {
	Path     string `yaml:"path"`
	JSONFile string `yaml:"file"`
}

func New() *Config {
	config := new(Config)
	defaults.SetDefaults(config)
	return config
}

func NewTLSConfig(certFile string, skipTLSVerify bool) (*tls.Config, error) {
	// read content of config.TLSCA and create TLS config
	// if config.TLSCA is not empty
	if certFile == "" {
		return nil, fmt.Errorf("tls_ca is empty!")
	}

	// Read the TLS CA file
	caCert, err := os.ReadFile(certFile)
	if err != nil {
		return nil, err
	}

	caPool := x509.NewCertPool()
	if !caPool.AppendCertsFromPEM(caCert) {
		return nil, err
	}

	tlsConfig := &tls.Config{
		InsecureSkipVerify: skipTLSVerify,
	}
	tlsConfig.RootCAs = caPool
	return tlsConfig, nil

}

func NewX509Cert(tlsCert string, tlsKey string) (*tls.Certificate, error) {
	var usrCert tls.Certificate

	if tlsCert != "" && tlsKey != "" {
		certBytes, err := os.ReadFile(tlsCert)
		if err != nil {
			return nil, fmt.Errorf("Read user cert file failed: %s", err)
		}
		keyBytes, err := os.ReadFile(tlsKey)
		if err != nil {
			return nil, fmt.Errorf("Read user key file failed: %s", err)
		}
		usrCert, err = tls.X509KeyPair(certBytes, keyBytes)
		if err != nil {
			return nil, fmt.Errorf("Failed to load user cert/key pair: %s", err)
		}
	}
	return &usrCert, nil
}
