package model

import "git.ucloudadmin.com/unetworks/app/pkg/app"

type (
	HTTPS struct {
		Enable bool
		Port   uint
		Cert   string
		Key    string
	}

	WXAuth struct {
		URL    string
		APPID  string
		Secret string
		Pkey   string
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
