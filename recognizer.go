package main

import (
	"context"
	log "github.com/sirupsen/logrus"
	"os"
)

func recognize(ycToken string, folderID string, lang string, audioStream chan []byte, textStream chan string, sigFinish chan os.Signal) {
	defer close(textStream)

	ctx := context.Background()
	sdk, err := /*yandex.*/ NewSDK(ctx, ycToken, folderID)
	if err != nil {
		log.Fatalf("Unable to create: %v", err)
	}
	defer sdk.Close()

	rClient, err := sdk.NewRecognitionClient(ctx)
	if err != nil {
		log.Fatalf("Unable to create STT client: %v", err)
	}
	defer rClient.Close()

	go func() {
		err := rClient.sendAudio(lang, audioStream, sigFinish)
		if err != nil {
			log.Warnf("Unable to send audio: %v", err)
		}
	}()

	err = rClient.receiveRecognitions(textStream)
	if err != nil {
		log.Warnf("Unable to recognize audio: %v", err)
	}
}
