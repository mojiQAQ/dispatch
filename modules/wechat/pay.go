package wechat

import (
	"context"
	"github.com/wechatpay-apiv3/wechatpay-go/core"
	"github.com/wechatpay-apiv3/wechatpay-go/core/auth/verifiers"
	"github.com/wechatpay-apiv3/wechatpay-go/core/downloader"
	"github.com/wechatpay-apiv3/wechatpay-go/core/notify"
	"github.com/wechatpay-apiv3/wechatpay-go/core/option"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments"
	"github.com/wechatpay-apiv3/wechatpay-go/services/payments/jsapi"
	"github.com/wechatpay-apiv3/wechatpay-go/services/transferbatch"
	"github.com/wechatpay-apiv3/wechatpay-go/utils"
	"net/http"
)

func certLoader(mchID, mchCertificateSerialNumber, mchAPIv3Key, pKey string) (*notify.Handler, error) {
	mchPrivateKey, err := utils.LoadPrivateKeyWithPath(pKey)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	// 1. 使用 `RegisterDownloaderWithPrivateKey` 注册下载器
	err = downloader.MgrInstance().RegisterDownloaderWithPrivateKey(ctx, mchPrivateKey, mchCertificateSerialNumber, mchID, mchAPIv3Key)
	if err != nil {
		return nil, err
	}

	// 2. 获取商户号对应的微信支付平台证书访问器
	certificateVisitor := downloader.MgrInstance().GetCertificateVisitor(mchID)
	// 3. 使用证书访问器初始化 `notify.Handler`
	return notify.NewRSANotifyHandler(mchAPIv3Key, verifiers.NewSHA256WithRSAVerifier(certificateVisitor))
}

func loadPrivateKey(mchID, mchCertificateSerialNumber, mchAPIv3Key, pKey string) (*core.Client, error) {
	mchPrivateKey, err := utils.LoadPrivateKeyWithPath(pKey)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	// 使用商户私钥等初始化 client，并使它具有自动定时获取微信支付平台证书的能力
	opts := []core.ClientOption{
		option.WithWechatPayAutoAuthCipher(mchID, mchCertificateSerialNumber, mchPrivateKey, mchAPIv3Key),
	}
	return core.NewClient(ctx, opts...)
}

func (c *Ctl) CreateWechatPrePayOrder(openID, tradeNo, desc string, amount int64) (*jsapi.PrepayWithRequestPaymentResponse, error) {

	svc := jsapi.JsapiApiService{Client: c.wClient}
	resp, _, err := svc.PrepayWithRequestPayment(context.Background(), jsapi.PrepayRequest{
		Appid:       core.String(c.Conf.Publisher.AppID),
		Mchid:       core.String(c.Conf.Mch.MchID),
		Description: core.String(desc),
		OutTradeNo:  core.String(tradeNo),
		Attach:      core.String(tradeNo),
		NotifyUrl:   core.String("https://www.todistribute.cn:7443/dispatch/wechat_prepay_callback"),
		Amount: &jsapi.Amount{
			Total:    core.Int64(amount),
			Currency: core.String("CNY"),
		},
		Payer: &jsapi.Payer{
			Openid: core.String(openID),
		},
	})

	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *Ctl) PrepayCallback(req *http.Request) (*payments.Transaction, error) {

	transaction := new(payments.Transaction)
	_, err := c.handle.ParseNotifyRequest(context.Background(), req, transaction)
	if err != nil {
		return nil, err
	}

	return transaction, nil
}

func (c *Ctl) TransferToWorker(openid, tradeID, desc string, amount int64) (*transferbatch.InitiateBatchTransferResponse, error) {

	svc := transferbatch.TransferBatchApiService{Client: c.wClient}
	resp, _, err := svc.InitiateBatchTransfer(context.Background(), transferbatch.InitiateBatchTransferRequest{
		Appid:       core.String(c.Conf.Worker.AppID),
		OutBatchNo:  core.String(tradeID),
		BatchName:   core.String(desc),
		BatchRemark: core.String(desc),
		TotalAmount: core.Int64(amount),
		TotalNum:    core.Int64(1),
		TransferDetailList: []transferbatch.TransferDetailInput{
			{
				OutDetailNo:    core.String(tradeID),
				TransferAmount: core.Int64(amount),
				TransferRemark: core.String(desc),
				Openid:         core.String(openid),
			},
		},
		TransferSceneId: core.String("1001"),
	})
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *Ctl) CheckUnCompleteTransferOrder(tradeID string) (*transferbatch.TransferBatchEntity, error) {
	svc := transferbatch.TransferBatchApiService{Client: c.wClient}
	resp, _, err := svc.GetTransferBatchByOutNo(context.Background(),
		transferbatch.GetTransferBatchByOutNoRequest{
			OutBatchNo:      core.String(tradeID),
			NeedQueryDetail: core.Bool(false),
		},
	)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
