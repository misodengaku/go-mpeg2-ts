package main

import (
	"fmt"

	mpeg2ts "github.com/misodengaku/go-mpeg2-ts"
)

const BUFSIZE = 188

func main() {
	mpeg2, err := mpeg2ts.LoadStandardTS("d443e813-631c-42b5-a25c-6b40558e4477_2022-02-02_055000.h264_gpac.ts")
	if err != nil {
		panic(err)
	}

	for i, p := range mpeg2.Packets {
		//fmt.Printf("%#v\r\n", p.GetHeader())
		//fmt.Printf("%#v\r\n", p.GetPayload())
		// if p.ParseHeader() == nil {
		fmt.Printf("%d sync:%x tei:%t pusi:%t tpi:%t pid:%x tsc:%d afc:%d cci:%d\r\n",
			i,
			p.SyncByte,
			p.TransportErrorIndicator,
			p.PayloadUnitStartIndicator,
			p.TransportPriorityIndicator,
			p.PID,
			p.TransportScrambleControl,
			p.AdaptationFieldControl,
			p.ContinuityCheckIndex)
		if p.HasAdaptationField() {
			fmt.Printf("\tAdaptationField dump: size:%d di:%t rai:%t espi:%t pcr:%t opcr:%t spf:%t tpdf:%t ef:%t\r\n",
				p.AdaptationField.Size,
				p.AdaptationField.DiscontinuityIndicator,
				p.AdaptationField.RandomAccessIndicator,
				p.AdaptationField.ESPriorityIndicator,
				p.AdaptationField.PCRFlag,
				p.AdaptationField.OPCRFlag,
				p.AdaptationField.SplicingPointFlag,
				p.AdaptationField.TransportPrivateDataFlag,
				p.AdaptationField.ExtensionFlag)
		}
		// }
		if i == 100 {
			break
		}

	}
	fmt.Println("Continuity check")
	if cr := mpeg2.CheckStream(); cr.DropCount > 0 {
		fmt.Println("frame drop detected!!")
		for _, v := range cr.DropList {
			fmt.Printf("frame index: %d\n", v.Index)
		}
	} else {
		fmt.Println("OK")
	}

	// go func() {
	// 	pat := mpeg2.PIDFilter(mpeg2ts.PID_PAT)
	// 	for _, p := range pat.Packets {
	// 		// fmt.Println("PAT frame:", p.Index, p.PID, p.Data)
	// 		patx, _ := p.ParsePAT()
	// 		fmt.Printf("%#v\r\n", patx)
	// 	}
	// }()
	// var ch chan struct{}
	// go func() {
	pmt := mpeg2.PIDFilter(mpeg2ts.PID_EIT)
	for _, p := range pmt.Packets {
		fmt.Printf("%#v\r\n", p)
	}
	// 	ch <- struct{}{}
	// }()
	// <-ch
}
