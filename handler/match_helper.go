package handler

import (
	"context"
	"encoding/csv"
	"fmt"
	match_evaluator "match_evaluator/proto"
	"math/rand"
	"os"
	"sync"
	"sync/atomic"
	"time"

	match_frontend "github.com/askldfhjg/match_apis/match_frontend/proto"
	"google.golang.org/protobuf/proto"

	match_helper "match_helper/proto"

	"github.com/micro/micro/v3/service/broker"
	"github.com/micro/micro/v3/service/client"
	"github.com/micro/micro/v3/service/logger"
)

type Match_helper struct {
	userRet map[string]int
}

var UserMap *sync.Map
var UserList []string
var Index uint64

func Init() {
	file, err := os.Open("uuids.csv")
	if err != nil {
		return
	}
	defer file.Close()
	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return
	}
	UserMap = &sync.Map{}
	for _, record := range records {
		uuid := record[0]
		UserMap.Store(uuid, 0)
		UserList = append(UserList, uuid)
	}
}

// Call is a single request handler called via client.Call or the generated client code
func (e *Match_helper) FillData(ctx context.Context, req *match_helper.FillDataReq, rsp *match_helper.FillDataRsp) error {
	if len(UserList) <= 0 {
		Init()
	}
	if req.Tt <= 0 {
		rsp.Name = "tt error"
		return nil
	}
	e.userRet = map[string]int{}
	if req.Count <= 0 {
		rsp.Name = "count error"
		return nil
	}
	if int(req.Count) > len(UserList) {
		rsp.Name = fmt.Sprintf("%d:%d", req.Count, len(UserList))
		return nil
	}
	go loop(int(req.Count), int(req.Tt))
	return nil
}

func (e *Match_helper) HandlerMsg(raw *broker.Message) error {
	msg := &match_evaluator.MatchDetail{}
	err := proto.Unmarshal(raw.Body, msg)
	if err != nil {
		return err
	}
	cc := 0
	for _, bb := range msg.Ids {
		if bb == "robot" {
			continue
		}
		e.userRet[bb]++
		if e.userRet[bb] > 1 {
			cc = 1
		}
	}
	logger.Infof("fffff %d %d", len(e.userRet), int(cc))
	logger.Infof("%+v", msg.Ids)
	return nil
}

func loop(count int, tt int) {
	ticket := time.NewTicker(time.Second * 1)
	defer ticket.Stop()
	cc := 0
	perCount := count / tt
	for {
		select {
		case <-ticket.C:
			cc++
			go send(perCount)
			if cc >= tt {
				ticket.Stop()
				break
			}
		}
	}
}

func send(count int) {
	for i := 0; i < count; i++ {
		newIndex := atomic.AddUint64(&Index, 1)
		uuid := UserList[int(newIndex)%len(UserList)]
		matchSrv := match_frontend.NewMatchFrontendService("match_frontend", client.DefaultClient)
		rr := &match_frontend.EnterMatchReq{
			Param: &match_frontend.MatchInfo{
				PlayerId: uuid,
				Score:    int64(rand.Intn(100000) + 1),
				GameId:   "aaaa",
				SubType:  1,
				Status:   match_frontend.MatchStatus_idle,
			},
		}
		_, err := matchSrv.EnterMatch(context.Background(), rr)
		if err != nil {
			logger.Infof("EnterMatch error %+v", err)
		} else {
			UserMap.Store(uuid, 0)
			//logger.Infof("EnterMatch result %+v", rspp)
		}
	}
	logger.Infof("finish %d", count)
}
