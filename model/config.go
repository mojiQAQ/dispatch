package model

import "git.ucloudadmin.com/unetworks/app/pkg/app"

type (
	HTTPS struct {
		Port uint
		Cert string
		Key  string
	}

	WXAuth struct {
		URL    string
		APPID  string
		Secret string
	}

	Config struct {
		*app.ApplicationConfig
		HTTPSServer HTTPS
		WXAuth      WXAuth
	}
)
