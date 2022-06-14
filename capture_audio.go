package main

import (
	"github.com/gordonklaus/portaudio"
	"os"
	"unsafe"
)

func captureAudio(chunks chan []byte, sig_finish chan os.Signal) os.Signal {
	portaudio.Initialize()
	defer portaudio.Terminate()
	in := make([]int16, 4000)
	stream, err := portaudio.OpenDefaultStream(1, 0, 48000, len(in), in)
	chk(err)
	defer stream.Close()

	var s os.Signal = nil
	chk(stream.Start())
	for s == nil {
		chk(stream.Read())

		p := unsafe.Pointer(&in[0])
		//goland:noinspection GoRedundantConversion
		data := unsafe.Slice((*byte)(p), 8000)
		dataCopy := make([]byte, 8000)
		copy(dataCopy, data)
		chunks <- dataCopy

		select {
		case s = <-sig_finish:
		default:
		}
	}
	chk(stream.Stop())
	return s
}

func chk(err error) {
	if err != nil {
		panic(any(err))
	}
}
