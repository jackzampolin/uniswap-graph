package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/machinebox/graphql"
	"golang.org/x/sync/errgroup"
)

type Config struct {
	Client *graphql.Client
}

func main() {
	t0 := time.Now()
	config := Config{graphql.NewClient("https://api.thegraph.com/subgraphs/name/uniswap/uniswap-v2")}

	out := &PairData{}
	eg, ctx := errgroup.WithContext(context.Background())
	for i := 0; i < 10; i++ {
		i := i
		eg.Go(func() error {
			ps, err := config.GetPairs(ctx, 100, i*100)
			if err != nil {
				return err
			}
			out.Lock()
			out.Pairs = append(out.Pairs, ps...)
			out.Unlock()
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		log.Fatal(err)
	}
	t1 := time.Now()
	fmt.Println("Got", len(out.Pairs), "pairs in", t1.Sub(t0))
}

// GetPairs returns the top N pairs from the Uniswap Subgraph
func (c Config) GetPairs(ctx context.Context, first, skip int) ([]UniswapPairs, error) {
	req := graphql.NewRequest(fmt.Sprintf(`{
		pairs(first: %d, skip: %d, orderBy: volumeUSD, orderDirection: desc) {
			id
			volumeUSD
			reserveUSD
			token0 {
				id
				name
				symbol
			}
			token1 {
				id
				name
				symbol
			}
		}
	}`, first, skip))
	out := &PairData{}
	err := c.Client.Run(ctx, req, out)
	return out.Pairs, err
}

type Token struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Symbol string `json:"symbol"`
}
type UniswapPairs struct {
	ID         string `json:"id"`
	ReserveUSD string `json:"reserveUSD"`
	Token0     Token  `json:"token0"`
	Token1     Token  `json:"token1"`
	VolumeUSD  string `json:"volumeUSD"`
}
type PairData struct {
	sync.Mutex
	Pairs []UniswapPairs `json:"pairs"`
}
