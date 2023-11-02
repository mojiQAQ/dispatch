package wechat

import (
	"encoding/json"
	"fmt"
	"git.ucloudadmin.com/unetworks/app/pkg/httpclient"
	"git.ucloudadmin.com/unetworks/app/pkg/log"
	"github.com/mojiQAQ/dispatch/model"
	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/core/notify"
	"net/http"
)

type (
	Ctl struct {
		*log.Logger
		Conf   model.WXAuth
		client *httpclient.HttpClient

		token   *AccessToken
		wClient *core.Client
		handle  *notify.Handler
	}

	ErrInfo struct {
		ErrMsg  string `json:"errmsg"`
		ErrCode int32  `json:"errcode"`
	}

	AuthKey struct {
		*ErrInfo
		SessionKey string `json:"session_key"`
		UnionID    string `json:"unionid"`
		OpenID     string `json:"openid"`
	}

	AccessToken struct {
		*ErrInfo
		AccessToken string `json:"access_token"`
		ExpiresIn   uint   `json:"expires_in"`
	}

	watermark struct {
		Timestamp int    `json:"timestamp"`
		Appid     string `json:"appid"`
	}

	PhoneInfo struct {
		PhoneNumber     string    `json:"phoneNumber"`
		PurePhoneNumber string    `json:"purePhoneNumber"`
		CountryCode     string    `json:"countryCode"`
		Watermark       watermark `json:"watermark"`
	}

	PhoneResp struct {
		*ErrInfo
		PhoneInfo *PhoneInfo `json:"phone_info"`
	}
)

func NewCtl(logger *log.Logger, client *httpclient.HttpClient, cfg model.WXAuth) *Ctl {
	ctl := &Ctl{
		Logger: logger,
		client: client,
		Conf:   cfg,
	}

	wClient, err := loadPrivateKey(cfg.Mch.MchID, cfg.Mch.CertSN, cfg.Mch.APIV3Key, cfg.Mch.PrivateKey)
	if err != nil {
		panic(err)
	}

	handle, err := certLoader(cfg.Mch.MchID, cfg.Mch.CertSN, cfg.Mch.APIV3Key, cfg.Mch.PrivateKey)
	if err != nil {
		panic(err)
	}

	ctl.wClient = wClient
	ctl.handle = handle
	return ctl
}

func (c *Ctl) GetAuthKey(code string, role model.Role) (*AuthKey, error) {

	appid := ""
	secret := ""
	switch role {
	case model.RolePublisher:
		appid = c.Conf.Publisher.AppID
		secret = c.Conf.Publisher.Secret
	case model.RoleWorker:
		appid = c.Conf.Worker.AppID
		secret = c.Conf.Worker.Secret
	}

	url := fmt.Sprintf("%s/sns/jscode2session?appid=%s&secret=%s&grant_type=authorization_code&js_code=%s",
		c.Conf.URL, appid, secret, code)

	c.Debugf("login request: %s", url)
	resp := &httpclient.HttpResp{}
	resp, err := c.client.Get(map[string]string{"Content-Type": "application/json"}, url)
	if err != nil {
		return nil, err
	}

	authkey := &AuthKey{
		ErrInfo: &ErrInfo{},
	}
	err = json.Unmarshal(resp.Body, authkey)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK || authkey.ErrCode != 0 {
		return nil, fmt.Errorf(authkey.ErrMsg)
	}

	return authkey, nil
}

func (c *Ctl) getAccessToken(role model.Role) (*AccessToken, error) {

	appid := ""
	secret := ""
	switch role {
	case model.RolePublisher:
		appid = c.Conf.Publisher.AppID
		secret = c.Conf.Publisher.Secret
	case model.RoleWorker:
		appid = c.Conf.Worker.AppID
		secret = c.Conf.Worker.Secret
	}

	url := fmt.Sprintf("%s/cgi-bin/token?appid=%s&secret=%s&grant_type=client_credential",
		c.Conf.URL, appid, secret)

	c.Debugf("get access_token request: %s", url)
	resp := &httpclient.HttpResp{}
	resp, err := c.client.Get(map[string]string{"Content-Type": "application/json"}, url)
	if err != nil {
		return nil, err
	}

	token := &AccessToken{
		ErrInfo: &ErrInfo{},
	}
	err = json.Unmarshal(resp.Body, token)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK || token.ErrCode != 0 {
		return nil, fmt.Errorf(token.ErrMsg)
	}

	return token, nil
}

func (c *Ctl) GetAccessToken(role model.Role) (*AccessToken, error) {

	if c.token != nil {
		return c.token, nil
	}

	t, err := c.getAccessToken(role)
	if err != nil {
		return nil, err
	}

	return t, nil
}

func (c *Ctl) GetPhoneNumber(code string, role model.Role) (*PhoneInfo, error) {

	token, err := c.GetAccessToken(role)
	if err != nil {
		return nil, nil
	}

	url := fmt.Sprintf("%s/wxa/business/getuserphonenumber?access_token=%s",
		c.Conf.URL, token.AccessToken)

	req := map[string]string{
		"code": code,
	}

	resp := &PhoneResp{}
	err = c.client.PostJson(url, req, resp)
	if err != nil {
		return nil, err
	}

	if token.ErrCode != 0 {
		return nil, fmt.Errorf(token.ErrMsg)
	}

	return resp.PhoneInfo, nil
}
