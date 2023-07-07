package main

import (
	"match_helper/handler"
	pb "match_helper/proto"

	"github.com/micro/micro/v3/service"
	"github.com/micro/micro/v3/service/broker"
	"github.com/micro/micro/v3/service/logger"
)

func main() {
	// Create service
	srv := service.New(
		service.Name("match_helper"),
		service.Version("latest"),
	)

	// Register handler
	ggg := new(handler.Match_helper)
	pb.RegisterMatchHelperHandler(srv.Server(), ggg)

	go func() {
		broker.Subscribe("match_result", ggg.HandlerMsg)
	}()

	// Run service
	if err := srv.Run(); err != nil {
		logger.Fatal(err)
	}
}
