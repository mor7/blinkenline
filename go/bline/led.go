package bline

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"

	"flag"
)

type LedConfig struct {
	Servers []struct {
		Address  string
		Port     int
		LedCount int
		Reverse  bool
	}
}

type LedClient struct {
	Addr       *net.UDPAddr
	Connection *net.UDPConn
	LedCount   int
	Reverse    bool
}

var clients []LedClient
var buffer []byte
var LedCount int
var alpha float64

var palpha = flag.Int("alpha", 255, "Brightness of the colors with 0 being black and 255 being unchanged")
var pconf = flag.String("c", "config.json", "Path to the confguration file.")

// Initialises the library
// Call it after setting your flags and before using this library
func Init() error {
	if !flag.Parsed() {
		flag.Parse()
	}

	alpha = float64(*palpha)
	if alpha < 0 {
		alpha = 0
	} else if alpha > 255 {
		alpha = 255
	}
	alpha /= 255

	file, err := os.Open(*pconf)
	if err != nil {
		return err
	}

	decoder := json.NewDecoder(file)
	config := LedConfig{}
	err = decoder.Decode(&config)
	if err != nil {
		return err
	}
	for _, server := range config.Servers {
		addr := fmt.Sprintf("%s:%d", server.Address, server.Port)
		client, err := createClient(addr)
		if err != nil {
			return err
		}

		client.LedCount = server.LedCount
		client.Reverse = server.Reverse
		clients = append(clients, client)

		LedCount += server.LedCount
	}
	buffer = make([]byte, LedCount*3)
	return nil
}

func createClient(address string) (LedClient, error) {
	ledClient := LedClient{}
	if address == "" {
		return ledClient, errors.New("No Address was given.")
	}
	udpaddr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		return ledClient, err
	}

	ledClient.Addr = udpaddr
	conn, err := net.DialUDP(udpaddr.Network(), nil, udpaddr)
	if err != nil {
		return ledClient, err
	}
	ledClient.Connection = conn
	return ledClient, nil
}

func ClearBuffer() {
	buffer = make([]byte, LedCount*3)
}

func SetColor(pos int, color int) {
	r, g, b := SplitRGB(color)
	buffer[pos*3] = r
	buffer[pos*3+1] = g
	buffer[pos*3+2] = b
}

func GetColor(pos int) (byte, byte, byte) {
	r := buffer[pos*3]
	g := buffer[pos*3+1]
	b := buffer[pos*3+2]
	return r, g, b
}

// Sends the buffer
// Doesn't clear the buffer
func SendBuffer() error {
	buffer = buffer[0 : len(buffer)-len(buffer)%3]
	i := 0
	for _, client := range clients {
		subBuf := buffer[i : i+client.LedCount*3]
		if client.Reverse {
			revBuf := make([]byte, len(buffer))
			for i := range subBuf {
				ri := (len(subBuf)-i-1)/3*3 + i%3
				revBuf[ri] = subBuf[i]
			}
			subBuf = revBuf
		}
		for i := range subBuf {
			subBuf[i] = byte(float64(subBuf[i]) * alpha)
		}

		_, err := client.Connection.Write(subBuf)
		if err != nil {
			return err
		}

		i += client.LedCount * 3
		if i > len(buffer)-1 {
			break
		}
	}
	return nil
}

func Close() {
	for _, client := range clients {
		client.Connection.Close()
	}
}
