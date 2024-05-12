package gnmi_nxos

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"

	"github.com/dsx1123/gnmi_go/pkg/action"
	"github.com/dsx1123/gnmi_go/pkg/utils"
	"github.com/openconfig/gnmi/proto/gnmi"
	"google.golang.org/grpc"
)

type NXOSGNMIClient struct {
	Address       string
	opts          *[]grpc.DialOption
	TLSConfig     *tls.Config
	GNMIClient    gnmi.GNMIClient
	GNMISubClient gnmi.GNMI_SubscribeClient
}

type Update struct {
	Path     string
	JSONFile string
}

type Subscription struct {
	Mode     string
	Path     string
	Origin   string
	Interval int // Interval in second
}

func NewNXOSGNMIClient(ctx context.Context, addr string, opts ...grpc.DialOption) (*NXOSGNMIClient, error) {
	c := new(NXOSGNMIClient)
	c.opts = &opts

	dialCtx := context.Context(ctx)
	conn, err := grpc.DialContext(dialCtx, addr, opts...)
	if err != nil {
		log.Fatalf("Failed to connect to the target %s: %v", addr, err)
	}
	c.GNMIClient = gnmi.NewGNMIClient(conn)

	return c, nil
}

func (c *NXOSGNMIClient) GetCapbilites(ctx context.Context) (*gnmi.CapabilityResponse, error) {
	response, err := c.GNMIClient.Capabilities(ctx, &gnmi.CapabilityRequest{})
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (c *NXOSGNMIClient) Get(ctx context.Context, path string, encoding gnmi.Encoding) (*gnmi.GetResponse, error) {
	_, pElem, err := utils.ParsePath(path)
	if err != nil {
		return nil, fmt.Errorf("Parse xpath error: %s", err)
	}

	pathList := []*gnmi.Path{pElem}

	getRequest := &gnmi.GetRequest{
		Encoding: encoding,
		Path:     pathList,
	}

	getResponse, err := c.GNMIClient.Get(ctx, getRequest)

	if err != nil {
		return nil, fmt.Errorf("Get failed: %v", err)
	}
	return getResponse, nil

}

func (c *NXOSGNMIClient) Set(ctx context.Context, path string, value *gnmi.TypedValue, opt action.SubOptValue) (*gnmi.SetResponse, error) {
	_, pElem, err := utils.ParsePath(path)
	if err != nil {
		return nil, fmt.Errorf("Parse xpath error: %s", err)
	}
	updateElem := &gnmi.Update{
		Path: pElem,
		Val:  value,
	}

	updatePathList := []*gnmi.Update{updateElem}
	var setRequest *gnmi.SetRequest
	switch opt {
	case action.Merge:
		setRequest = &gnmi.SetRequest{
			Update: updatePathList,
		}
	case action.Replace:
		setRequest = &gnmi.SetRequest{
			Replace: updatePathList,
		}

	}
	response, err := c.GNMIClient.Set(ctx, setRequest)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (c *NXOSGNMIClient) Subscribe(ctx context.Context, subscriptions []*gnmi.Subscription, encoding gnmi.Encoding) error {
	subMode := gnmi.SubscriptionList_STREAM
	subRequest := &gnmi.SubscribeRequest{
		Request: &gnmi.SubscribeRequest_Subscribe{
			Subscribe: &gnmi.SubscriptionList{
				Encoding:     encoding,
				Mode:         subMode,
				Subscription: subscriptions,
				UpdatesOnly:  false,
			},
		},
	}
	subClient, err := c.GNMIClient.Subscribe(ctx)
	c.GNMISubClient = subClient
	if err != nil {
		return fmt.Errorf("Create subscribe client failed: %v", err)
	}
	if err := c.GNMISubClient.Send(subRequest); err != nil {
		return fmt.Errorf("Failed to send request: %v", err)
	}
	return nil
}
