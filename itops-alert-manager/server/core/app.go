package core

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-manager/server/common/log"
	"devops.aishu.cn/AISHUDevOps/AnyRobot/_git/itops-alert-manager/server/config"
	"github.com/gin-gonic/gin"
)

type HttpRouter interface {
	SetRouter(*gin.Engine)
}

type RouterQuote struct {
	Routes []HttpRouter
}

func InitHttpServer(routers ...HttpRouter) {
	gin.SetMode(config.Get().HttpServer.RunMode)
	ginEngine := gin.New()
	ginEngine.Use(gin.Logger(), gin.Recovery())
	for _, r := range routers {
		r.SetRouter(ginEngine)
	}
	gc := config.Get()
	httpServer := &http.Server{
		Handler:      ginEngine,
		Addr:         ":" + strconv.Itoa(gc.HttpServer.Addr),
		ReadTimeout:  gc.HttpServer.ReadTimeout * time.Second,
		WriteTimeout: gc.HttpServer.WriteTimeout * time.Second,
	}
	go func() {
		if err := httpServer.ListenAndServe(); err != nil {
			if err != http.ErrServerClosed {
				panic("srv.ListenAndServe err" + err.Error())
			}
			return
		}
	}()
	quit := make(chan os.Signal, 1)
	signal.Reset(syscall.SIGTERM, syscall.SIGINT)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit
	//关闭服务
	log.Info("Server Exiting")
	ctx2, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(ctx2); err != nil {
		log.Errorf("Server Shutdown: %v", err)
	}
}
