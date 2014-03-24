package chansnstruct

import (
	. "net"
	"os"
	"time"
)

const (
	UP         = 0 // is this correct?
	DOWN       = 1 // si this correct?
	N_BUTTONS  = 8 // @Yngve is this correct??
	MAXWAIT    = 5 * time.Second
	PORT       = ":20019"
	N_FLOORS   = 4
	N_ELEV     = 3
	TIMEOFSTOP = 1 // Time spent in one floor
	TIMETRAVEL = 2 // Time of travel between floors
	LOCALHOST  = "129.241.187.157"
	IP1        = "129.241.187.147"
	IP2        = "129.241.187.153"
	//IP3 = 129.241.187.xxx
)

var ExNetChans NetworkExternalChannels
var ExSlaveChans ExternalSlaveChannels
var ExMasterChans ExternalMasterChannels
var ExCommChans ExternalCommunicationChannels
var ExStateMChans ExternalStateMachineChannels
var ExOptimalChans ExternalOptimalizationChannels

var InteruptChan chan os.Signal

type Master struct {
	SlaveIp      []*UDPAddr
	ExternalList map[*UDPAddr]*[N_FLOORS][2]bool
	Statelist    map[*UDPAddr]IpState
}

type Slave struct {
	IP           *UDPAddr
	ExternalList map[*UDPAddr]*[N_FLOORS][2]bool
	InternalList []bool
	CurrentFloor int
	Direction    int
}
type State struct {
	Direction    int
	CurrentFloor int
}
type IpState struct {
	Ip  *UDPAddr
	Sta State
}

type Order struct {
	Floor      int
	ButtonType int
	TurnOn     bool
}
type IpOrderMessage struct {
	Ip  *UDPAddr
	Ord Order
}

type IpOrderList struct {
	Ip           *UDPAddr
	ExternalList map[*UDPAddr]*[N_FLOORS][2]bool
}

type NetworkExternalChannels struct {
	ToNetwork  chan []byte
	ToComm     chan []byte
	ToCommAddr chan *UDPAddr
	StopWrite  chan bool
}
type ExternalOptimalizationChannels struct {
	//InMasterChans.OptimizationInitChan = make(chan Master)
	OptimizationTriggerChan chan IpOrderMessage
	OptimizationReturnChan  chan IpOrderMessage
}

type ExternalCommunicationChannels struct {
	//communication channels
	ToMasterOrderListReceivedChan        chan IpOrderList    //"ore"-
	ToMasterImSlaveChan                  chan IpOrderMessage //"ias"-
	ToMasterOrderExecutedChan            chan IpOrderMessage //"oex"-
	ToMasterOrderExecutedReConfirmedChan chan IpOrderMessage //"oce"-
	ToMasterExternalButtonPushedChan     chan IpOrderMessage //"ebp"-
	ToMasterUpdateState                  chan IpState        //"ust"-

	ToSlaveNetworkInitChan            chan IpOrderList    //"ini"-
	ToSlaveNetworkInitRespChan        chan IpOrderList    //"inr"-
	ToSlaveOrderListChan              chan IpOrderList    //"ord"-
	ToSlaveOrderExecutedConfirmedChan chan IpOrderMessage //"eco"-
	ToSlaveButtonPressedConfirmedChan chan IpOrderMessage //"bpc"-
	ToSlaveUpdateStateReceivedChan    chan IpState        //"sus"-
	ToSlaveImMasterChan               chan string         //"iam"-

}
type ExternalSlaveChannels struct {
	ToCommNetworkInitChan              chan IpOrderList    //"ini"
	ToCommOrderListReceivedChan        chan IpOrderList    //"ore"
	ToCommNetworkInitRespChan          chan IpOrderList    //"ire"
	ToCommOrderExecutedChan            chan Order          //"oex"
	ToCommOrderExecutedReConfirmedChan chan Order          //"oce"
	ToCommExternalButtonPushedChan     chan Order          //"ebp"
	ToCommImSlaveChan                  chan IpOrderMessage //"ias"
	ToCommButtonPressedConfirmedChan   chan Order          //"bpc"
	ToCommUpdatedStateChan             chan State          //"ust"

}
type ExternalMasterChannels struct {
	ToCommOrderListChan              chan IpOrderList    //"ord"
	ToCommOrderExecutedConfirmedChan chan IpOrderMessage //"eco"
	ToCommImMasterChan               chan string         //"iam"
	ToCommUpdateStateReceivedChan    chan IpState        //"sus"
}
type ExternalStateMachineChannels struct {
	ButtonPressedChan      chan Order
	OrderServedChan        chan Order
	CurrentStateChan       chan State
	GetSlaveStructChan     chan bool
	ReturnSlaveStructChan  chan Slave
	DirectionUpdateChan    chan int
	SingleExternalListChan chan [N_FLOORS][2]bool
	LightChan              chan [N_FLOORS][2]bool
}

func Channels_init() {
	network_external_chan_init()
	Communication_external_channels_init()
	Slave_external_chans_init()
	Master_external_chans_init()
	Master_external_chans_init()
	External_state_machine_channels_init()
}

func network_external_chan_init() {
	ExNetChans.ToNetwork = make(chan []byte)
	ExNetChans.ToComm = make(chan []byte)
	ExNetChans.ToCommAddr = make(chan *UDPAddr)
}

func Communication_external_channels_init() {

	ExCommChans.ToMasterOrderListReceivedChan = make(chan IpOrderList)           //"ore"
	ExCommChans.ToMasterOrderExecutedChan = make(chan IpOrderMessage)            //"oex"
	ExCommChans.ToMasterOrderExecutedReConfirmedChan = make(chan IpOrderMessage) //"oce"
	ExCommChans.ToMasterExternalButtonPushedChan = make(chan IpOrderMessage)     //"ebp"
	ExCommChans.ToMasterImSlaveChan = make(chan IpOrderMessage)                  //"ias"
	ExCommChans.ToMasterUpdateState = make(chan IpState)                         //"ust"
	ExCommChans.ToSlaveOrderExecutedConfirmedChan = make(chan IpOrderMessage)    //"eco"
	ExCommChans.ToSlaveOrderListChan = make(chan IpOrderList)                    //"ord"
	ExCommChans.ToSlaveImMasterChan = make(chan string)                          //"iam"
	ExCommChans.ToSlaveButtonPressedConfirmedChan = make(chan IpOrderMessage)    //"bpc"
	ExCommChans.ToSlaveNetworkInitRespChan = make(chan IpOrderList)              //"inr"
	ExCommChans.ToSlaveNetworkInitChan = make(chan IpOrderList)                  //"ini"
	ExCommChans.ToSlaveUpdateStateReceivedChan = make(chan IpState)              //"sus"

}

func Slave_external_chans_init() {
	ExSlaveChans.ToCommOrderListReceivedChan = make(chan IpOrderList)  //"ore"
	ExSlaveChans.ToCommOrderExecutedChan = make(chan Order)            //"oex"
	ExSlaveChans.ToCommOrderExecutedReConfirmedChan = make(chan Order) //"oce"
	ExSlaveChans.ToCommExternalButtonPushedChan = make(chan Order)     //"ebp"
	ExSlaveChans.ToCommImSlaveChan = make(chan IpOrderMessage)
	ExSlaveChans.ToCommUpdatedStateChan = make(chan State)
	ExSlaveChans.ToCommNetworkInitRespChan = make(chan IpOrderList)
	ExSlaveChans.ToCommNetworkInitChan = make(chan IpOrderList)

}
func Master_external_chans_init() {
	ExMasterChans.ToCommOrderListChan = make(chan IpOrderList)                 //"ord"
	ExMasterChans.ToCommOrderExecutedConfirmedChan = make(chan IpOrderMessage) //"eco"
	ExMasterChans.ToCommImMasterChan = make(chan string)                       //"iam"
}

func External_state_machine_channels_init() {
	ExStateMChans.ButtonPressedChan = make(chan Order)
	ExStateMChans.OrderServedChan = make(chan Order)
	ExStateMChans.CurrentStateChan = make(chan State)
	ExStateMChans.GetSlaveStructChan = make(chan bool)
	ExStateMChans.ReturnSlaveStructChan = make(chan Slave)
	ExStateMChans.DirectionUpdateChan = make(chan int)
	ExStateMChans.SingleExternalListChan = make(chan [N_FLOORS][2]bool)
	ExStateMChans.LightChan = make(chan [N_FLOORS][2]bool)
}
func External_optimization_channel_init() {
	ExOptimalChans.OptimizationTriggerChan = make(chan IpOrderMessage)
	ExOptimalChans.OptimizationReturnChan = make(chan IpOrderMessage)
}

//MEMBER FUNCTIONS
func (m Master) Set_external_list_order(ip *UDPAddr, floor int, buttonType int, ipOrder IpOrderMessage) {
	m.ExternalList[ip][floor][buttonType] = ipOrder.Ord.TurnOn
}
func (m Master) Get_external_list() map[*UDPAddr]*[N_FLOORS][2]bool {
	return m.ExternalList
}
func (s Slave) Overwrite_external_list(newExternalList map[*UDPAddr]*[N_FLOORS][2]bool) {
	s.ExternalList = newExternalList
}
func (s Slave) Get_ip() *UDPAddr {
	return s.IP
}
