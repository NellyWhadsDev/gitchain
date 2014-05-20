package api

import (
	"encoding/hex"
	"errors"
	"log"
	"net/http"

	"github.com/gitchain/gitchain/router"
	"github.com/gitchain/gitchain/transaction"
	"github.com/gitchain/gitchain/util"
)

type NameService struct{}

type NameReservationArgs struct {
	Alias string
	Name  string
}

type NameReservationReply struct {
	Id     string
	Random string
}

func (NameService) NameReservation(r *http.Request, args *NameReservationArgs, reply *NameReservationReply) error {
	key, err := srv.DB.GetKey(args.Alias)
	if err != nil {
		return err
	}
	if key == nil {
		return errors.New("can't find the key")
	}
	tx, random := transaction.NewNameReservation(args.Name)

	hash, err := srv.DB.GetPreviousEnvelopeHashForPublicKey(&key.PublicKey)
	if err != nil {
		log.Printf("Error while preparing transaction: %v", err)
	}
	txe := transaction.NewEnvelope(hash, tx)
	txe.Sign(key)

	reply.Id = hex.EncodeToString(txe.Hash())
	reply.Random = hex.EncodeToString(random)
	// We save sha(random+name)=txhash to scraps to be able to find
	// the transaction hash by random and number during allocation
	srv.DB.PutScrap(util.SHA256(append(random, []byte(args.Name)...)), txe.Hash())
	router.Send("/transaction", make(chan *transaction.Envelope), txe)
	return nil
}

type NameAllocationArgs struct {
	Alias  string
	Name   string
	Random string
}

type NameAllocationReply struct {
	Id string
}

func (NameService) NameAllocation(r *http.Request, args *NameAllocationArgs, reply *NameAllocationReply) error {
	key, err := srv.DB.GetKey(args.Alias)
	if err != nil {
		return err
	}
	if key == nil {
		return errors.New("can't find the key")
	}
	random, err := hex.DecodeString(args.Random)
	if err != nil {
		return err
	}
	tx, err := transaction.NewNameAllocation(args.Name, random)
	if err != nil {
		return err
	}

	hash, err := srv.DB.GetPreviousEnvelopeHashForPublicKey(&key.PublicKey)
	if err != nil {
		log.Printf("Error while preparing transaction: %v", err)
	}

	txe := transaction.NewEnvelope(hash, tx)
	txe.Sign(key)

	reply.Id = hex.EncodeToString(txe.Hash())
	router.Send("/transaction", make(chan *transaction.Envelope), txe)
	return nil
}
