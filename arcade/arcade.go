package arcade

import (
	"arcade/arcade/message"
	"flag"
	"fmt"
	"log"
	"os"
	"time"
)

type Arcade struct {
	Distributor bool
	Port        int
	LAN         bool

	Server *Server
}

var arcade = NewArcade()

func NewArcade() *Arcade {
	return &Arcade{
		Distributor: false,
	}
}

func Start() {
	dist := flag.Bool("distributor", false, "Run as a distributor")
	flag.BoolVar(dist, "d", false, "Run as a distributor")

	distributorAddr := flag.String("distributor-addr", "149.28.43.157:6824", "Distributor address")
	flag.StringVar(distributorAddr, "da", "149.28.43.157:6824", "Distributor address")

	port := flag.Int("port", 6824, "Port to listen on")
	flag.IntVar(port, "p", 6824, "Port to listen on")
	flag.Parse()

	// Create log file
	f, err := os.OpenFile(fmt.Sprintf("log-%d", *port), os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)

	if err != nil {
		panic(err)
	}

	defer f.Close()
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	log.SetOutput(f)

	// Register messages
	message.Register(AckGameUpdateMessage{Message: message.Message{Type: "ack_game_update"}})
	message.Register(ClientUpdateMessage[TronClientState]{Message: message.Message{Type: "client_update"}})
	message.Register(DisconnectMessage{Message: message.Message{Type: "disconnect"}})
	message.Register(EndGameMessage{Message: message.Message{Type: "end_game"}})
	message.Register(ErrorMessage{Message: message.Message{Type: "error"}})
	message.Register(GameUpdateMessage[TronGameState, TronClientState]{Message: message.Message{Type: "game_update"}})
	message.Register(HeartbeatMessage{Message: message.Message{Type: "heartbeat"}})
	message.Register(HeartbeatReplyMessage{Message: message.Message{Type: "heartbeat_reply"}})
	message.Register(HelloMessage{Message: message.Message{Type: "hello"}})
	message.Register(JoinMessage{Message: message.Message{Type: "join"}})
	message.Register(JoinReplyMessage{Message: message.Message{Type: "join_reply"}})
	message.Register(LeaveMessage{Message: message.Message{Type: "leave"}})
	message.Register(LobbyEndMessage{Message: message.Message{Type: "lobby_end"}})
	message.Register(LobbyInfoMessage{Message: message.Message{Type: "lobby_info"}})
	message.Register(StartGameMessage{Message: message.Message{Type: "start_game"}})

	arcade.Distributor = *dist
	arcade.Port = *port

	if arcade.Distributor {
		arcade.Server = NewServer(fmt.Sprintf("0.0.0.0:%d", *port), *port, *dist, nil)
		arcade.Server.Start()
		os.Exit(0)
	}

	// Start host server
	mgr := NewViewManager()
	arcade.Server = NewServer(fmt.Sprintf("0.0.0.0:%d", *port), *port, *dist, mgr)
	arcade.Server.Network.Delegate = mgr

	go arcade.Server.Start()

	// TODO: Make better solution for this later -- wait for server to start
	time.Sleep(10 * time.Millisecond)

	// Connect to distributor
	go arcade.Server.Network.Connect(*distributorAddr, "", nil)

	// Start view manager
	splashView := NewSplashView(mgr)
	mgr.Start(splashView)
}
