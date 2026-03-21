package core

import (
	"MixFound/global"
	"MixFound/searcher"
	"MixFound/searcher/words"
	"MixFound/web/controller"
	"MixFound/web/router"
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func NewContainer(tokenizer *words.Tokenizer) *searcher.Container {
	container := &searcher.Container{
		Dir:       global.CONFIG.Data,
		Debug:     global.CONFIG.Debug,
		Shard:     global.CONFIG.Shard,
		Timeout:   global.CONFIG.Timeout,
		BufferNum: global.CONFIG.BufferNum,
		Tokenizer: tokenizer,
	}
	if err := container.Init(); err != nil {
		panic(err)
	}

	return container
}

func NewTokenizer(dictionaryPath string) *words.Tokenizer {
	return words.NewTokenizer(dictionaryPath)
}

func Initialize() {
	//加载配置
	global.CONFIG = Parse()

	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("panic:%s\n", r)
		}
	}()

	//加载词典目录，初始化分词器
	tokenizer := NewTokenizer(global.CONFIG.Dictionary)
	//初始化容器
	global.Container = NewContainer(tokenizer)

	//初始化业务逻辑
	controller.NewServices()

	//注册路由
	r := router.SetupRouter()

	//启动服务
	srv := &http.Server{
		Addr:    global.CONFIG.Addr,
		Handler: r,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	//优雅关机
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Println("Server shutdown:", err)
	}

	log.Println("Server exiting")
}
