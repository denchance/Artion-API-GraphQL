package resolvers

import (
	"artion-api-graphql/internal/repository"
	"artion-api-graphql/internal/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"math/big"
)

// Collection object is constructed from query, data from db are loaded on demand into "dbCollection" field.
type Collection struct {
	Address      common.Address
	dbCollection *types.Collection // data for collection loaded from Mongo
}

type CollectionEdge struct {
	Node *Collection
}

func (edge CollectionEdge) Cursor() (types.Cursor, error) {
	return types.CursorFromId(edge.Node.Address.Bytes()), nil
}

type CollectionConnection struct {
	Edges      []CollectionEdge
	TotalCount hexutil.Big
	PageInfo   PageInfo
}

func NewCollectionConnection(list *types.CollectionList) (con *CollectionConnection, err error) {
	con = new(CollectionConnection)
	con.TotalCount = (hexutil.Big)(*big.NewInt(list.TotalCount))
	con.Edges = make([]CollectionEdge, len(list.Collection))
	for i := 0; i < len(list.Collection); i++ {
		resolverCollection := Collection{
			Address:      list.Collection[i].Address,
			dbCollection: list.Collection[i],
		}
		con.Edges[i].Node = &resolverCollection
	}
	con.PageInfo.HasNextPage = list.HasNext
	con.PageInfo.HasPreviousPage = list.HasPrev
	if len(list.Collection) > 0 {
		startCur, err := con.Edges[0].Cursor()
		if err != nil {
			return nil, err
		}
		endCur, err := con.Edges[len(con.Edges)-1].Cursor()
		if err != nil {
			return nil, err
		}
		con.PageInfo.StartCursor = &startCur
		con.PageInfo.EndCursor = &endCur
	}
	return con, err
}


func (t *Collection) load() error {
	if t.dbCollection == nil {
		tok, err := repository.R().GetCollection(t.Address)
		if err != nil {
			return err
		}
		t.dbCollection = tok
	}
	return nil
}

func (t Collection) Type() (string, error) {
	err := t.load()
	if err != nil {
		return "", err
	}
	return t.dbCollection.Type, nil
}

func (t Collection) Name() (string, error) {
	err := t.load()
	if err != nil {
		return "", err
	}
	return t.dbCollection.Name, nil
}

func (t Collection) Symbol() (string, error) {
	err := t.load()
	if err != nil {
		return "", err
	}
	return t.dbCollection.Symbol, nil
}

func (t Collection) Created() (types.Time, error) {
	err := t.load()
	if err != nil {
		return types.Time{}, err
	}
	return types.Time(t.dbCollection.Created), nil
}

func (t Collection) IsActive() (bool, error) {
	err := t.load()
	if err != nil {
		return false, err
	}
	return t.dbCollection.IsActive, nil
}

func (t *Collection) Tokens(args struct{ PaginationInput }) (con *TokenConnection, err error) {
	cursor, count, backward, err := args.ToRepositoryInput()
	if err != nil {
		return nil, err
	}
	list, err := repository.R().ListCollectionTokens(t.Address, cursor, count, backward)
	if err != nil {
		return nil, err
	}
	return NewTokenConnection(list)
}