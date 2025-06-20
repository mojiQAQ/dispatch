package model

import "git.ucloudadmin.com/unetworks/app/pkg/app"

type (
	HTTPS struct {
		Enable bool
		Port   uint
		Cert   string
		Key    string
	}

	AppConf struct {
		AppID  string
		Secret string
	}

	MchConf struct {
		CertSN     string
		APIV3Key   string
		PrivateKey string
		MchID      string
	}

	COS struct {
		APPID     string
		SecretID  string
		SecretKey string
		Bucket    string
		Region    string
	}

	WXAuth struct {
		URL       string
		Publisher AppConf
		Worker    AppConf
		Mch       MchConf
		COS       COS
	}

	Config struct {
		*app.ApplicationConfig
		HTTPSServer HTTPS
		WXAuth      WXAuth
		ImageBed    ImageBed
	}

	ImageBed struct {
		RelativePath string
		Path         string
	}
)
