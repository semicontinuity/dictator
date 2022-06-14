package main

import (
	"context"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	stt "github.com/yandex-cloud/go-genproto/yandex/cloud/ai/stt/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"io"
	"os"
)

// STTAPIEndpoint is used for all voice recognition requests
const STTAPIEndpoint = "stt.api.cloud.yandex.net:443"

// RecognitionClient wraps Recognizer_RecognizeStreamingClient
type RecognitionClient struct {
	conn     *grpc.ClientConn
	stt      stt.Recognizer_RecognizeStreamingClient
	iamToken string
	folder   string
}

func (sdk *SDK) NewRecognitionClient(ctx context.Context) (*RecognitionClient, error) {
	iamToken, err := sdk.IAMToken(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create IAM token")
	}

	conn, err := grpc.DialContext(ctx, STTAPIEndpoint,
		grpc.WithTransportCredentials(credentials.NewTLS(nil)),
		grpc.WithPerRPCCredentials(tokenAuth{token: iamToken, folderID: folderID}),
	)
	if err != nil {
		return nil, errors.Wrap(err, "unable to establish gRPC connection")
	}

	sttClient, err := stt.NewRecognizerClient(conn).RecognizeStreaming(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "unable to create STT service client")
	}

	return &RecognitionClient{conn: conn, stt: sttClient, folder: sdk.folder, iamToken: iamToken}, nil
}

func (rc *RecognitionClient) Close() error {
	return rc.conn.Close()
}

// =============================================================================================================

func (rc *RecognitionClient) sendAudio(
	lang string,
	chAudioChunks chan []byte,
	finish chan os.Signal,
) error {
	confReq := rc.NewConfigRequest(lang)
	if err := rc.Send(confReq); err != nil && err != io.EOF {
		return errors.Wrap(err, "Unable to send config request")
	}

	var audio []byte
	for {
		select {
		case <-finish:
			log.Infof("Finish")

			if err := rc.CloseSend(); err != nil {
				return errors.Wrap(err, "Unable to close recognition sending")
			}

			return nil

		case audio = <-chAudioChunks:
			log.Infof("Chunk")

			contentReq := rc.NewChunkRequest(audio)
			if err := rc.Send(contentReq); err != nil && err != io.EOF {
				return errors.Wrap(err, "Unable to send audio request")
			}
		}
	}
}

// =============================================================================================================

func (rc *RecognitionClient) receiveRecognitions(textStream chan string) error {
	// TODO:
	// these are remains of the example code;
	// no need to RecvAll, just push to textStream!

	sttResponses, err := rc.RecvAll()
	if err != nil {
		return errors.Wrap(err, "unable to receive recognition response")
	}

	for _, resp := range sttResponses {
		final := resp.GetFinal()
		if final != nil {
			log.Infof("Final")
			textStream <- final.GetAlternatives()[0].GetText()
		}
		// TODO: FinalRefinement?
	}

	return nil
}

// =============================================================================================================

// NewConfigRequest returns a properly set StreamingRequest for config
func (rc *RecognitionClient) NewConfigRequest(lang string) *stt.StreamingRequest {
	return &stt.StreamingRequest{
		Event: &stt.StreamingRequest_SessionOptions{
			SessionOptions: &stt.StreamingOptions{
				RecognitionModel: &stt.RecognitionModelOptions{
					AudioFormat: &stt.AudioFormatOptions{
						AudioFormat: &stt.AudioFormatOptions_RawAudio{
							RawAudio: &stt.RawAudio{
								SampleRateHertz:   48000,
								AudioEncoding:     stt.RawAudio_LINEAR16_PCM,
								AudioChannelCount: 1,
							},
						},
					},

					LanguageRestriction: &stt.LanguageRestrictionOptions{
						RestrictionType: stt.LanguageRestrictionOptions_WHITELIST,
						LanguageCode:    []string{lang},
					},
				},
			},
		},
	}
}

func (rc *RecognitionClient) NewChunkRequest(audio []byte) *stt.StreamingRequest {
	return &stt.StreamingRequest{
		Event: &stt.StreamingRequest_Chunk{
			Chunk: &stt.AudioChunk{
				Data: audio,
			},
		},
	}
}

func (rc *RecognitionClient) Send(req *stt.StreamingRequest) error {
	return rc.stt.Send(req)
}

func (rc *RecognitionClient) CloseSend() error {
	return rc.stt.CloseSend()
}

func (rc *RecognitionClient) Recv() (*stt.StreamingResponse, error) {
	return rc.stt.Recv()
}

// RecvAll accumulates all responses from the RecognitionClient
func (rc *RecognitionClient) RecvAll() ([]*stt.StreamingResponse, error) {
	result := make([]*stt.StreamingResponse, 0)
	for {
		resp, err := rc.Recv()
		switch err {
		case nil:
			result = append(result, resp)
		case io.EOF:
			return result, nil
		default:
			return nil, err
		}
	}
}
