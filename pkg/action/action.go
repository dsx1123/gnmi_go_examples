package action

import "github.com/dsx1123/gnmi_go/pkg/config"

type OptValue int
type SubOptValue int

const (
	Capabilities OptValue = iota + 1
	Get
	Set
	Subscribe
	EDA
)

const (
	Merge SubOptValue = iota + 1
	Replace
)

type Action struct {
	Opt          OptValue
	SubOpt       SubOptValue
	Path         string
	Data         *[]byte
	Subscrptions *config.Subscriptions
}
