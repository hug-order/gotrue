package sms_provider

import (
	"errors"
	"fmt"
	"strings"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	dysmsapi "github.com/alibabacloud-go/dysmsapi-20170525/v3/client"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
	"github.com/supabase/auth/internal/conf"
)

type AliSmsProvider struct {
	Config *conf.AliSmsProviderConfiguration
	Client *dysmsapi.Client
	codes  map[string]struct{}
}

type AliSmsResponseRecipients struct {
	TotalSentCount int `json:"totalSentCount"`
}

type AliSmsResponse struct {
	Code      string `json:"Code"`
	Message   string `json:"Message"`
	BizId     string `json:"BizId"`
	RequestId string `json:"RequestId"`
}

// Creates a SmsProvider with the AliSms Config
func NewAliSmsProvider(config conf.AliSmsProviderConfiguration) (SmsProvider, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	codes := map[string]struct{}{}
	for _, code := range strings.Split(config.Codes, ",") {
		codes[code] = struct{}{}
	}

	client, err := dysmsapi.NewClient(&openapi.Config{
		Endpoint:        tea.String("dysmsapi.aliyuncs.com"),
		AccessKeyId:     tea.String(config.AccessKey),
		AccessKeySecret: tea.String(config.AccessSecret),
	})

	if err != nil {
		return nil, err
	}

	return &AliSmsProvider{&config, client, codes}, nil
}

func (t *AliSmsProvider) SendMessage(phone, message, channel, otp string) (string, error) {
	switch channel {
	case SMSProvider:
		return t.SendSms(phone, message, otp)
	default:
		return "", fmt.Errorf("channel type %q is not supported for AliSms", channel)
	}
}

// Send an SMS containing the OTP with AliSms's API
func (t *AliSmsProvider) SendSms(phone, message, code string) (string, error) {
	if _, ok := t.codes[code]; !ok {
		return "", errors.New("Invalid ali sms code!")
	}

	sendSmsRequest := &dysmsapi.SendSmsRequest{
		SignName:      tea.String(t.Config.SignName),
		TemplateCode:  tea.String(code),
		PhoneNumbers:  tea.String(phone),
		TemplateParam: tea.String(message),
	}

	resp, err := t.Client.SendSmsWithOptions(sendSmsRequest, &util.RuntimeOptions{})
	if err != nil {
		return "", err
	}
	if resp.Body.Code != nil && *resp.Body.Code != "ok" {
		return "", errors.New(*resp.Body.Message)
	}

	return *resp.Body.BizId, nil
}
