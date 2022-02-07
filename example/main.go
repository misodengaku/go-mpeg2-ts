package main

import (
	"fmt"
	"os"
	"time"

	mpeg2ts "github.com/misodengaku/go-mpeg2-ts"
)

const BUFSIZE = 188

func main() {
	file, err := os.Open("test.ts")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	var fsize int64

	if fi, err := file.Stat(); err == nil {
		fsize = fi.Size()
	}

	mpeg2 := mpeg2ts.Mpeg2TS{}.New(fsize / 188)

	buf := make([]byte, BUFSIZE)
	i := 0
	for {
		// fmt.Println(i)
		n, err := file.Read(buf)
		if err != nil {
			// Readエラー処理
			if err.Error() == "EOF" {

			} else {
				fmt.Println("errbreak")
				fmt.Println(err)
			}
			break
		}
		if n == 0 {
			fmt.Println("0break")
			time.Sleep(100 * time.Millisecond)
			break
		}

		mpeg2.Packets[i].Load(buf, i)
		mpeg2.Packets[i].ParseHeader()
		// go func() {

		// }()
		// go func(index int) {
		// 	mpeg2.Packets[index].ParseHeader()
		// }(i)

		i++
	}
	for i = 0; i < 100; i++ {
		//fmt.Printf("%#v\r\n", mpeg2.Packets[i].GetHeader())
		//fmt.Printf("%#v\r\n", mpeg2.Packets[i].GetPayload())
		// if mpeg2.Packets[i].ParseHeader() == nil {
		fmt.Printf("%d %x %t %t %t %d %d %d %d\r\n",
			i,
			mpeg2.Packets[i].SyncByte,
			mpeg2.Packets[i].TransportErrorIndicator,
			mpeg2.Packets[i].PayloadUnitStartIndicator,
			mpeg2.Packets[i].TransportPriorityIndicator,
			mpeg2.Packets[i].PID,
			mpeg2.Packets[i].TransportScrambleControl,
			mpeg2.Packets[i].AdaptationFieldControl,
			mpeg2.Packets[i].ContinuityCheckIndex)
		if mpeg2.Packets[i].AdaptationField != nil {
			fmt.Printf("\tAdaptationField dump: %d %t %t %t %t %t %t %t %t\r\n",
				mpeg2.Packets[i].AdaptationField.Size,
				mpeg2.Packets[i].AdaptationField.DiscontinuityIndicator,
				mpeg2.Packets[i].AdaptationField.RandomAccessIndicator,
				mpeg2.Packets[i].AdaptationField.ESPriorityIndicator,
				mpeg2.Packets[i].AdaptationField.PCRFlag,
				mpeg2.Packets[i].AdaptationField.OPCRFlag,
				mpeg2.Packets[i].AdaptationField.SplicingPointFlag,
				mpeg2.Packets[i].AdaptationField.TransportPrivateDataFlag,
				mpeg2.Packets[i].AdaptationField.ExtensionFlag)
		}
		// }

	}
	fmt.Println("Continuity check")
	if err := mpeg2.CheckStream(); err != nil {
		fmt.Println(err.Error())
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
