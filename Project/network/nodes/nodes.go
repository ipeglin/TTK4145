package nodes

import (
	"fmt"
	"net"
	"network/conn"
	"sort"
	"time"
)

type NetworkNodeRegistry struct {
	Nodes []string
	New   string
	Lost  []string
}

const interval = 150 * time.Millisecond
const timeout = 500 * time.Millisecond

func Client(port int, id string, enableTransmit <-chan bool) {
	conn := conn.DialBroadcastUDP(port)
	addr, _ := net.ResolveUDPAddr("udp", fmt.Sprintf("255.255.255.255:%d", port))

	enable := true
	for {
		select {
		case enable = <-enableTransmit:
		case <-time.After(interval):
		}
		if enable {
			conn.WriteTo([]byte(id), addr)
		}
	}
}

func Server(port int, updateChannel chan<- NetworkNodeRegistry) {
	var buffer [1024]byte
	var reg NetworkNodeRegistry
	lastSeen := make(map[string]time.Time)

	conn := conn.DialBroadcastUDP(port)

	go func() {
		for {
			fmt.Println("Known nodes:", reg)
			time.Sleep(1 * time.Second)
		}
	}()

	for {
		updated := false

		conn.SetReadDeadline(time.Now().Add(interval))
		n, _, _ := conn.ReadFrom(buffer[0:])

		id := string(buffer[:n])

		// Adding new connection
		reg.New = ""
		if id != "" {
			if _, idExists := lastSeen[id]; !idExists {
				reg.New = id
				updated = true
			}

			lastSeen[id] = time.Now()
		}

		// Removing dead connection
		reg.Lost = make([]string, 0)
		for k, v := range lastSeen {
			if time.Now().Sub(v) > timeout {
				updated = true
				reg.Lost = append(reg.Lost, k)
				delete(lastSeen, k)
			}
		}

		// Sending update
		if updated {
			reg.Nodes = make([]string, 0, len(lastSeen))

			for k, _ := range lastSeen {
				reg.Nodes = append(reg.Nodes, k)
			}

			sort.Strings(reg.Nodes)
			sort.Strings(reg.Lost)
			updateChannel <- reg
		}
	}
}
