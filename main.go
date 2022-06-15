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
	midiDevice string
	ycToken    string
	folderID   string
	key1lang   string
	key2lang   string
)

func init() {
	pflag.StringVar(&midiDevice, "midi-device", "MidiStomp", "MIDI Device name")
	pflag.StringVar(&ycToken, "token", "", "Yandex Cloud OAuth token")
	pflag.StringVar(&folderID, "folder-id", "", "Yandex Cloud folder ID")
	pflag.StringVar(&key1lang, "key1-lang", "en-US", "Language to detect when first key is pressed")
	pflag.StringVar(&key2lang, "key2-lang", "ru-RU", "Language to detect when second key is pressed")
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
	control := make(chan string, 1)

	listener := NewMidiControllerHandler(midiDevice, func(keyIndex uint8, pressed bool) {
		var command = "" // Empty string is STOP
		if pressed {
			if keyIndex == 0 {
				command = key1lang
			} else {
				command = key2lang
			}
		}
		control <- command
	})
	defer listener.Close()

	launcher(control)
}

func launcher(control chan string) {
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
		lang := <-control // Language code is expected
		if lang == "" {
			continue
		}
		log.Infof("Starting, language: %v", lang)
		// =====================================================================
		sigFinish := make(chan os.Signal, 1)
		audioStream := make(chan []byte, 4)
		textStream := make(chan string, 4)

		go captureAudio(audioStream, sigFinish)
		go recognize(ycToken, folderID, lang, audioStream, textStream, sigFinish)
		log.Infof("Awaiting command to stop")
		// =====================================================================
		<-control // STOP("") is expected, but stop on any message
		// =====================================================================
		log.Infof("Stopping")
		sigFinish <- syscall.SIGINT
		typeKeys(kb, lang, textStream)
		log.Infof("Stopped")
	}
}
