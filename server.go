package db

import (
	"context"
	"encoding/json"
	"os"

	"github.com/plzfgme/mfast"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Server struct {
	UnimplementedServerServiceServer
	iServer    *mfast.Server
	mDB        *mongo.Database
	mCtx       context.Context
	mCtxCancel context.CancelFunc
	config     *ServerConfig
}

type ServerConfig struct {
	StorePath     string `json:"store_path,omitempty"`
	MongoDBConfig struct {
		URI            string `json:"uri,omitempty"`
		DBName         string `json:"db_name,omitempty"`
		CollectionName string `json:"collection_name,omitempty"`
	} `json:"mongo_db_config,omitempty"`
}

func ReadServerConfig(path string) (*ServerConfig, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	config := &ServerConfig{}
	err = json.Unmarshal(file, config)
	if err != nil {
		return nil, err
	}
	return config, nil
}

func NewServer(config *ServerConfig) (*Server, error) {
	iServerConfig := &mfast.ServerConfig{
		StorePath: config.StorePath,
	}
	iServer, err := mfast.NewServer(iServerConfig)
	if err != nil {
		return nil, err
	}
	mCtx, mCtxCancel := context.WithCancel(context.Background())
	mClient, err := mongo.Connect(mCtx, options.Client().ApplyURI(config.MongoDBConfig.URI))
	if err != nil {
		iServer.Close()
		mCtxCancel()
		return nil, err
	}
	mDB := mClient.Database(config.MongoDBConfig.DBName)

	return &Server{
		iServer:    iServer,
		mDB:        mDB,
		mCtx:       mCtx,
		mCtxCancel: mCtxCancel,
		config:     config,
	}, nil
}

func (server *Server) Close() error {
	server.mCtxCancel()
	return server.iServer.Close()
}

func (server *Server) Insert(ctx context.Context, q *InsertQuery) (*InsertResponse, error) {
	encodedDocs := q.GetDocs()
	docs := make([]interface{}, len(encodedDocs))
	for i, v := range encodedDocs {
		docs[i] = bson.M{}
		bson.Unmarshal(v.GetBinary(), &docs[i])
	}
	_, err := server.mDB.Collection(server.config.MongoDBConfig.CollectionName).InsertMany(ctx, docs)
	if err != nil {
		return &InsertResponse{
			Msg: "Failed to insert documents",
		}, err
	}
	encodedTkns := q.GetTkns()
	tkns := make([]*mfast.UpdateToken, len(encodedTkns))
	for i, v := range encodedTkns {
		tkns[i] = &mfast.UpdateToken{}
		json.Unmarshal(v.GetBinary(), tkns[i])
	}
	// TODO
	for _, tkn := range tkns {
		err := server.iServer.Update(tkn)
		if err != nil {
			return &InsertResponse{
				Msg: err.Error(),
			}, nil
		}
	}

	return &InsertResponse{
		Msg: "Ok",
	}, nil
}

func (server *Server) PreFind(ctx context.Context, q *PreFindQuery) (*PreFindResponse, error) {
	resultTkns := make([]*Token, len(q.GetTkns()))
	for i, tkn := range q.GetTkns() {
		preSearchTkn := &mfast.PreSearchToken{}
		err := json.Unmarshal([]byte(string(tkn.GetBinary())), preSearchTkn)
		if err != nil {
			return &PreFindResponse{
				Msg: "Failed to unmarshal token",
			}, err
		}
		preSearchResult, err := server.iServer.PreSearch(preSearchTkn)
		if err != nil {
			return &PreFindResponse{
				Msg: err.Error(),
			}, err
		}
		if preSearchResult == nil {
			resultTkns[i] = &Token{
				Binary: nil,
			}
			continue
		}
		encodedResult, err := json.Marshal(preSearchResult)
		if err != nil {
			return &PreFindResponse{
				Msg: err.Error(),
			}, err
		}
		resultTkns[i] = &Token{
			Binary: encodedResult,
		}
	}

	return &PreFindResponse{
		Msg:  "Ok",
		Tkns: resultTkns,
	}, nil
}

func (server *Server) Find(ctx context.Context, q *FindQuery) (*FindResponse, error) {
	encodedTkns := q.GetTkns()
	tkns := make([]*mfast.SearchToken, len(encodedTkns))
	for i, v := range encodedTkns {
		tkns[i] = &mfast.SearchToken{}
		bson.Unmarshal(v.GetBinary(), tkns[i])
	}
	// TODO
	ids := make([]primitive.ObjectID, 0)
	for _, tkn := range tkns {
		rawPartIds, err := server.iServer.Search(tkn)
		if err != nil {
			return &FindResponse{
				Msg: "Failed to search index",
			}, err
		}
		for _, rawId := range rawPartIds {
			id, err := primitive.ObjectIDFromHex(rawId)
			if err != nil {
				return &FindResponse{
					Msg: "Failed to parse id",
				}, err
			}
			ids = append(ids, id)
		}
	}
	mQ := bson.M{"_id": bson.M{"$in": ids}}
	project := bson.M{}
	for _, field := range q.GetFields() {
		project[field] = 1
	}
	cur, err := server.mDB.Collection(server.config.MongoDBConfig.CollectionName).Find(ctx, mQ, options.Find().SetProjection(project))
	if err != nil {
		return &FindResponse{
			Msg: "Failed to retrive documents",
		}, err
	}
	docs := make([]bson.M, 0)
	cur.All(ctx, &docs)
	encodedDocs := make([]*Document, len(docs))
	for i, doc := range docs {
		encodedDoc, err := bson.Marshal(doc)
		if err != nil {
			return &FindResponse{
				Msg: "Failed to encode documents",
			}, err
		}
		encodedDocs[i] = &Document{
			Binary: encodedDoc,
		}
	}

	return &FindResponse{
		Msg:  "Ok",
		Docs: encodedDocs,
	}, err
}
