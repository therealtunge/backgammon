package server

import (
	"encoding/json"
	"io"
	"math/rand"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/therealtunge/backgammon/common"
	"github.com/therealtunge/backgammon/common/logging"
	pb "github.com/therealtunge/backgammon/pb/autogen"
	"google.golang.org/protobuf/proto"
)

func handleErr(e error) bool {
	if e != nil {
		if e.Error() == "EOF" {
			return true
		}
		panic(e)
	}

	return true
}

var conns map[uint32]net.Conn = make(map[uint32]net.Conn)

type ServerConfig struct {
	// password, must be the same between server and client
	// i.e "v3RyS3cUr3P455w0rd" (dont use this)
	Password string `json:"password"`

	// address to listen on, i.e "0.0.0.0:443"
	ListenAddr string `json:"listenAddr"`

	// use tls?
	UseTLS bool `json:"useTLS"`

	// key file for TLS
	TLSKeyFile string `json:"TLSKeyFile"`

	// cert file for TLS
	TLSCertFile string `json:"TLSCertFile"`
}

var config ServerConfig

func addConn(conn net.Conn) uint32 {
	id := rand.Uint32()
	if conns[id] != nil {
		return addConn(conn)
	}

	conns[id] = conn
	return id
}

func handle(responseWriter http.ResponseWriter, HTTPrequest *http.Request) {

	var err error

	request := pb.Request{}

	wire, err := io.ReadAll(HTTPrequest.Body)
	handleErr(err)

	if err := proto.Unmarshal(wire, &request); err != nil {
		panic(err)
	}

	if request.GetPassword() != config.Password {
		responseWriter.Write([]byte("err: wrong password, nice try"))
		return
	}

	switch request.GetT() {
	case pb.Type_TYPE_CONN:
		dest := request.GetConnect().GetDest()

		addr := HTTPrequest.RemoteAddr

		addr = strings.Split(addr, ":")[0]

		conn, err := net.Dial("tcp", dest)

		if err != nil {
			resp, err := common.CreateResponse(0, nil, err.Error())
			if err != nil {
				logging.Fatal(err)
			}

			responseWriter.Write(resp)
		}
		fd := addConn(conn)

		resp, err := common.CreateResponse(fd, nil, "")
		if err != nil {
			logging.Fatal(err)
		}
		responseWriter.Write(resp)
	case pb.Type_TYPE_WRITE:
		writeRequest := request.GetWrite()
		data := writeRequest.GetData()

		fd := writeRequest.GetFd()

		conns[fd].Write(data)
	case pb.Type_TYPE_READ:
		readRequest := request.GetRead()
		fd := readRequest.GetFd()
		read_data, err := common.FlushConn(conns[fd])

		handleErr(err)

		responseWriter.Write(read_data)
	case pb.Type_TYPE_CLOSE:
		closeRequest := request.GetClose()
		fd := closeRequest.GetFd()
		conns[fd].Close()
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

	decoder.Decode(&config)

	mux := http.NewServeMux()

	mux.HandleFunc("/", handle)

	if config.UseTLS {
		err = http.ListenAndServeTLS(config.ListenAddr, config.TLSCertFile, config.TLSKeyFile, mux)
	} else {
		err = http.ListenAndServe(config.ListenAddr, mux)
	}

	if err != nil {
		panic(err)
	}
}
