package main

import (
	"context"
	"github.com/pkg/errors"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/iam/v1"
	ycsdk "github.com/yandex-cloud/go-sdk"
)

type SDK struct {
	ctx      context.Context
	token    string
	folder   string
	internal *ycsdk.SDK
}

func NewSDK(ctx context.Context, token, folder string) (*SDK, error) {
	sdk, err := ycsdk.Build(ctx, ycsdk.Config{
		Credentials: ycsdk.OAuthToken(token),
	})
	if err != nil {
		return nil, errors.Wrap(err, "unable to build Yandex SDK")
	}

	return &SDK{ctx: ctx, token: token, folder: folder, internal: sdk}, nil
}

func (sdk *SDK) Close() error {
	return sdk.internal.Shutdown(sdk.ctx)
}

func (sdk *SDK) IAMToken(ctx context.Context) (string, error) {
	iamResp, err := sdk.internal.IAM().IamToken().Create(ctx, &iam.CreateIamTokenRequest{
		Identity: &iam.CreateIamTokenRequest_YandexPassportOauthToken{YandexPassportOauthToken: sdk.token},
	})
	if err != nil {
		return "", errors.Wrap(err, "unable to make request for IAM token creation")
	}

	return iamResp.GetIamToken(), nil
}
