package wechat

import (
	"context"

	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/core/option"
	"github.com/wechatpay-apiv3/wechatpay-go/services/certificates"
	"github.com/wechatpay-apiv3/wechatpay-go/utils"
)

func (c *Ctl) loadPrivateKey(mchID, mchCertificateSerialNumber, mchAPIv3Key string) {
	mchPrivateKey, err := utils.LoadPrivateKeyWithPath(c.Conf.Pkey)
	if err != nil {
		c.Errorf("load privatekey failed, err=%v", err)
		return
	}

	ctx := context.Background()
	// 使用商户私钥等初始化 client，并使它具有自动定时获取微信支付平台证书的能力
	opts := []core.ClientOption{
		option.WithWechatPayAutoAuthCipher(mchID, mchCertificateSerialNumber, mchPrivateKey, mchAPIv3Key),
	}
	client, err := core.NewClient(ctx, opts...)
	if err != nil {
		c.Errorf("new wechat pay client err:%s", err)
		return
	}

	// 发送请求，以下载微信支付平台证书为例
	// https://pay.weixin.qq.com/wiki/doc/apiv3/wechatpay/wechatpay5_1.shtml
	svc := certificates.CertificatesApiService{Client: client}
	resp, result, err := svc.DownloadCertificates(ctx)
	c.Debugf("status=%d resp=%s", result.Response.StatusCode, resp)
}
