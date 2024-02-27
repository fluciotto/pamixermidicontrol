package korg

import (
	"slices"

	"github.com/samber/lo"
)

func MidiDataToData(midiData []byte) []byte {
	data := []byte{}
	chunks := lo.Chunk(midiData, 8)
	for _, chunk := range chunks {
		for i := range chunk[:len(chunk)-1] {
			data = append(data, (chunk[0]&(0x1<<i))<<(7-i)+chunk[i+1])
		}
	}
	return data
}

func DataToMidiData(data []byte) []byte {
	midiData := []byte{}
	chunks := lo.Chunk(midiData, 7)
	for _, chunk := range chunks {
		var msbs byte
		var lsbs []byte
		for i, bt := range chunk {
			msbs += bt & 0x80 >> (7 - i)
			lsbs = append(lsbs, bt&0x7f)
		}
		midiData = append(midiData, msbs)
		midiData = slices.Concat(midiData, lsbs)
	}
	return midiData
}
