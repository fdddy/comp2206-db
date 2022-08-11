package db_test

import (
	"context"
	"net"
	"testing"
	"time"

	db "comp2206-db"

	"google.golang.org/grpc"
)

func TestDB(t *testing.T) {
	serverCfg, err := db.ReadServerConfig("configs/server.json")
	if err != nil {
		t.Error(err)
	}
	server, err := db.NewServer(serverCfg)
	if err != nil {
		t.Error(err)
	}
	go func() {
		lis, _ := net.Listen("tcp", "127.0.0.1:15184")
		s := grpc.NewServer()
		db.RegisterServerServiceServer(s, server)
		s.Serve(lis)
	}()
	time.Sleep(5 * time.Second)
	ownerCfg := &db.OwnerConfig{
		StorePath:  "tmp/owner",
		SetList:    []string{"A", "B", "C", "D"},
		ServerAddr: "127.0.0.1:15184",
	}
	owner, err := db.NewOwner(ownerCfg)
	if err != nil {
		t.Error(err)
	}
	dK, err := owner.DelegateKeys("B")
	if err != nil {
		t.Error(err)
	}
	searcherCfg := &db.SearcherConfig{
		SetList:    []string{"A", "B", "C", "D"},
		Keys:       dK,
		ServerAddr: "127.0.0.1:15184",
	}
	searcher, err := db.NewSearcher(searcherCfg)
	if err != nil {
		t.Error(err)
	}
	err = owner.Insert(context.Background(), []*db.Record{
		{
			UserId:   "PB222",
			Location: "YYY",
			Time:     time.Now(),
			Set:      "B",
		},
		{
			UserId:   "PB444",
			Location: "YYY",
			Time:     time.Now(),
			Set:      "A",
		},
		{
			UserId:   "PB464",
			Location: "YYY",
			Time:     time.Now(),
			Set:      "B",
		},
	})
	if err != nil {
		panic(err)
	}
	timeA := time.Now().Add(-5 * time.Second)
	timeB := time.Now().Add(5 * time.Second)
	res, err := searcher.FindB(context.Background(), "B", "YYY", timeA, timeB)
	if err != nil {
		t.Error(err)
	}
	t.Log(res[1].UserId)
	res, err = owner.FindB(context.Background(), "B", "YYY", timeA, timeB)
	if err != nil {
		t.Error(err)
	}
	t.Log(res[1].UserId)
	time.Sleep(4)
}
