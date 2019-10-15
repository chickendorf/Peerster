package gossip

import (
	"fmt"
  "net"
	"github.com/dedis/protobuf"
	P "../packets"
	U "../utils"
	S "../status"
	M "../messages"
	"math/rand"
	"time"
)

//#Move
const maxMsgSize = 10000

type Gossiper struct {
	Name string
  gossipAddr *net.UDPAddr
  uiAddr *net.UDPAddr
	uiConn     *net.UDPConn
	gossipConn *net.UDPConn
  peers []string
	simple bool
	rumorDatabase map[string][]P.GossipPacket
	messages []string
	status []S.PeerStatus
	countID uint32
	antiE int
}

func InitGossiper(
	name,
	uiPort,
	gossipAddr string,
	peerList []string,
	simpleMode bool,
	aE int,
) *Gossiper {
  var n = name

  uiAddress, _ := net.ResolveUDPAddr("udp4", "127.0.0.1:"+uiPort)
  uiConnection, _ := net.ListenUDP("udp4", uiAddress)

  gossipAddress, _ := net.ResolveUDPAddr("udp4", gossipAddr)
	gossipConnection, _ := net.ListenUDP("udp4", gossipAddress)
	//fmt.Println("######################## : ", gossipConnection);

  //peersListArg := []string

	ret := Gossiper{Name : n,
									gossipAddr : gossipAddress,
									uiAddr : uiAddress,
									uiConn : uiConnection,
									gossipConn : gossipConnection,
									peers : peerList,
									simple: simpleMode,
									countID : 1,
									rumorDatabase: make(map[string][]P.GossipPacket),
									antiE: aE};


	go ret.ListenGossip();
	return &ret;

}

func (g *Gossiper) checkStatusPeriod(){
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		_ = <-ticker.C
		go g.checkACK()
	}
}

func (g *Gossiper) checkACK(){
	for _, ps := range g.status{
		haveToSend := 0
		for i, gp := range g.rumorDatabase[ps.Identifier]{
			haveToSend = i
			if gp.Rumor.ID == ps.NextID - 1{
				haveToSend = -1
				break
			}
		}

		if haveToSend != -1{
			addr := g.getNextPeers()
			if len(addr) >0{
				peerA, _ := net.ResolveUDPAddr(
					"udp4",
					addr[0],
				)
				g.sendGPacket(g.rumorDatabase[ps.Identifier][haveToSend],peerA)
			}

		}
	}
}

func (g *Gossiper) ListenClient(){
	if g.simple {
		g.ListenClientSimpleMode();
	}else{
		g.ListenClientFull();
	}
}

func (g *Gossiper) ListenClientFull(){
	for {
		g.ListenClientSimpleMode();
	}
}

func (g *Gossiper) ListenClientSimpleMode(){
	for {

		packet := &P.ClientPacket{}
		bytes := make([]byte, 10*maxMsgSize)
		length, sender, err := g.uiConn.ReadFromUDP(bytes)

		if err != nil{
			fmt.Println("Error")
		}

		if length > maxMsgSize {
			fmt.Println(
				"Message size", length, "is not a correct value the limit is", maxMsgSize,)
			continue
		}

		protobuf.Decode(bytes, packet)
		//gossiper.handleClientPacket(packet, sender)

		//fmt.Println("CLIENT MESSAGE :", packet.Simple.Contents)
		packet.PrintMessage();
		g.forwardClientMessage(packet, sender)
	}
}

func (g *Gossiper) SendClientMessage(packet *P.ClientPacket, sender *net.UDPAddr){
	packet.PrintMessage();
	g.forwardClientMessage(packet, sender)
}

func (g *Gossiper) AddPeer(p string){
	g.peers = append(g.peers, p)
}

func (g *Gossiper) ListenGossip(){
	if g.simple {
		g.ListenGossipSimpleMode();
	}else{
		g.ListenGossipFull();
	}
}

func (g *Gossiper) ListenGossipFull(){
	for {
		packet := &P.GossipPacket{}
		bytes := make([]byte, 10*maxMsgSize)
		//fmt.Println("*********************CHECK ERROR", g.gossipConn)
		length, sender, err := g.gossipConn.ReadFromUDP(bytes)

		if err != nil{
			fmt.Println("Error")
		}

		if length > maxMsgSize {
			fmt.Println(
				"Message size", length, "is not a correct value the limit is", maxMsgSize,)
			continue
		}

		protobuf.Decode(bytes, packet)

		packet.PrintMessage(sender)
		g.printPeers()

		if packet.Rumor != nil {
			g.updateStatus(packet.Rumor)
			if g.addRumorInDB(packet){
				//fmt.Println("OK ADD RUMOR IN DB")
				//fine
			}else{
				//fmt.Println("FAIL ADD RUMOR IN DB")
				//Error
			}

			g.forwardGossipMessage(packet, sender)

			// GOroutine après
			go g.sendAck(/*packet,*/ sender)
		}else if packet.Status != nil{
			//Handle the ack
			go g.handleStatusPacket(packet, sender)
		}
	}
}

func (g *Gossiper) ListenGossipSimpleMode(){
	for {
		packet := &P.GossipPacket{}
		bytes := make([]byte, 10*maxMsgSize)
		length, sender, err := g.gossipConn.ReadFromUDP(bytes)

		if err != nil{
			fmt.Println("Error")
		}

		if length > maxMsgSize {
			fmt.Println(
				"Message size", length, "is not a correct value the limit is", maxMsgSize,)
			continue
		}

		protobuf.Decode(bytes, packet)
		//gossiper.handleClientPacket(packet, sender)


		packet.PrintMessage(sender);
		g.printPeers()
		g.forwardGossipMessage(packet, sender)
	}

}


func (g *Gossiper) fwd(newPacket *P.GossipPacket, sourceAddr *net.UDPAddr){
	destinations := g.getNextPeers()
	for _, peer := range destinations{
		if(peer == sourceAddr.String()){
			continue
		}

		peerA, _ := net.ResolveUDPAddr(
			"udp4",
			peer,
		)

	  buf, err := protobuf.Encode(newPacket)

	  if err != nil {
	    fmt.Println("Error (encoding):", err.Error(), "with the packet :", newPacket)
	    return
	  }

	  //Try to send
		fmt.Println("MONGERING with", peer)
		size , err := g.gossipConn.WriteToUDP(buf, peerA)

		if err != nil {
			fmt.Println("Error (sending) :", err.Error(), "size :", size)

		}
	}
}

func (g *Gossiper) forwardClientMessage(packet *P.ClientPacket, sourceAddr *net.UDPAddr){
	newPacket := P.InitSimpleGossipPacket("", "", "")

	if g.simple {
		newPacket = P.InitSimpleGossipPacket(g.Name, g.gossipAddr.String(), packet.Simple.Contents)
	}else{
		newPacket = P.InitRumorGossipPacket(g.Name, g.countID, packet.Simple.Contents)
		g.countID += 1
		if g.addRumorInDB(newPacket){
			//fine
		}else{
			//error
		}
		g.updateStatus(newPacket.Rumor)
	}
	g.fwd(newPacket, sourceAddr)
}

func (g *Gossiper) forwardGossipMessage(packet *P.GossipPacket, sourceAddr *net.UDPAddr){
	//Add peers
	if(!U.ContainsString(g.peers, sourceAddr.String())){
		g.peers = append(g.peers, sourceAddr.String())
	}

	newPacket := P.InitSimpleGossipPacket("", "", "")

	if g.simple {
		//newPacket = P.InitSimpleGossipPacket(g.Name, g.gossipAddr.String(), packet.Simple.Contents)
		newPacket = P.InitSimpleGossipPacket(packet.Simple.OriginalName, g.gossipAddr.String(), packet.Simple.Contents)
	}else{
		newPacket = P.InitRumorGossipPacket(packet.Rumor.Origin, packet.Rumor.ID, packet.Rumor.Text)
		g.updateStatus(newPacket.Rumor)
	}

	g.fwd(newPacket, sourceAddr)
}

func (g *Gossiper) getNextPeers() []string{
	toto := []string{}

	if len(g.peers) == 0{
		return toto
	}

	if(!g.simple){
		if len(g.peers) == 1 && g.peers[0] == g.gossipAddr.String(){
			return toto
		}

		for{
			index := rand.Intn(len(g.peers))
			toto := g.peers[index:index+1]

			if toto[0] == g.gossipAddr.String(){
				continue
			}
			return toto
		}
	}else{
		return g.peers;
	}
}

func (g *Gossiper) getLastRumorFromOrigin(origin string) P.GossipPacket{
	return g.rumorDatabase[origin][len(g.rumorDatabase[origin]) - 1]
}

func (g *Gossiper) rumorDBIsEmptyForOrigin(origin string) bool{
	if _, ok := g.rumorDatabase[origin]; ok {
    return (len(g.rumorDatabase[origin]) == 0)
	}
	return false
}

func (g *Gossiper) sendAck(/*packet *P.GossipPacket,*/ sourceAddr *net.UDPAddr){

  //Open the connection
  //conn, _ := net.DialUDP("udp4", nil, sourceAddr)

	ack := P.InitStatusGossipPacket(g.status)
  buf, err := protobuf.Encode(ack)

  if err != nil {
    fmt.Println("Error (encoding):", err.Error(), "with the packet (ACK) :", ack)
    return
  }

  //Try to send
	size , err := g.gossipConn.WriteToUDP(buf, sourceAddr)
  //size , err := conn.Write(buf)

	if err != nil {
		fmt.Println("Error (sending) ACK :", err.Error(), "size :", size)
	}
}

func (g *Gossiper) originInDB(origin string) bool{
	if _, ok := g.rumorDatabase[origin]; ok {
	    return true
	}

	return false
}


func (g *Gossiper) addRumorInDB(packet *P.GossipPacket) bool{
	g.messages = append(g.messages, packet.Rumor.Origin + " : " + packet.Rumor.Text)
	if g.originInDB(packet.Rumor.Origin){
		if len(g.rumorDatabase[packet.Rumor.Origin]) == int(packet.Rumor.ID - 1){
			//test := g.rumorDatabase[packet.Rumor.Origin]
			g.rumorDatabase[packet.Rumor.Origin] = append(g.rumorDatabase[packet.Rumor.Origin], *packet)
			//test = test.append(test, *packet)
			return true
		}else{
			//Drop rumor : Waiting for previous ones
			return false
		}
	}else{
		if packet.Rumor.ID == 1{
			//Add rumor in DB
			g.rumorDatabase[packet.Rumor.Origin] = append(g.rumorDatabase[packet.Rumor.Origin], *packet)
			return true
		}else{
			//Drop rumor : Waiting for previous ones
			return false
		}
	}
}


func (g *Gossiper) updateStatus(r *M.RumorMessage){
	done := false
	index := -1

	//Check if status already exists
	for i, ps := range g.status{
		if r.Origin == ps.Identifier{
			done = true

			if(r.ID == ps.NextID){
				index = i
				//ps.NextID += 1
			}
		}
	}

	if done && index != -1{
		g.status[index].NextID += 1
	}

	//If new origin
	if !done{
		newS := S.InitPeerStatus(r.Origin, 2)
		if r.ID == 1{
			g.status = append(g.status, *newS)
		}else{
			///CHECK BEHAVIOR
			newS.NextID = 1
			g.status = append(g.status, *newS)
		}

	}
}

func (g *Gossiper) isUpToDate(packet *P.GossipPacket) bool{
	if len(packet.Status.Want) != len(g.status){
		return false
	}

	for _, ps := range packet.Status.Want{
		for _, is := range g.status {
			if ps.Identifier == is.Identifier {
				if ps.NextID != is.NextID{
					return false
				}
			}
		}
	}

	return true
}

func (g *Gossiper) handleStatusPacket(packet *P.GossipPacket, sourceAddr *net.UDPAddr){
	//Check if up to date
	if g.isUpToDate(packet){
		fmt.Println("IN SYNC WITH", sourceAddr.String())
		coin := rand.Int() % 2
		if coin == 1 {
			//select right Rumor
			//dest := g.getNextPeers()[0]
			if g.rumorDBIsEmptyForOrigin(g.Name){
				fmt.Println("Problem db is empty for key", g.Name)
			}else{
				newPacket := g.getLastRumorFromOrigin(g.Name)
				destinations := g.getNextPeers()
				for _, peer := range destinations{
					if(peer == sourceAddr.String()){
						continue
					}

					peerA, _ := net.ResolveUDPAddr(
						"udp4",
						peer,
					)

					fmt.Println("FLIPPED COIN sending rumor to ", peer)
					g.sendGPacket(newPacket, peerA)
				}
			}

			//g.fwd(&newPacket, sourceAddr)
		}
	}else{
		leave := false
		for _, ps := range packet.Status.Want{
			if leave {
				break
			}

			for _, is := range g.status {
				if is.Identifier > ps.Identifier{
					//Send messages from g to source
					leave = true
					g.sendMissingPackets(sourceAddr, ps)
					break
				}
			}
		}

		if !leave {
			leave = true

			for _, ps := range packet.Status.Want{
				if leave{
					break
				}

				for _, is := range g.status {
					if is.Identifier < ps.Identifier {
						//Send status to source but for all messages
						g.sendAck(sourceAddr)
						//Maybe optional
						leave = true
						break
					}
				}
			}
		}
	}
}

func (g *Gossiper) sendMissingPackets(dest *net.UDPAddr, ps S.PeerStatus){
	for _, gp := range g.rumorDatabase[ps.Identifier]{
		if gp.Rumor.ID == ps.NextID{
			g.sendGPacket(gp, dest)
			break
		}
	}
}

func (g *Gossiper) sendGPacket(packet P.GossipPacket, dest *net.UDPAddr){
  buf, err := protobuf.Encode(&packet)

  if err != nil {
    fmt.Println("Error (encoding):", err.Error(), "with the packet (rumor because miss packet from status) :", packet)
		fmt.Println("**********************DETAILS:", packet.Rumor)
    return
  }

  //Try to send
	fmt.Println("MONGERING with", dest.String())
	size , err := g.gossipConn.WriteToUDP(buf, dest)
  //size , err := conn.Write(buf)

	if err != nil {
		fmt.Println("Error (sending) (rumor because miss packet from status) :", err.Error(), "size :", size)
	}
}

func (g *Gossiper) present() {
	fmt.Println("Presentation of the gossiper :",g.Name)
}

func (g *Gossiper) printPeers() {
	peersString := "PEERS "
	for _, p := range g.peers{
		peersString += p +","
	}
	fmt.Println(peersString)
}

func (g *Gossiper) GetMessages() []string{
	return g.messages
}

func (g *Gossiper) GetPeers() []string{
	return g.peers
}

func (g *Gossiper) GetName() string{
	return g.Name
}
