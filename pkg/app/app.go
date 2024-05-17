package app

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/dsx1123/gnmi_go/pkg/action"
	"github.com/dsx1123/gnmi_go/pkg/config"
	"github.com/dsx1123/gnmi_go/pkg/gnmi_nxos"
	"github.com/dsx1123/gnmi_go/pkg/target"
	"github.com/dsx1123/gnmi_go/pkg/utils"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/encoding/protojson"
	"gopkg.in/yaml.v2"
)

type App struct {
	ctx        context.Context
	cfn        context.CancelFunc
	RootCmd    *cobra.Command
	Config     *config.Config
	Target     *target.Target
	Wg         *sync.WaitGroup
	ActionChan chan *action.Action
	errChan    chan error
}

func New() *App {
	ctx, cancel := context.WithCancel(context.Background())
	a := App{
		ctx:        ctx,
		cfn:        cancel,
		RootCmd:    new(cobra.Command),
		Config:     config.New(),
		Target:     new(target.Target),
		Wg:         new(sync.WaitGroup),
		ActionChan: make(chan *action.Action),
		errChan:    make(chan error),
	}
	return &a
}

func (a *App) InitFlags() {
	a.RootCmd.ResetFlags()
	a.RootCmd.PersistentFlags().
		StringVar(&a.Config.CfgFile, "config", "./config.yaml", "target config file")
}

func (a *App) PreRunE(cmd *cobra.Command, args []string) error {
	yamlFile, err := os.ReadFile(a.Config.CfgFile)
	if err != nil {
		log.Fatalf("Failed to read YAML file: %v", err)
	}

	// Unmarshal the YAML data into the config struct
	err = yaml.Unmarshal(yamlFile, &a.Config)
	if err != nil {
		log.Fatalf("Failed to unmarshal YAML data: %v", err)
	}

	a.Target.Config = a.Config
	if a.Target.TLSConfig, err = config.NewTLSConfig(a.Config.TLSCA, a.Config.InsecureSkipVerify); err != nil {
		log.Fatalf("Failed to create TLS Config : %v", err)
	}

	if a.Target.UsrCert, err = config.NewX509Cert(a.Config.TLSCert, a.Config.TLSKey); err != nil {
		log.Fatalf("Failed to create User Cert: %v", err)
	}

	if a.Target.UsrCert != nil {
		a.Target.TLSConfig.Certificates = []tls.Certificate{*a.Target.UsrCert}
	}

	var creds credentials.TransportCredentials

	if a.Target.TLSConfig != nil {
		creds = credentials.NewTLS(a.Target.TLSConfig)
	} else {
		creds = insecure.NewCredentials()
	}

	a.Target.GRPCOpts = []grpc.DialOption{
		grpc.WithTransportCredentials(creds),
	}

	client, err := gnmi_nxos.NewNXOSGNMIClient(a.ctx, a.Config.Address, a.Target.GRPCOpts...)
	if err != nil {
		return err
	}
	a.Target.Client = client

	a.Wg.Add(1)
	go a.Worker()

	return nil
}

func (a *App) RunE(cmd *cobra.Command, args []string) error {
	act := new(action.Action)
	use := cmd.Use
	switch use {
	case "cap":
		act.Opt = action.Capabilities
	case "get":
		act.Opt = action.Get
		act.Path = a.Config.GetPath
	case "merge":
		act.Opt = action.Set
		act.SubOpt = action.Merge
		act.Path = a.Config.SetMerge.Path
		jsonConfig, err := os.ReadFile(a.Config.SetMerge.JSONFile)
		if err != nil {
			log.Fatalf("Read file %s faild", a.Config.SetMerge.JSONFile)
		}
		act.Data = &jsonConfig
	case "replace":
		act.Opt = action.Set
		act.SubOpt = action.Replace
		act.Path = a.Config.GetPath
		jsonConfig, err := os.ReadFile(a.Config.SetReplace.JSONFile)
		if err != nil {
			log.Fatalf("Read file %s faild", a.Config.SetMerge.JSONFile)
		}
		act.Data = &jsonConfig
	case "subscribe":
		act.Opt = action.Subscribe
		act.Subscrptions = &a.Config.Subscriptions
	case "eda":
		act.Opt = action.EDA
	}
	a.ActionChan <- act
	a.Wg.Wait()
	return nil
}

func (a *App) gNMICapabilities() error {
	log.Println("Get Capabilities request!")
	cap, err := a.Target.GetCapbilites(a.ctx)
	if err != nil {
		return fmt.Errorf(
			"Failed to get Capabilities from target: %s, %v",
			a.Target.Client.Address,
			err,
		)
	}
	msg, _ := protojson.Marshal(cap)
	output := utils.PrettyJSON(&msg)
	log.Println(output)
	return nil
}

func (a *App) gNMIGet() error {
	response, err := a.Target.Get(
		a.ctx,
		a.Config.GetPath,
		"JSON",
	) // NXOS only supports JSON for GET/SET
	if err != nil {
		return fmt.Errorf(
			"Failed to get path %s from target: %s, %v",
			a.Target.Config.GetPath,
			a.Target.Client.Address,
			err,
		)
	}
	output := new(utils.Output)
	output.Value = make(map[string]interface{})

	for _, n := range response.GetNotification() {
		for _, update := range n.GetUpdate() {
			xpath := utils.GetXPath(update.GetPath())

			val := update.GetVal().GetJsonVal()
			output.Path = xpath
			err := json.Unmarshal(val, &output.Value)
			if err != nil {
				output.Value["val"] = string(val)
			}
			mapJson, _ := json.Marshal(output)
			log.Printf("Get Response: \n %s \n", utils.PrettyJSON(&mapJson))
		}
	}
	return nil
}

func (a *App) gNMISet(act action.SubOptValue) error {
	var path string
	var file string
	switch act {
	case action.Merge:
		path = a.Config.SetMerge.Path
		file = a.Config.SetMerge.JSONFile
	case action.Replace:
		path = a.Config.SetReplace.Path
		file = a.Config.SetReplace.JSONFile
	}
	jsonConfig, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("Read file %s faild", file)
	}
	response, err := a.Target.Set(a.ctx, path, &jsonConfig, act)

	if err != nil {
		return fmt.Errorf(
			"Failed to Set path %s on target: %s, %v",
			path,
			a.Target.Client.Address,
			err,
		)
	}

	for _, n := range response.GetResponse() {
		xpath := utils.GetXPath(n.GetPath())
		log.Printf("Set path %s successfully!", xpath)
	}

	return nil
}

func (a *App) gNMISubscribe() error {
	log.Println("Get Subscribe request!")
	err := a.Target.Subscribe(a.ctx, &a.Config.Subscriptions, a.Config.Encoding)
	if err != nil {
		return err
	}
	return nil
}

func (a *App) gNMIEDA() error {
	// This is a static demo
	// Subscribe to Syslog path, when any user change the configuration from CLI
	// Use gNMI to replace the configuration with Golden Configuration
	log.Println("Get EDA request!")
	err := a.Target.EDADemo(a.ctx)
	if err != nil {
		return err
	}
	return nil
}

func (a *App) Worker() {
	defer a.Wg.Done()
	for {
		select {
		case act := <-a.ActionChan:
			switch act.Opt {
			case action.Capabilities:
				err := a.gNMICapabilities()
				if err != nil {
					log.Fatalln(err)
				}
				return
			case action.Get:
				err := a.gNMIGet()
				if err != nil {
					log.Fatalln(err)
				}
				return
			case action.Set:
				err := a.gNMISet(act.SubOpt)
				if err != nil {
					log.Fatalln(err)
				}
				return
			case action.Subscribe:
				err := a.gNMISubscribe()
				if err != nil {
					log.Fatalln(err)
				}
				return
			case action.EDA:
				err := a.gNMIEDA()
				if err != nil {
					log.Fatalln(err)
				}
				return
			}
		case err := <-a.errChan:
			log.Fatalln(err)
			return
		}
	}
}
