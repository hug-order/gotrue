package sms_provider

import (
	"errors"
	"fmt"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	dysmsapi "github.com/alibabacloud-go/dysmsapi-20170525/v3/client"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/supabase/auth/internal/conf"
)

type AliSmsProvider struct {
	Config *conf.AliSmsProviderConfiguration
	Client *dysmsapi.Client
}

// Creates a SmsProvider with the AliSms Config
func NewAliSmsProvider(config conf.AliSmsProviderConfiguration) (SmsProvider, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	client, err := dysmsapi.NewClient(&openapi.Config{
		Endpoint:        tea.String("dysmsapi.aliyuncs.com"),
		AccessKeyId:     tea.String(config.AccessKey),
		AccessKeySecret: tea.String(config.AccessSecret),
	})

	if err != nil {
		return nil, err
	}

	return &AliSmsProvider{&config, client}, nil
}

func (t *AliSmsProvider) SendMessage(phone, message, channel, otp string) (string, error) {
	switch channel {
	case SMSProvider:
		return t.SendSms(phone, otp)
	default:
		return "", fmt.Errorf("channel type %q is not supported for AliSms", channel)
	}
}

// Send an SMS containing the OTP with AliSms's API
func (t *AliSmsProvider) SendSms(phone, code string) (string, error) {
	sendSmsRequest := &dysmsapi.SendSmsRequest{
		SignName:      tea.String(t.Config.SignName),
		TemplateCode:  tea.String(t.Config.Code),
		PhoneNumbers:  tea.String(phone),
		TemplateParam: tea.String(fmt.Sprintf("{\"code\":\"%s\"}", code)),
	}

	resp, err := t.Client.SendSmsWithOptions(sendSmsRequest, &util.RuntimeOptions{})
	if err != nil {
		return "", err
	}
	if resp.Body.Code != nil && *resp.Body.Code != "OK" {
		return "", errors.New(*resp.Body.Message)
	}

	return *resp.Body.BizId, nil
}
