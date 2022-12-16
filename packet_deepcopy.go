// generated by deep-copy -o packet_deepcopy.go --type Packet .; DO NOT EDIT.

package mpeg2ts

// DeepCopy generates a deep copy of Packet
func (o Packet) DeepCopy() Packet {
	var cp Packet = o
	if o.Data != nil {
		cp.Data = make([]byte, len(o.Data))
		copy(cp.Data, o.Data)
	}
	if o.AdaptationField.TransportPrivateData.Data != nil {
		cp.AdaptationField.TransportPrivateData.Data = make([]byte, len(o.AdaptationField.TransportPrivateData.Data))
		copy(cp.AdaptationField.TransportPrivateData.Data, o.AdaptationField.TransportPrivateData.Data)
	}
	if o.AdaptationField.Stuffing != nil {
		cp.AdaptationField.Stuffing = make([]byte, len(o.AdaptationField.Stuffing))
		copy(cp.AdaptationField.Stuffing, o.AdaptationField.Stuffing)
	}
	return cp
}