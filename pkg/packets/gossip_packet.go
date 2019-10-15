package packets

import (
  M "../messages"
  S "../status"
  "net"
  "fmt"
)

type GossipPacket struct {
  Simple *M.SimpleMessage
  Rumor *M.RumorMessage
  Status *StatusPacket
}

func initGossipPacket(
  simple *M.SimpleMessage,
  rumor *M.RumorMessage,
  status *StatusPacket,
  ) *GossipPacket{
    return &GossipPacket{Simple : simple, Rumor : rumor, Status : status}
}

func InitSimpleGossipPacket(
	senderName,
	relayAddr,
	content string,
) *GossipPacket {
  msg := M.InitSimpleMessage(senderName, relayAddr, content);
  ret := initGossipPacket(msg, nil, nil);
  return ret;
}

func InitRumorGossipPacket(
	origin string,
	id uint32,
	txt string,
) *GossipPacket {
  rumor := M.InitRumorMessage(origin, id, txt);
  ret := initGossipPacket(nil, rumor, nil);
  return ret;
}

func InitStatusGossipPacket(
	want []S.PeerStatus,
) *GossipPacket {
  statusPacket := InitStatusPacket(want)
  ret := initGossipPacket(nil, nil, statusPacket);
  return ret;
}

func (p *GossipPacket) PrintMessage(sender *net.UDPAddr){
  if p.Simple != nil{
    fmt.Println("SIMPLE MESSAGE origin", p.Simple.OriginalName, "from", sender, "contents", p.Simple.Contents)
  } else if p.Rumor != nil {
    fmt.Println("RUMOR origin", p.Rumor.Origin, "from", sender, "ID", p.Rumor.ID, "contents", p.Rumor.Text)
  } else if p.Status != nil{
    statusString := ""
    for _, ps := range p.Status.Want{
      statusString += " peer " + ps.Identifier + " nextID " + fmt.Sprint(ps.NextID)
    }

    fmt.Println("STATUS from", sender, statusString)
  }

}
