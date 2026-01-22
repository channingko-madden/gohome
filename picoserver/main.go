package main

import (
	"bufio"
	"encoding/json"
	"io"
	"machine"
	"net/netip"
	"time"

	"log/slog"

	"github.com/soypat/cyw43439"

	"github.com/soypat/seqs/httpx"
	"github.com/soypat/seqs/stacks"
)

const (
	connTimeout = 3 * time.Second
	maxconns    = 3
	tcpbufsize  = 2030
	listenPort  = 80
	hostname    = "picotemp"
)

type climate struct {
	TempC float64 `json:"tempC"`
	TempF float64 `json:"tempF"`
	RH    float64 `json:"rh"`
}

var logger *slog.Logger // One logger to reduce memory usage.

func init() {
	logger = slog.New(
		slog.NewTextHandler(machine.Serial, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}))
}

func changeLEDState(dev *cyw43439.Device, state bool) {
	if err := dev.GPIOSet(0, state); err != nil {
		logger.Error("failed to change LED state:", slog.String("err", err.Error()))
	}
}

func setupDevice() (*stacks.PortStack, *cyw43439.Device) {
	_, stack, dev, err := SetupWithDHCP(SetupConfig{
		Hostname: hostname,
		Logger:   logger,
		TCPPorts: 1,
	})

	if err != nil {
		panic("setup DHCP:" + err.Error())
	}

	// Turn LED on
	changeLEDState(dev, true)

	return stack, dev
}

func newListener(stack *stacks.PortStack) *stacks.TCPListener {
	// Start TCP server.
	listenAddr := netip.AddrPortFrom(stack.Addr(), listenPort)
	listener, err := stacks.NewTCPListener(
		stack, stacks.TCPListenerConfig{
			MaxConnections: maxconns,
			ConnTxBufSize:  tcpbufsize,
			ConnRxBufSize:  tcpbufsize,
		},
	)

	if err != nil {
		panic("listener create:" + err.Error())
	}

	err = listener.StartListening(listenPort)
	if err != nil {
		panic("listener start:" + err.Error())
	}

	logger.Info("listening", slog.String("addr", "http://"+listenAddr.String()))

	return listener

}

// go in goroutine, pass count into the channel to blink
func blinkLED(dev *cyw43439.Device, blink chan uint) {
	for {
		select {
		case n := <-blink:
			lastLedState := true
			if n == 0 {
				n = 5
			}
			for i := uint(0); i < n; i++ {
				lastLedState = !lastLedState
				changeLEDState(dev, lastLedState)
				time.Sleep(250 * time.Millisecond)
			}

			// Ensure LED is on at the end
			changeLEDState(dev, true)
		}
	}
}

func getPicoTemperature() *climate {
	curTemp := machine.ReadTemperature() // in millicelcius

	return &climate{
		TempC: float64(curTemp) / 1000,
		TempF: ((float64(curTemp) / 1000) * 9 / 5) + 32,
	}
}

type TemperatureSensor interface {
	ReadClimate() *climate
}

func HTTPHandler(respWriter io.Writer, resp *httpx.ResponseHeader, sensor TemperatureSensor) {
	resp.SetConnectionClose()
	logger.Info("Got request...")

	t := sensor.ReadClimate()

	if t == nil {
		resp.SetStatusCode(500)
		respWriter.Write(resp.Header())
		return
	}

	body, err := json.Marshal(t)

	if err != nil {
		logger.Error(
			"temperature json:",
			slog.String("err", err.Error()),
		)
		resp.SetStatusCode(500)
	} else {
		resp.SetContentType("application/json")
		resp.SetContentLength(len(body))
	}

	respWriter.Write(resp.Header())
	respWriter.Write(body)

}

func handleConnection(listener *stacks.TCPListener, blink chan uint) {
	// Reuse the same buffers for each connection to avoid heap allocations.
	// This is an embedded device remember!
	var resp httpx.ResponseHeader
	buf := bufio.NewReaderSize(nil, 1024)

	sensor := configureSensor()

	for {
		conn, err := listener.Accept()
		if err != nil {
			logger.Error(
				"listener accept:",
				slog.String("err", err.Error()),
			)
			time.Sleep(time.Second)
			continue
		}

		logger.Info(
			"new connection",
			slog.String("remote", conn.RemoteAddr().String()),
		)

		err = conn.SetDeadline(time.Now().Add(connTimeout))
		if err != nil {
			conn.Close()
			logger.Error(
				"conn set deadline:",
				slog.String("err", err.Error()),
			)
			continue
		}

		buf.Reset(conn)
		resp.Reset()
		HTTPHandler(conn, &resp, sensor)
		conn.Close()

		blink <- 3
	}
}

func main() {
	stack, dev := setupDevice()
	listener := newListener(stack)

	blink := make(chan uint, 3)
	go blinkLED(dev, blink)
	go handleConnection(listener, blink)

	for {
		select {
		case <-time.After(1 * time.Minute):
			logger.Info("Waiting for connections...")
		}
	}
}
