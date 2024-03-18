package tpu

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"math"
	"net"
	"sort"
	"sync"
	"time"

	"github.com/Rabbitcoccus/solana-go"
	"github.com/Rabbitcoccus/solana-go/rpc"
	"github.com/Rabbitcoccus/solana-go/rpc/ws"
	quic "github.com/quic-go/quic-go"
)

var MAX_SLOT_SKIP_DISTANCE uint64 = 48
var DEFAULT_FANOUT_SLOTS uint64 = 12
var MAX_FANOUT_SLOTS uint64 = 100

var cert, _ = CreateTlsCertificate()

type LeaderTPUCache struct {
	LeaderTPUMap      map[string]string
	Connection        *rpc.Client
	FirstSlot         uint64
	SlotsInEpoch      uint64
	LastEpochInfoSlot uint64
	Leaders           []solana.PublicKey
}

func (leaderTPUCache *LeaderTPUCache) Load(connection *rpc.Client, startSlot uint64) error {
	leaderTPUCache.Connection = connection
	epochInfo, err := leaderTPUCache.Connection.GetEpochInfo(rpc.CommitmentProcessed)
	if err != nil {
		return err
	}
	leaderTPUCache.SlotsInEpoch = epochInfo.SlotsInEpoch
	slotLeaders, err := leaderTPUCache.FetchSlotLeaders(startSlot, leaderTPUCache.SlotsInEpoch)
	if err != nil {
		return err
	}
	leaderTPUCache.Leaders = slotLeaders
	clusterTPUSockets, err := leaderTPUCache.FetchClusterTPUSockets()
	if err != nil {
		return err
	}
	leaderTPUCache.LeaderTPUMap = clusterTPUSockets
	return nil
}

func (leaderTPUCache *LeaderTPUCache) FetchSlotLeaders(startSlot uint64, slotsInEpoch uint64) ([]solana.PublicKey, error) {
	fanout := uint64(math.Min(float64(2*MAX_FANOUT_SLOTS), float64(slotsInEpoch)))
	slotLeaders, err := leaderTPUCache.Connection.GetSlotLeaders(startSlot, fanout)
	if err != nil {
		return nil, err
	}
	return slotLeaders, nil
}

func (leaderTPUCache *LeaderTPUCache) FetchClusterTPUSockets() (map[string]string, error) {
	var clusterTPUSockets = make(map[string]string)
	clusterNodes, err := leaderTPUCache.Connection.GetClusterNodes()
	if err != nil {
		return nil, err
	}
	for _, contactInfo := range clusterNodes {
		if contactInfo.TPU != nil {
			clusterTPUSockets[contactInfo.Pubkey.String()] = *contactInfo.TPUQuic
		}
	}
	return clusterTPUSockets, nil
}

func (leaderTPUCache *LeaderTPUCache) LastSlot() uint64 {
	return leaderTPUCache.FirstSlot + uint64(len(leaderTPUCache.Leaders)) - 1
}

func (leaderTPUCache *LeaderTPUCache) GetSlotLeader(slot uint64) solana.PublicKey {
	if slot >= leaderTPUCache.FirstSlot {
		return leaderTPUCache.Leaders[slot-leaderTPUCache.FirstSlot]
	} else {
		return solana.PublicKey{}
	}
}

func (leaderTPUCache *LeaderTPUCache) GetLeaderSockets(fanoutSlots uint64) []string {
	var alreadyCheckedLeaders []string
	var leaderTPUSockets []string
	var checkedSlots uint64 = 0
	for _, leader := range leaderTPUCache.Leaders {
		tpuSocket := leaderTPUCache.LeaderTPUMap[leader.String()]
		if tpuSocket != "" {
			isDuplicate := CheckIfDuplicate(alreadyCheckedLeaders, leader.String())
			if !isDuplicate {
				alreadyCheckedLeaders = append(alreadyCheckedLeaders, leader.String())
				leaderTPUSockets = append(leaderTPUSockets, tpuSocket)
			}
		}
		checkedSlots++
		if checkedSlots == fanoutSlots {
			return leaderTPUSockets
		}
	}
	return leaderTPUSockets
}

func (leaderTPUCache *LeaderTPUCache) GetLeaderSocketsConverted(fanoutSlots uint64) []*net.UDPAddr {
	var alreadyCheckedLeaders []string
	var leaderTPUSockets []*net.UDPAddr
	var checkedSlots uint64 = 0
	for _, leader := range leaderTPUCache.Leaders {
		tpuSocket := leaderTPUCache.LeaderTPUMap[leader.String()]
		if tpuSocket != "" {
			isDuplicate := CheckIfDuplicate(alreadyCheckedLeaders, leader.String())
			if !isDuplicate {
				alreadyCheckedLeaders = append(alreadyCheckedLeaders, leader.String())
				leaderAddress, _ := net.ResolveUDPAddr("udp", tpuSocket)
				leaderTPUSockets = append(leaderTPUSockets, leaderAddress)
			}
		}
		checkedSlots++
		if checkedSlots == fanoutSlots {
			return leaderTPUSockets
		}
	}
	return leaderTPUSockets
}

type RecentLeaderSlots struct {
	RecentSlots []float64
}

func (recentLeaderSlots *RecentLeaderSlots) Load(currentSlot uint64) {
	recentLeaderSlots.RecentSlots = append(recentLeaderSlots.RecentSlots, float64(currentSlot))
}

func (recentLeaderSlots *RecentLeaderSlots) RecordSlot(currentSlot uint64) {
	recentLeaderSlots.RecentSlots = append(recentLeaderSlots.RecentSlots, float64(currentSlot))
	for len(recentLeaderSlots.RecentSlots) > 12 {
		recentLeaderSlots.RecentSlots = recentLeaderSlots.RecentSlots[1:]
	}
}

func (recentLeaderSlots *RecentLeaderSlots) EstimatedCurrentSlot() uint64 {
	if len(recentLeaderSlots.RecentSlots) == 0 {
		return 0
	}
	recentSlots := recentLeaderSlots.RecentSlots
	sort.Float64s(recentSlots)
	maxIndex := len(recentSlots) - 1
	medianIndex := maxIndex / 2
	medianRecentSlot := recentSlots[medianIndex]
	expectedCurrentSlot := uint64(medianRecentSlot) + uint64(maxIndex-medianIndex)
	maxReasonableCurrentSlot := expectedCurrentSlot + MAX_SLOT_SKIP_DISTANCE
	sort.Sort(sort.Reverse(sort.Float64Slice(recentSlots)))
	var slotToReturn uint64 = 0
	for _, slot := range recentSlots {
		if uint64(slot) <= maxReasonableCurrentSlot && uint64(slot) > slotToReturn {
			slotToReturn = uint64(slot)
		}
	}
	return slotToReturn
}

type LeaderTPUService struct {
	RecentSlots       *RecentLeaderSlots
	LTPUCache         *LeaderTPUCache
	Subscription      *ws.SlotsUpdatesSubscription
	Connection        *rpc.Client
	WSConnection      *ws.Client
	LeaderConnections []net.Conn
}

func (leaderTPUService *LeaderTPUService) Load(connection *rpc.Client, websocketURL string, fanout uint64) error {
	leaderTPUService.Connection = connection
	slot, err := leaderTPUService.Connection.GetSlot(rpc.CommitmentProcessed)
	if err != nil {
		return err
	}
	recentSlots := RecentLeaderSlots{}
	recentSlots.Load(slot)
	leaderTPUService.RecentSlots = &recentSlots
	leaderTPUCache := LeaderTPUCache{}
	err = leaderTPUCache.Load(leaderTPUService.Connection, slot)
	if err != nil {
		return err
	}
	leaderTPUService.LTPUCache = &leaderTPUCache
	if websocketURL != "" {
		wsConnection, err := ws.Connect(context.TODO(), websocketURL)
		if err == nil {
			subscription, err := wsConnection.SlotsUpdatesSubscribe()
			if err == nil {
				leaderTPUService.Subscription = subscription
				go func() {
					for {
						message, err := leaderTPUService.Subscription.Recv()
						if err == nil {
							//Slot already full, skip over 1 slot.
							if message.Type == ws.SlotsUpdatesCompleted {
								leaderTPUService.RecentSlots.RecordSlot(message.Slot + 1)
							}
							//Slot received first shred, it's still accepting transactions so we record.
							if message.Type == ws.SlotsUpdatesFirstShredReceived {
								leaderTPUService.RecentSlots.RecordSlot(message.Slot)
							}
						}
					}
				}()
			} else {
				leaderTPUService.Connection = nil
			}
		} else {
			leaderTPUService.Connection = nil
		}
	} else {
		leaderTPUService.Connection = nil
	}
	go leaderTPUService.Run(fanout)
	return nil
}

func (leaderTPUService *LeaderTPUService) LeaderTPUSockets(fanoutSlots uint64) []string {
	return leaderTPUService.LTPUCache.GetLeaderSockets(fanoutSlots)
}

func (leaderTPUService *LeaderTPUService) Run(fanout uint64) {
	var lastClusterRefreshTime = time.Now()
	var sleepMs = 1000
	for {
		if time.Now().UnixMilli()-lastClusterRefreshTime.UnixMilli() > 5000*60 {
			latestTPUSockets, err := leaderTPUService.LTPUCache.FetchClusterTPUSockets()
			if err != nil || latestTPUSockets == nil {
				sleepMs = 100
				continue
			}
			leaderTPUService.LTPUCache.LeaderTPUMap = latestTPUSockets
			lastClusterRefreshTime = time.Now()
		}

		time.Sleep(time.Duration(sleepMs) * time.Millisecond)

		currentSlot := leaderTPUService.RecentSlots.EstimatedCurrentSlot()
		if currentSlot >= (leaderTPUService.LTPUCache.LastSlot() - fanout) {
			slotLeaders, err := leaderTPUService.LTPUCache.FetchSlotLeaders(currentSlot, leaderTPUService.LTPUCache.SlotsInEpoch)
			if err != nil || slotLeaders == nil {
				sleepMs = 100
				continue
			}
			leaderTPUService.LTPUCache.FirstSlot = currentSlot
			leaderTPUService.LTPUCache.Leaders = slotLeaders
		}
		sleepMs = 1000
	}
}

type TPUClientConfig struct {
	FanoutSlots uint64
}

type TPUClient struct {
	FanoutSlots uint64
	LTPUService *LeaderTPUService
	Exit        bool
	Connection  *rpc.Client
}

func (tpuClient *TPUClient) Load(connection *rpc.Client, websocketURL string, config TPUClientConfig) error {
	tpuClient.Connection = connection
	tpuClient.FanoutSlots = uint64(math.Max(math.Min(float64(config.FanoutSlots), float64(MAX_FANOUT_SLOTS)), 1))
	tpuClient.Exit = false
	leaderTPUService := LeaderTPUService{}
	tpuClient.LTPUService = &leaderTPUService
	err := tpuClient.LTPUService.Load(tpuClient.Connection, websocketURL, tpuClient.FanoutSlots)
	if err != nil {
		return err
	}
	return nil
}

func (tpuClient *TPUClient) SendTransaction(transaction *solana.Transaction, amount int) (solana.Signature, error) {
	fmt.Println("sending transaction (ver=2) ...")
	rawTransaction, err := transaction.MarshalBinary()
	if err != nil {
		return solana.Signature{}, err
	}
	err = tpuClient.SendRawTransaction(rawTransaction, amount)
	if err != nil {
		return solana.Signature{}, err
	}
	return transaction.Signatures[0], nil
}

func (tpuClient *TPUClient) SendTransactionThroughSocket(transaction *solana.Transaction, amount int, socket *net.UDPConn) (solana.Signature, error) {
	rawTransaction, err := transaction.MarshalBinary()
	if err != nil {
		return solana.Signature{}, err
	}
	err = tpuClient.SendRawTransactionThroughSocket(rawTransaction, amount, socket)
	if err != nil {
		return solana.Signature{}, err
	}
	return transaction.Signatures[0], nil
}

func (tpuClient *TPUClient) SendRawTransaction(transaction []byte, amount int) error {
	var successes = 0
	var lastError = ""
	leaderTPUSockets := tpuClient.LTPUService.LeaderTPUSockets(tpuClient.FanoutSlots)
	fmt.Println(leaderTPUSockets)
	var wg sync.WaitGroup
	for _, leader := range leaderTPUSockets {
		wg.Add(1)
		go func(leader string) {
			defer wg.Done()
			var connectionTries = 0
			var failed = false
			var connection quic.Connection
			var stream quic.SendStream
			var err1, err2 error
			for {
				fmt.Println(time.Now().Format("2006-01-02T15:04:05.000Z07"), "dialing ", leader)
				connection, err1 = quic.DialAddr(context.Background(), leader, &tls.Config{
					Certificates:       []tls.Certificate{cert},
					NextProtos:         []string{"solana-tpu"},
					InsecureSkipVerify: true,
				}, &quic.Config{
					HandshakeIdleTimeout: 10 * time.Second,
				})
				if err1 == nil {
					fmt.Println(time.Now(), "creating stream ", leader)
					stream, err2 = connection.OpenUniStreamSync(context.Background())
				}
				if err1 != nil || err2 != nil {
					fmt.Printf("%s err1: %v, err2: %v\n", leader, err1, err2)
					if err1 != nil {
						lastError = err1.Error()
					} else {
						lastError = err2.Error()
					}
					if connectionTries < 3 {
						connectionTries++
						continue
					} else {
						failed = true
						break
					}
				}
				fmt.Println(time.Now().Format("2006-01-02T15:04:05.000Z07"), "connected")
				break
			}
			if failed {
				return
			}
			for i := 0; i < amount; i++ {
				println("send tx to tpu")
				// err := connection.SendDatagram(transaction)
				_, err := stream.Write(transaction)
				if err != nil {
					println("send error ", err.Error())
					lastError = err.Error()
				} else {
					successes++
				}
			}
			stream.Close()
			connection.CloseWithError(0, "")
		}(leader)

	}
	wg.Wait()
	if successes == 0 {
		return errors.New(lastError)
	} else {
		return nil
	}
}

func (tpuClient *TPUClient) SendRawTransactionThroughSocket(transaction []byte, amount int, socket *net.UDPConn) error {
	for _, leader := range tpuClient.LTPUService.LTPUCache.GetLeaderSocketsConverted(tpuClient.FanoutSlots) {
		for i := 0; i < amount; i++ {
			socket.WriteToUDP(transaction, leader)
		}
	}
	return nil
}

func New(connection *rpc.Client, websocketURL string, config TPUClientConfig) (*TPUClient, error) {
	tpuClient := TPUClient{}
	err := tpuClient.Load(connection, websocketURL, config)
	return &tpuClient, err
}

func sendPacket(stream quic.Stream, transaction []byte) error {
	const PACKET_SIZE = 1232
	offset := 0
	for offset < len(transaction) {
		var packet []byte
		if offset+PACKET_SIZE >= len(transaction) {
			packet = transaction[offset:]
		} else {
			packet = transaction[offset : offset+PACKET_SIZE]
		}
		_, err := stream.Write(packet)
		if err != nil {
			return err
		}
		offset += PACKET_SIZE
	}
	return nil
}
