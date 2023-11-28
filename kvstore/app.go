package main

import (
	"bytes"
	"log"

	"github.com/dgraph-io/badger/v3"
	abcitypes "github.com/cometbft/cometbft/abci/types"
)

//variavel KVStoreApplication com os metodos para comunicacao com a interface
type KVStoreApplication struct {
	db           *badger.DB
	onGoingBlock *badger.Txn
}

var _ abcitypes.Application = (*KVStoreApplication)(nil)

func NewKVStoreApplication(db *badger.DB) *KVStoreApplication {
	return &KVStoreApplication{db: db}
}

// metodo para ealizar a sincronizacao entre CometBFT e a aplicacao
func (app *KVStoreApplication) Info(info abcitypes.RequestInfo) abcitypes.ResponseInfo {
	return abcitypes.ResponseInfo{}
}

//requisitar leitura de informcoes sobre os valores-chaves da aplicacao
func (app *KVStoreApplication) Query(req abcitypes.RequestQuery) abcitypes.ResponseQuery {
	resp := abcitypes.ResponseQuery{Key: req.Data}

	//buscar valor da chave no banco de dados
	dbErr := app.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get(req.Data)
		if err != nil {
			if err != badger.ErrKeyNotFound {
				return err
			}
			resp.Log = "key does not exist"
			return nil
		}

		//retornar valor da chave se ela for achada
		return item.Value(func(val []byte) error {
			resp.Log = "exists"
			resp.Value = val
			return nil
		})
	})
	if dbErr != nil {
		log.Panicf("Error reading database, unable to execute query: %v", dbErr)
	}
	return resp
}

//validar formato da transacao
func (app *KVStoreApplication) isValid(tx []byte) uint32 {
	// check formato
	parts := bytes.Split(tx, []byte("="))
	if len(parts) != 2 {
		return 1
	}

	return 0
}

//mwtodo usado pelo CometBFT para validar transacoes -> invoca a funcao isValid 
func (app *KVStoreApplication) CheckTx(req abcitypes.RequestCheckTx) abcitypes.ResponseCheckTx {
	code := app.isValid(req.Tx)
	return abcitypes.ResponseCheckTx{Code: code}
}

//metodo usado pelo CometBFT para iniciar blockchain
func (app *KVStoreApplication) InitChain(chain abcitypes.RequestInitChain) abcitypes.ResponseInitChain {
	return abcitypes.ResponseInitChain{}
}

//metodo para preparar proposta para mudanca em cima de algum grupo de transacoes 
func (app *KVStoreApplication) PrepareProposal(proposal abcitypes.RequestPrepareProposal) abcitypes.ResponsePrepareProposal {
	return abcitypes.ResponsePrepareProposal{Txs: proposal.Txs}
}

//metodo para processar as propostas aceitas
func (app *KVStoreApplication) ProcessProposal(proposal abcitypes.RequestProcessProposal) abcitypes.ResponseProcessProposal {
	return abcitypes.ResponseProcessProposal{Status: abcitypes.ResponseProcessProposal_ACCEPT}
}

//metodo para iniciar bloco
func (app *KVStoreApplication) BeginBlock(req abcitypes.RequestBeginBlock) abcitypes.ResponseBeginBlock {
	app.onGoingBlock = app.db.NewTransaction(true)
	return abcitypes.ResponseBeginBlock{}
}

//metodo para enviar transacao
func (app *KVStoreApplication) DeliverTx(req abcitypes.RequestDeliverTx) abcitypes.ResponseDeliverTx {
	//valida novamente
	if code := app.isValid(req.Tx); code != 0 {
		return abcitypes.ResponseDeliverTx{Code: code}
	}

	parts := bytes.SplitN(req.Tx, []byte("="), 2)
	key, value := parts[0], parts[1]

	//enviar chave e valor
	if err := app.onGoingBlock.Set(key, value); err != nil {
		log.Panicf("Error writing to database, unable to execute tx: %v", err)
	}

	return abcitypes.ResponseDeliverTx{Code: 0}
}

//finalizar bloco
func (app *KVStoreApplication) EndBlock(block abcitypes.RequestEndBlock) abcitypes.ResponseEndBlock {
	return abcitypes.ResponseEndBlock{}
}

//comitar transacoes para o estado da aplicacao
func (app *KVStoreApplication) Commit() abcitypes.ResponseCommit {
	if err := app.onGoingBlock.Commit(); err != nil {
		log.Panicf("Error writing to database, unable to commit block: %v", err)
	}
	return abcitypes.ResponseCommit{Data: []byte{}}
}


//metodo usado para descobrir snapshots disposniveis
func (app *KVStoreApplication) ListSnapshots(snapshots abcitypes.RequestListSnapshots) abcitypes.ResponseListSnapshots {
	return abcitypes.ResponseListSnapshots{}
}

//metodo para oferecer snapshot ao aplicativo
func (app *KVStoreApplication) OfferSnapshot(snapshot abcitypes.RequestOfferSnapshot) abcitypes.ResponseOfferSnapshot {
	return abcitypes.ResponseOfferSnapshot{}
}

//metodo para recuperar pedacos de snapshots do aplicativo para enviar aos peers
func (app *KVStoreApplication) LoadSnapshotChunk(chunk abcitypes.RequestLoadSnapshotChunk) abcitypes.ResponseLoadSnapshotChunk {
	return abcitypes.ResponseLoadSnapshotChunk{}
}

//metodo para entregar pedacos de snapshots ao aplicativo
func (app *KVStoreApplication) ApplySnapshotChunk(chunk abcitypes.RequestApplySnapshotChunk) abcitypes.ResponseApplySnapshotChunk {
	return abcitypes.ResponseApplySnapshotChunk{}
}
