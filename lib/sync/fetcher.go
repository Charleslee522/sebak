package sync

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"time"

	"boscoin.io/sebak/lib/block"
	"boscoin.io/sebak/lib/network"
	"boscoin.io/sebak/lib/node"
	"boscoin.io/sebak/lib/node/runner"
	"boscoin.io/sebak/lib/storage"
	"boscoin.io/sebak/lib/transaction"
	"github.com/inconshreveable/log15"
)

type BlockFetcher struct {
	network           network.Network
	connectionManager network.ConnectionManager
	apiClient         Doer
	storage           *storage.LevelDBBackend
	localNode         *node.LocalNode

	fetchTimeout  time.Duration
	retryInterval time.Duration

	logger log15.Logger
}

type BlockFetcherOption = func(f *BlockFetcher)

func NewBlockFetcher(nw network.Network,
	cManager network.ConnectionManager,
	st *storage.LevelDBBackend,
	localNode *node.LocalNode,
	opts ...BlockFetcherOption) *BlockFetcher {

	f := &BlockFetcher{
		network:           nw,
		connectionManager: cManager,
		apiClient:         &http.Client{},
		storage:           st,
		localNode:         localNode,
		logger:            NopLogger(),

		fetchTimeout:  1 * time.Minute,
		retryInterval: 30 * time.Second,
	}

	for _, opt := range opts {
		opt(f)
	}

	return f
}

func (f *BlockFetcher) Fetch(ctx context.Context, syncInfo *SyncInfo) (*SyncInfo, error) {
	height := syncInfo.BlockHeight

	TryForever(func(attempt int) (bool, error) {
		select {
		case <-ctx.Done():
			return false, ctx.Err()
		default:
			f.logger.Debug("Try to fetch", "height", height, "attempt", attempt)
			if err := f.fetch(ctx, syncInfo); err != nil {
				if err == context.Canceled {
					return false, ctx.Err()
				}

				f.logger.Error(err.Error(), "err", err)
				c := time.After(f.retryInterval) //afterFunc?
				select {
				case <-ctx.Done():
					return false, ctx.Err()
				case <-c:
					return true, err
				}
			}
			return false, nil
		}
	})

	return syncInfo, nil
}

func (f *BlockFetcher) fetch(ctx context.Context, si *SyncInfo) error {
	var height = si.BlockHeight
	f.logger.Debug("Fetch start", "height", height)

	n := f.pickRandomNode()
	f.logger.Info(fmt.Sprintf("fetching items from node: %v", n), "fetching_node", n, "height", height)
	if n == nil {
		return errors.New("Fetch: node not found")
	}

	apiURL := apiClientURL(n, height)
	f.logger.Debug("apiClient", "url", apiURL.String())

	req, err := http.NewRequest("GET", apiURL.String(), nil)
	if err != nil {
		return err
	}

	ctx, cancelF := context.WithTimeout(ctx, f.fetchTimeout)
	defer cancelF()

	resp, err := f.apiClient.Do(req)
	defer func() {
		if resp != nil && resp.Body != nil {
			resp.Body.Close()
		}
	}()

	if err != nil {
		return err
	}
	if resp.StatusCode == http.StatusNotFound {
		//TODO:
		err := errors.New("Fetch: block not found")
		return err
	}

	items, err := f.unmarshalResp(resp.Body)
	if err != nil {
		return err
	}

	f.logger.Info("fetch get items", "items", len(items), "height", height)

	blocks, ok := items[runner.NodeItemBlock]
	if !ok || len(blocks) <= 0 {
		err := errors.New("Fetch: block not found in resp")
		return err
	}

	//TODO(anarcher): check items
	bts, ok := items[runner.NodeItemBlockTransaction]
	//ops, ok := items[runner.NodeItemBlockOperation]

	blk := blocks[0].(block.Block)
	si.Block = &blk

	for _, bt := range bts {
		bt, ok := bt.(block.BlockTransaction)
		if !ok {
			//TODO(anarcher): define sential error
			return errors.New("Invalid block transaction")
		}

		var tx transaction.Transaction
		if err := json.Unmarshal(bt.Message, &tx); err != nil {
			return err
		}

		si.Txs = append(si.Txs, &tx)
	}

	/*
		for _, op := range ops {
			op := op.(block.BlockOperation)
			info.Ops = append(info.Ops, &op)
		}
	*/

	return nil
}

// pickRandomNode choose one node by random. It is very protype for choosing fetching which node
func (f *BlockFetcher) pickRandomNode() node.Node {
	ac := f.connectionManager.AllConnected()
	if len(ac) <= 0 {
		return nil
	}

	var addressList []string
	for _, a := range ac {
		if f.localNode.Address() != a {
			addressList = append(addressList, a)
		}
	}

	if len(addressList) <= 0 {
		return nil
	}

	idx := rand.Intn(len(addressList))
	node := f.connectionManager.GetNode(addressList[idx])
	return node
}

func (f *BlockFetcher) existsBlockHeight(height uint64) bool {
	exists, err := block.ExistsBlockByHeight(f.storage, height)
	if err != nil {
		f.logger.Error("block.ExistsBlockByHeight", "err", err)
		return false
	}
	return exists
}

func (f *BlockFetcher) unmarshalResp(body io.ReadCloser) (map[runner.NodeItemDataType][]interface{}, error) {
	items := map[runner.NodeItemDataType][]interface{}{}

	sc := bufio.NewScanner(body)
	for sc.Scan() {
		itemType, b, err := runner.UnmarshalNodeItemResponse(sc.Bytes())
		if err != nil {
			return nil, err
		}
		items[itemType] = append(items[itemType], b)
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}

	return items, nil
}

func apiClientURL(n node.Node, height uint64) *url.URL {
	ep := n.Endpoint()
	u := url.URL(*ep)
	u.Path = network.UrlPathPrefixNode + runner.GetBlocksPattern
	q := u.Query()
	q.Set("height-range", fmt.Sprintf("%d-%d", height, height+1))
	q.Set("mode", "full")
	u.RawQuery = q.Encode()

	return &u
}