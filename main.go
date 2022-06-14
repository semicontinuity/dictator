package main

import (
	kbd "github.com/micmonay/keybd_event"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"os"
	"runtime"
	"syscall"
	"time"
)

var (
	midiDevice    string
	ycToken       string
	folderID      string
	lang          string
)

func init() {
	pflag.StringVar(&midiDevice, "midi-device", "MidiStomp", "MIDI Device name")
	pflag.StringVar(&ycToken, "token", "", "Yandex Cloud OAuth token")
	pflag.StringVar(&folderID, "folder-id", "", "Yandex Cloud folder ID")
	pflag.StringVar(&lang, "lang", "en-US", "Language to detect")
	level := pflag.String("log-level", "INFO", "Logrus log level (DEBUG, WARN, etc.)")
	pflag.Parse()

	logLevel, err := log.ParseLevel(*level)
	if err != nil {
		log.Fatalf("Unknown log level: %s", *level)
	}
	log.SetFormatter(&log.TextFormatter{FullTimestamp: true})
	log.SetLevel(logLevel)

	if ycToken == "" || folderID == "" || midiDevice == "" {
		pflag.PrintDefaults()
		os.Exit(1)
	}
}

func main() {
	control := make(chan bool, 1)

	listener := NewPedalListener(midiDevice, func(pressed bool) {
        control <- pressed
    })
    defer listener.Close()

    launcher(control)
}


func launcher(control chan bool) {
	kb, err := kbd.NewKeyBonding()
	if err != nil {
		panic(any(err))
	}
	// For linux, it is VERY important to wait 2 seconds
	if runtime.GOOS == "linux" {
		time.Sleep(2 * time.Second)
	}

	for {
		// =====================================================================
		cmd1 := <-control	// START(true) is expected
		if !cmd1 { continue }
		// =====================================================================
		sigFinish := make(chan os.Signal, 1)
		audioStream := make(chan []byte, 4)
		textStream := make(chan string, 4)

		go captureAudio(audioStream, sigFinish)
		go recognize(ycToken, folderID, lang, audioStream, textStream, sigFinish)
		log.Infof("Awaiting command to stop")
		// =====================================================================
		<-control	// STOP(false) is expected, but stop anyway
		// =====================================================================
		log.Infof("Stopping")
		sigFinish <- syscall.SIGINT
		typeKeys(kb, textStream)
		log.Infof("Stopped")
	}
}
