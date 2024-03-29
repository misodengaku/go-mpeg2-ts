// generated by deep-copy -o pes_deepcopy.go --type PES .; DO NOT EDIT.

package mpeg2ts

// DeepCopy generates a deep copy of PES
func (o PES) DeepCopy() PES {
	var cp PES = o
	if o.ElementaryStream != nil {
		cp.ElementaryStream = make([]byte, len(o.ElementaryStream))
		copy(cp.ElementaryStream, o.ElementaryStream)
	}
	if o.PacketDataStream != nil {
		cp.PacketDataStream = make([]byte, len(o.PacketDataStream))
		copy(cp.PacketDataStream, o.PacketDataStream)
	}
	if o.Padding != nil {
		cp.Padding = make([]byte, len(o.Padding))
		copy(cp.Padding, o.Padding)
	}
	return cp
}
