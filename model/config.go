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

	WXAuth struct {
		URL       string
		Publisher AppConf
		Worker    AppConf
		Pkey      string
		MchID     string
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
