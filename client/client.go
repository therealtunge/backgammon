package client

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/therealtunge/backgammon/common"
	"github.com/therealtunge/backgammon/common/logging"
	pb "github.com/therealtunge/backgammon/pb/autogen"
	"google.golang.org/protobuf/proto"
)

const BUFFER_SIZE = 1024

func flush(c net.Conn) ([]byte, error) {
	var tmp []byte = nil
	var ret []byte = make([]byte, 0)
	var n int = BUFFER_SIZE
	var err error
	for n == 1024 && err == nil {
		tmp = make([]byte, BUFFER_SIZE)
		n, err = c.Read(tmp)
		if err != nil {
			return nil, err
		}
		tmp = tmp[:n] // truncate
		ret = append(ret, tmp...)
	}

	return ret, nil
}

const (
	TYPE_READ  = 1
	TYPE_WRITE = 2
	TYPE_CONN  = 3
	TYPE_CLOSE = 4
)

func createReq(T pb.Type, data []byte, target *string, fd uint32) []byte {
	builder := pb.Request_builder{
		Password: &Config.password,
		T:        &T,
	}

	switch T {
	case pb.Type_TYPE_CONN:
		{
			builder.Connect = common.CreateConnReq(*target)
			break
		}

	case pb.Type_TYPE_READ:
		{
			builder.Read = common.CreateReadReq(fd)
			break
		}

	case pb.Type_TYPE_WRITE:
		{
			builder.Write = common.CreateWriteReq(data, fd)
			break
		}
	case pb.Type_TYPE_CLOSE:
		{
			builder.Close = common.CreateCloseReq(fd)
			break
		}
	}

	wire, err := proto.Marshal(builder.Build())

	if err != nil {
		panic(err)
	}

	return wire
}

type ClientConfigJson struct {
	ProxyPort int `json:"proxyPort"`

	// where to bind (i.e "0.0.0.0")
	ProxyBindIP string `json:"proxyBindIP"`

	// must match the server
	Password string `json:"password"`

	// i.e "myvps.com:443"
	ServerAddress string `json:"serverAddress"`

	// use HTTPS?
	UseHTTPS bool `json:"useHTTPS"`

	// how many milliseconds to wait before sending another read request
	ReadWaitTime int `json:"readWaitTimeMS"`
}

type ClientConfig struct {
	// where to bind (i.e "0.0.0.0:1080")
	proxyBind string

	// must match the server
	password string

	// i.e "https://myvps.com:443"
	serverAddress string

	// how many milliseconds to wait before sending another read request
	readWaitTime int
}

var Config ClientConfig

func onConnect(n net.Conn) {
	tmp := make([]byte, 4)
	n.Read(tmp)

	noauth := []byte{0x05, 0x00}

	n.Write(noauth)

	addr := n.RemoteAddr().String()
	_port, _ := strconv.Atoi(strings.Split(addr, ":")[1])
	port := uint16(_port)
	err, data, pipe, remote := parsePacket(bufio.NewReader(n), port)

	if err != nil {
		panic(err)
	}

	n.Write(data)

	if pipe {
		c := make(chan bool)
		conn_req, err := http.Post(Config.serverAddress, "", bytes.NewReader(createReq(pb.Type_TYPE_CONN, nil, remote, 0)))

		if err != nil {
			logging.Error("onConnect: ", err)
			return
		}

		body, err := io.ReadAll(conn_req.Body)
		if err != nil {
			n.Close()
			logging.Error("onConnect: ", err)
			return
		}

		resp := pb.Response{}
		err = proto.Unmarshal(body, &resp)

		if err != nil {
			n.Close()
			logging.Error("onConnect: ", err)
			return
		}

		if resp.HasError() {
			n.Close()
			logging.Error("server: ", resp.GetError())
			return
		}
		if !resp.HasFd() {
			panic("..")
		}

		fd := resp.GetFd()

		wg := sync.WaitGroup{}
		wg.Add(2)


		// pipe UP
		go func(n net.Conn, t net.Conn, p uint16, wg *sync.WaitGroup) {
			var err error
			var data []byte

			for {
				data, err = flush(n)

				if err != nil {
					wg.Done()
					break
				}

				wire := createReq(pb.Type_TYPE_WRITE, data, nil, fd)

				_, err = http.Post(Config.serverAddress, "", bytes.NewReader(wire))
			}

			wg.Done()
		}(n, conn2port[port], port, &wg)

		// pipe DOWN
		go func(n net.Conn, t net.Conn, p uint16, wg *sync.WaitGroup) {
			var err error
			var r *http.Response
			var data []byte

			for {
				time.Sleep(time.Millisecond * time.Duration(Config.readWaitTime))

				wire := createReq(pb.Type_TYPE_READ, nil, nil, fd)
				r, err = http.Post(Config.serverAddress, "", bytes.NewReader(wire))
				if err != nil {

					c <- true
					break
				}
				data, _ = io.ReadAll(r.Body)

				n.Write(data)
			}

			wg.Done()
		}(n, conn2port[port], port, &wg)

		wg.Wait()
	}
}

func Main() {
	if len(os.Args) < 3 {
		logging.Fatal("not enough arguments! usage: ", os.Args[0], " client {config file}")
	}

	file, err := os.Open(os.Args[2])

	if err != nil {
		logging.Fatal(err)
	}

	decoder := json.NewDecoder(file)

	configJson := ClientConfigJson{}
	decoder.Decode(&configJson)

	uriType := "http"

	if configJson.UseHTTPS {
		uriType = "https"
	}
	Config = ClientConfig{
		password:      configJson.Password,
		proxyBind:     fmt.Sprintf("%s:%d", configJson.ProxyBindIP, configJson.ProxyPort),
		serverAddress: fmt.Sprintf("%s://%s", uriType, configJson.ServerAddress),
		readWaitTime: configJson.ReadWaitTime,
	}

	logging.Info("listening on: ", Config.proxyBind)

	l, err := net.Listen("tcp", Config.proxyBind)
	if err != nil {
		logging.Fatal(err)
	}

	for {
		n, err := l.Accept()
		if err != nil {
			logging.Fatal(err)
		}
		onConnect(n)
	}
}
