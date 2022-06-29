package db

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"time"

	"github.com/fentec-project/gofe/abe"
	"github.com/plzfgme/mfast"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Searcher struct {
	mfastSearcher *mfast.Searcher
	aBE           *abe.FAME
	keys          *DelegatedKeys
	conn          *grpc.ClientConn
	client        ServerServiceClient
	config        *SearcherConfig
}

type SearcherConfig struct {
	SetList    []string
	Keys       *DelegatedKeys
	ServerAddr string
}

func ReadSearcherConfig(path string) (*SearcherConfig, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	config := &SearcherConfig{}
	err = json.Unmarshal(file, config)
	if err != nil {
		return nil, err
	}
	return config, nil
}

func NewSearcher(config *SearcherConfig) (*Searcher, error) {
	mfastConfig := &mfast.SearcherConfig{
		SetList: config.SetList,
		Keys:    config.Keys.MFastKeys,
	}
	mfastSearcher := mfast.NewSearcher(mfastConfig)
	aBE := abe.NewFAME()
	conn, err := grpc.Dial(config.ServerAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	client := NewServerServiceClient(conn)

	return &Searcher{
		mfastSearcher: mfastSearcher,
		aBE:           aBE,
		keys:          config.Keys,
		conn:          conn,
		client:        client,
		config:        config,
	}, nil
}

func (searcher *Searcher) FindA(ctx context.Context, set string, userId string, timeA, timeB time.Time) ([]*FindAResult, error) {
	tKWs := getBRCKWs(uint64(timeA.Unix()), uint64(timeB.Unix()))
	preFindTkns := make([]*Token, len(tKWs))
	kws := make([][]byte, len(tKWs))
	for i, tKW := range tKWs {
		kws[i] = []byte("A:" + userId + ":" + tKW)
		b, err := json.Marshal(searcher.mfastSearcher.GenPreSearchTkn(set, kws[i]))
		if err != nil {
			return nil, err
		}
		preFindTkns[i] = &Token{
			Binary: b,
		}
	}
	preFindRes, err := searcher.client.PreFind(ctx, &PreFindQuery{
		Tkns: preFindTkns,
	})
	if err != nil {
		return nil, err
	} else if preFindRes.GetMsg() != "Ok" {
		return nil, errors.New((preFindRes.GetMsg()))
	}
	findTkns := make([]*Token, 0)
	preResTkns := preFindRes.GetTkns()
	for i := range preResTkns {
		if preResTkns[i].GetBinary() == nil {
			continue
		}
		preSearchRes := &abe.FAMECipher{}
		err := json.Unmarshal(preResTkns[i].GetBinary(), preSearchRes)
		if err != nil {
			return nil, err
		}
		searchTkn, err := searcher.mfastSearcher.GenSearchTkn(set, kws[i], preSearchRes)
		if err != nil {
			return nil, err
		}
		b, err := bson.Marshal(searchTkn)
		if err != nil {
			return nil, err
		}
		findTkns = append(findTkns, &Token{Binary: b})
	}
	findQ := &FindQuery{
		Fields: []string{"Location", "Time"},
		Tkns:   findTkns,
	}
	findRes, err := searcher.client.Find(ctx, findQ)
	if err != nil {
		return nil, err
	} else if findRes.GetMsg() != "Ok" {
		return nil, errors.New(findRes.GetMsg())
	}
	bDocs := findRes.GetDocs()
	results := make([]*FindAResult, len(bDocs))
	for i, bDoc := range bDocs {
		doc := bson.M{}
		err := bson.Unmarshal(bDoc.GetBinary(), doc)
		if err != nil {
			return nil, err
		}
		cLoc := &abe.FAMECipher{}
		json.Unmarshal(doc["Location"].(primitive.Binary).Data, cLoc)
		loc, _ := searcher.aBE.Decrypt(cLoc, searcher.keys.ABEAttrK, searcher.keys.ABEPK)
		cTime := &abe.FAMECipher{}
		json.Unmarshal(doc["Time"].(primitive.Binary).Data, cTime)
		sTime, _ := searcher.aBE.Decrypt(cTime, searcher.keys.ABEAttrK, searcher.keys.ABEPK)
		t, err := time.Parse(time.RFC1123, sTime)
		if err != nil {
			return nil, err
		}

		results[i] = &FindAResult{
			Location: loc,
			Time:     t,
		}
	}
	return results, nil
}

func (searcher *Searcher) FindB(ctx context.Context, set string, loc string, timeA, timeB time.Time) ([]*FindBResult, error) {
	tKWs := getBRCKWs(uint64(timeA.Unix()), uint64(timeB.Unix()))
	preFindTkns := make([]*Token, len(tKWs))
	kws := make([][]byte, len(tKWs))
	for i, tKW := range tKWs {
		kws[i] = []byte("B:" + loc + ":" + tKW)
		b, err := json.Marshal(searcher.mfastSearcher.GenPreSearchTkn(set, kws[i]))
		if err != nil {
			return nil, err
		}
		preFindTkns[i] = &Token{
			Binary: b,
		}
	}
	preFindRes, err := searcher.client.PreFind(ctx, &PreFindQuery{
		Tkns: preFindTkns,
	})
	if err != nil {
		return nil, err
	} else if preFindRes.GetMsg() != "Ok" {
		return nil, errors.New((preFindRes.GetMsg()))
	}
	findTkns := make([]*Token, 0)
	preResTkns := preFindRes.GetTkns()
	for i := range preResTkns {
		if preResTkns[i].GetBinary() == nil {
			continue
		}
		preSearchRes := &abe.FAMECipher{}
		err := json.Unmarshal(preResTkns[i].GetBinary(), preSearchRes)
		if err != nil {
			return nil, err
		}
		searchTkn, err := searcher.mfastSearcher.GenSearchTkn(set, kws[i], preSearchRes)
		if err != nil {
			return nil, err
		}
		b, err := bson.Marshal(searchTkn)
		if err != nil {
			return nil, err
		}
		findTkns = append(findTkns, &Token{Binary: b})
	}
	findQ := &FindQuery{
		Fields: []string{"UserId"},
		Tkns:   findTkns,
	}
	findRes, err := searcher.client.Find(ctx, findQ)
	if err != nil {
		return nil, err
	} else if findRes.GetMsg() != "Ok" {
		return nil, errors.New(findRes.GetMsg())
	}
	bDocs := findRes.GetDocs()
	results := make([]*FindBResult, len(bDocs))
	for i, bDoc := range bDocs {
		doc := bson.M{}
		err := bson.Unmarshal(bDoc.GetBinary(), doc)
		if err != nil {
			return nil, err
		}
		cUserId := &abe.FAMECipher{}
		json.Unmarshal(doc["UserId"].(primitive.Binary).Data, cUserId)
		userId, _ := searcher.aBE.Decrypt(cUserId, searcher.keys.ABEAttrK, searcher.keys.ABEPK)

		results[i] = &FindBResult{
			UserId: userId,
		}
	}
	return results, nil
}
