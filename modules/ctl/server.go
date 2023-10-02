package ctl

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"git.ucloudadmin.com/unetworks/app/pkg/app"
	"github.com/gin-gonic/gin"

	"github.com/mojiQAQ/dispatch/model"
	"github.com/mojiQAQ/dispatch/modules/order"
	"github.com/mojiQAQ/dispatch/modules/trade"
	"github.com/mojiQAQ/dispatch/modules/user"
)

type Server struct {
	*app.Application

	ctx    context.Context
	cancel context.CancelFunc
	cfg    *model.Config

	h  *gin.Engine
	OC *order.Ctl
	UC *user.Ctl
}

func NewServer() *Server {

	ctx, cancel := context.WithCancel(context.Background())
	srv := &Server{
		ctx:    ctx,
		cancel: cancel,
		Application: &app.Application{
			ConfPath: "cmd",
		},
		cfg: &model.Config{},
	}
	return srv
}

func (s *Server) Init() {

	s.Application.Init(s.cfg)
	s.AddAdminHandler()
	err := s.Application.InitDatabase()
	if err != nil {
		s.Panicf("failed init database, err=%s", err.Error())
		return
	}

	s.h = gin.Default()
	s.Http = &http.Server{Handler: s.h}

	s.UC = user.NewCtl(s.Logger, s.Database.Write, trade.NewCtl(), s.HttpClient(), s.cfg.WXAuth)
	s.OC = order.NewCtl(s.Logger, s.Database.Write, s.UC)

	s.InitRouter()
	return
}

func (s *Server) InitRouter() {
	s.OC.InitRouter(&s.h.RouterGroup)
	s.UC.InitRouter(&s.h.RouterGroup)
}

func (s *Server) Start() {

	err := s.Application.Start()
	if err != nil {
		panic(err)
	}

	go func() {
		err = s.h.RunTLS(fmt.Sprintf("%s:%d", s.LocalIPStr, s.cfg.HTTPSServer.Port),
			s.cfg.HTTPSServer.Cert, s.cfg.HTTPSServer.Key)
		if err != nil {
			panic(err)
		}
	}()
	s.OC.Start()
}

func (s *Server) Stop() {

	s.Application.Stop()
	s.cancel()
}

func (s *Server) Run() {

	exit := make(chan os.Signal, 1)
	signal.Notify(exit, syscall.SIGINT, syscall.SIGTERM)

	s.Init()
	s.Start()
	sig := <-exit
	s.Infof("receive signal %s, stopping server...", sig.String())
	s.Stop()
}
