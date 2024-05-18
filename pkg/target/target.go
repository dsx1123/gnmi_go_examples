package target

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"regexp"

	"github.com/dsx1123/gnmi_go/pkg/action"
	"github.com/dsx1123/gnmi_go/pkg/config"
	"github.com/dsx1123/gnmi_go/pkg/gnmi_nxos"
	"github.com/dsx1123/gnmi_go/pkg/utils"
	"github.com/openconfig/gnmi/proto/gnmi"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type Target struct {
	ctx       context.Context
	Config    *config.Config
	GRPCOpts  []grpc.DialOption
	TLSConfig *tls.Config
	UsrCert   *tls.Certificate
	Client    *gnmi_nxos.NXOSGNMIClient
}

type Update struct {
	Path     string
	JSONFile string
}

func (t *Target) stream(subscribeClient gnmi.GNMI_SubscribeClient, act action.OptValue) error {
	for {
		switch act {
		case action.Subscribe:
			if closed, err := t.receiveNotifications(subscribeClient); err != nil {
				return err
			} else if closed {
				return nil
			}
		case action.EDA:
			if closed, err := t.waitForCondition(subscribeClient); err != nil {
				return err
			} else if closed {
				return nil
			}
		}
	}
}

func (t *Target) waitForCondition(subscribeClient gnmi.GNMI_SubscribeClient) (bool, error) {
	for {
		res, err := subscribeClient.Recv()
		if err == io.EOF {
			return true, nil
		}
		if err != nil {
			return false, err
		}
		switch res.Response.(type) {
		case *gnmi.SubscribeResponse_Update:
			var changeMsg = regexp.MustCompile(`Configured from vty by (?P<username>[a-z]+) on`)
			for _, update := range res.GetUpdate().Update {
				path := utils.GetXPath(update.GetPath())
				if path != "/text" {
					continue
				}
				msg := update.GetVal().GetStringVal()
				if !changeMsg.MatchString(msg) {
					continue
				}
				matches := changeMsg.FindStringSubmatch(msg)
				nameIndex := changeMsg.SubexpIndex("username")
				username := matches[nameIndex]
				if username == "admin" || username == "root" {
					continue
				}
				log.Printf("System configuration is changed from cli by %s, this is wrong!", username)

				response, err := t.RelaceSystem(t.ctx)
				if err != nil {
					return false, err
				}
				for _, n := range response.GetResponse() {
					xpath := utils.GetXPath(n.GetPath())
					log.Printf("Set path %s successfully!", xpath)
				}
			}
		}
	}
}

func (t *Target) receiveNotifications(subscribeClient gnmi.GNMI_SubscribeClient) (bool, error) {
	for {
		res, err := subscribeClient.Recv()
		if err == io.EOF {
			return true, nil
		}
		if err != nil {
			return false, err
		}
		switch res.Response.(type) {
		case *gnmi.SubscribeResponse_SyncResponse:
			log.Printf("SyncResponse received: %v", res)

		case *gnmi.SubscribeResponse_Update:
			output := new(utils.Output)
			for _, update := range res.GetUpdate().Update {
				xpath := utils.GetXPath(update.GetPath())
				val := update.GetVal().GetStringVal()
				output.Path = xpath
				output.Value = make(map[string]interface{})
				output.Value["val"] = val
				log.Println(val)
				mapJson, _ := json.Marshal(output)
				log.Printf("Get Response: \n %s \n", utils.PrettyJSON(&mapJson))
			}
		default:
			return false, errors.New("unexpected response type")
		}
	}
}

func (t *Target) RelaceSystem(ctx context.Context) (*gnmi.SetResponse, error) {
	file := "./config/system.json"
	path := "/System"
	opt := action.Replace

	jsonConfig, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("Read file %s faild", file)
	}

	val := &gnmi.TypedValue{
		Value: &gnmi.TypedValue_JsonVal{
			JsonVal: jsonConfig,
		},
	}
	log.Println("Replace System configuration with golden config!")
	response, err := t.Client.Set(t.appendMetadata(ctx), path, val, opt)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (t *Target) GetCapbilites(ctx context.Context) (*gnmi.CapabilityResponse, error) {

	response, err := t.Client.GetCapbilites(t.appendMetadata(ctx))
	if err != nil {
		return nil, err
	}
	return response, nil

}

func (t *Target) Get(ctx context.Context, path string, encoding string) (*gnmi.GetResponse, error) {
	gEncoding := gnmi.Encoding(gnmi.Encoding_value[encoding])

	response, err := t.Client.Get(t.appendMetadata(ctx), path, gEncoding)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (t *Target) Set(
	ctx context.Context,
	path string,
	value *[]byte,
	opt action.SubOptValue,
) (*gnmi.SetResponse, error) {
	var val *gnmi.TypedValue
	if opt == action.Merge || opt == action.Replace {
		val = &gnmi.TypedValue{
			Value: &gnmi.TypedValue_JsonVal{
				JsonVal: *value,
			},
		}
	}
	response, err := t.Client.Set(t.appendMetadata(ctx), path, val, opt)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (t *Target) Subscribe(
	ctx context.Context,
	subscriptions *config.Subscriptions,
	encoding string,
) error {
	gEncoding := gnmi.Encoding(gnmi.Encoding_value[encoding])
	gnmiSubs, err := GetSubscriptionList(subscriptions)
	if err != nil {
		return err
	}
	err = t.Client.Subscribe(t.appendMetadata(ctx), gnmiSubs, gEncoding)
	if err != nil {
		return err
	}
	if err := t.stream(t.Client.GNMISubClient, action.Subscribe); err != nil {
		return fmt.Errorf("Error using STREAM mode: %v", err)
	}
	return nil
}

func (t *Target) EDADemo(ctx context.Context) error {
	t.ctx = ctx
	subscriptions := &config.Subscriptions{
		config.Subscription{
			Origin: "Syslog-oper",
			Path:   "/syslog/message",
			Mode:   "ON_CHANGE",
		},
	}
	gEncoding := gnmi.Encoding(gnmi.Encoding_value["PROTO"])
	gnmiSubs, err := GetSubscriptionList(subscriptions)
	if err != nil {
		return err
	}
	err = t.Client.Subscribe(t.appendMetadata(ctx), gnmiSubs, gEncoding)
	if err != nil {
		return err
	}
	if err := t.stream(t.Client.GNMISubClient, action.EDA); err != nil {
		return fmt.Errorf("Error using STREAM mode: %v", err)
	}
	return nil
}

func (t *Target) appendMetadata(ctx context.Context) context.Context {
	// append username/password to the metadata if not using certificate
	ctx = metadata.AppendToOutgoingContext(ctx, "username", t.Config.Username)
	if t.Config.TLSCert == "" || t.Config.TLSKey == "" {
		ctx = metadata.AppendToOutgoingContext(ctx, "password", t.Config.Password)
	}
	return ctx
}

func GetSubscriptionList(s *config.Subscriptions) ([]*gnmi.Subscription, error) {
	var subscriptions []*gnmi.Subscription
	var subMode gnmi.SubscriptionMode
	var subscription *gnmi.Subscription
	for _, sub := range *s {
		subMode = gnmi.SubscriptionMode(gnmi.SubscriptionMode_value[sub.Mode])
		_, pElem, err := utils.ParsePath(sub.Path)
		pElem.Origin = sub.Origin
		if err != nil {
			return subscriptions, err
		}
		switch subMode {
		case gnmi.SubscriptionMode_SAMPLE:
			subscription = &gnmi.Subscription{
				Path:           pElem,
				Mode:           subMode,
				SampleInterval: uint64(sub.Interval * 1000000000),
			}
		case gnmi.SubscriptionMode_ON_CHANGE:
			subscription = &gnmi.Subscription{
				Path: pElem,
				Mode: subMode,
			}
		}
		subscriptions = append(subscriptions, subscription)
	}
	return subscriptions, nil
}
