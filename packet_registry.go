package typhoon

import (
	"log"
	"reflect"
)

var (
	packets map[int64]reflect.Type = make(map[int64]reflect.Type)
)

type Packet interface {
	Write(*Player) error
	Read(*Player, int) error
	Handle(*Player)
	Id() (int, Protocol)
}

func PacketTypeHash(state State, id int) int64 {
	return int64(id) ^ (int64(state) << 32)
}

func initPackets() {
	packets[PacketTypeHash(HANDSHAKING, 0x00)] = reflect.TypeOf((*PacketHandshake)(nil)).Elem()
	packets[PacketTypeHash(STATUS, 0x00)] = reflect.TypeOf((*PacketStatusRequest)(nil)).Elem()
	packets[PacketTypeHash(STATUS, 0x01)] = reflect.TypeOf((*PacketStatusPing)(nil)).Elem()
	packets[PacketTypeHash(LOGIN, 0x00)] = reflect.TypeOf((*PacketLoginStart)(nil)).Elem()
	packets[PacketTypeHash(PLAY, 0x01)] = reflect.TypeOf((*PacketPlayTabCompleteServerbound)(nil)).Elem()
	packets[PacketTypeHash(PLAY, 0x02)] = reflect.TypeOf((*PacketPlayChat)(nil)).Elem()
	packets[PacketTypeHash(PLAY, 0x03)] = reflect.TypeOf((*PacketPlayClientStatus)(nil)).Elem()
	packets[PacketTypeHash(PLAY, 0x09)] = reflect.TypeOf((*PacketPlayPluginMessage)(nil)).Elem()
	packets[PacketTypeHash(PLAY, 0x0B)] = reflect.TypeOf((*PacketPlayKeepAlive)(nil)).Elem()
}

func (player *Player) HandlePacket(id int, length int) (packet Packet, err error) {
	typ := packets[PacketTypeHash(player.state, id)]

	log.Printf("%d -> Test %d \n", player.id, typ)
	if typ == nil {
		if config.Logs {
			log.Printf("%d -> Unknown packet #%d\n", player.id, id)
		}

		var buff []byte
		nbr := 0
		if length > 500 {
			buff = make([]byte, 500)
		} else {
			buff = make([]byte, length)
		}

		for nbr < length {
			if length-nbr > 500 {
				player.io.rdr.Read(buff)
				nbr += 500
			} else {
				player.io.rdr.Read(buff[:length-nbr])
				nbr = length
			}
		}
		return nil, nil
	}

	packet, _ = reflect.New(typ).Interface().(Packet)
	if err = packet.Read(player, length); err != nil {
		return nil, err
	}
	return
}
