package typhoon

import (
	"bufio"
	"log"
	"math/rand"
	"net"
	"reflect"
	"time"

	uuid "github.com/TyphoonMC/go.uuid"
)

type Core struct {
	connCounter      int
	eventHandlers    map[reflect.Type][]EventCallback
	brand            string
	rootCommand      CommandNode
	compiledCommands []commandNode
	playerRegistry   *PlayerRegistry
	world            *Map
	gamemode         Gamemode
	difficulty       Difficulty
}

func Init() *Core {
	initConfig()
	initPackets()
	initHacks()
	c := &Core{
		0,
		make(map[reflect.Type][]EventCallback),
		"typhoon",
		CommandNode{
			commandNodeTypeRoot,
			nil,
			nil,
			nil,
			"",
			nil,
		},
		nil,
		newPlayerRegistry(),
		&Map{
			Location{0, 0, 0},
			END,
			[]*Chunk{},
		},
		SPECTATOR,
		PEACEFUL,
	}
	c.compileCommands()
	return c
}

func (c *Core) Start() {
	ln, err := net.Listen("tcp", config.ListenAddress)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Server launched on port", config.ListenAddress)
	go c.keepAlive()
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Print(err)
		} else {
			c.connCounter += 1
			go c.handleConnection(conn, c.connCounter)
		}
	}
}

func (c *Core) SetMap(world *Map) {
	c.world = world
}

func (c *Core) SetGamemode(gamemode Gamemode) {
	c.gamemode = gamemode
}

func (c *Core) SetBrand(brand string) {
	br := make([]byte, len(brand)+1)
	copy(br[:len(brand)], []byte(brand))
	c.brand = string(br)
}

func (c *Core) GetPlayerRegistry() *PlayerRegistry {
	return c.playerRegistry
}

func (c *Core) keepAlive() {
	r := rand.New(rand.NewSource(15768735131534))
	keepalive := &PacketPlayKeepAlive{
		Identifier: 0,
	}
	for {
		c.playerRegistry.ForEachPlayer(func(player *Player) {
			if player.state == PLAY {
				//TODO rework keepalive
				/*if player.keepalive != 0 {
					player.Kick("Timed out")
				}*/

				id := int(r.Int31())
				keepalive.Identifier = id
				player.keepalive = id
				player.WritePacket(keepalive)
			}
		})
		time.Sleep(5000000000)
	}
}

func (c *Core) handleConnection(conn net.Conn, id int) {
	log.Printf("%s(#%d) connected.", conn.RemoteAddr().String(), id)

	player := &Player{
		core:     c,
		id:       id,
		conn:     conn,
		state:    HANDSHAKING,
		protocol: V1_10,
		io: &ConnReadWrite{
			rdr: bufio.NewReader(conn),
			wtr: bufio.NewWriter(conn),
		},
		inaddr: InAddr{
			"",
			0,
		},
		name:         "",
		uuid:         uuid.FromStringOrNil("7065bc74-dfef-475a-a424-d3ab355fcd4a"),
		keepalive:    0,
		compression:  false,
		packetsQueue: make(chan Packet),
	}

	go func() {
		for {
			packet := <-player.packetsQueue
			err := player.privateWritePacket(packet)
			if err != nil {
				break
			}
		}
	}()

	for {
		_, err := player.ReadPacket()
		if err != nil {
			break
		}
	}

	if player.state == PLAY {
		player.core.CallEvent(&PlayerQuitEvent{player})
		player.unregister()
	}
	conn.Close()
	log.Printf("%s(#%d) disconnected.", conn.RemoteAddr().String(), id)
}
