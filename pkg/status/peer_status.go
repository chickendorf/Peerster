package status

type PeerStatus struct {
  Identifier string
  NextID uint32
}

func InitPeerStatus(
	identifier string,
	id uint32,
) *PeerStatus {
  ret := PeerStatus{Identifier : identifier, NextID : id};
  return &ret;
}
