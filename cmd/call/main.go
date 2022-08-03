package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"geth/contract"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
	"github.com/pkg/errors"
	"io/ioutil"
	"math/big"
	"net/http"
)

const (
	expBase = 10
)

type Account struct {
	Nonce   string `json:"nonce"`
	Code    string `json:"code"`
	Balance string `json:"balance"`
}

type OverrideAccounts map[common.Address]Account

type roundTripperExt struct {
	c          *http.Client
	appendData json.RawMessage
}

type reqMessage struct {
	JSONRPC string            `json:"jsonrpc"`
	ID      int               `json:"id"`
	Method  string            `json:"method"`
	Params  []json.RawMessage `json:"params"`
}

func main() {
	var (
		SimSwapAddress = common.HexToAddress("0x1111111111111111111111111111111111111101")
		MyWallet       = common.HexToAddress("0x54F42802c6B381682c48B530f234De0a896Af1DD")
		CommonContract = OverrideAccounts{
			SimSwapAddress: {
				Nonce: "0x10",
				Code:  "0x6080604052600436106100295760003560e01c80637e5465ba1461002e578063db31da581461006b575b600080fd5b34801561003a57600080fd5b506100556004803603810190610050919061055e565b61009b565b6040516100629190610883565b60405180910390f35b6100856004803603810190610080919061059e565b61015a565b6040516100929190610815565b60405180910390f35b6000808390508073ffffffffffffffffffffffffffffffffffffffff1663095ea7b3847f80000000000000000000000000000000000000000000000000000000000000006040518363ffffffff1660e01b81526004016100fc9291906107ec565b602060405180830381600087803b15801561011657600080fd5b505af115801561012a573d6000803e3d6000fd5b505050506040513d601f19601f8201168201806040525081019061014e919061066f565b50600091505092915050565b6060600073c02aaa39b223fe8d0a0e5c4f27ead9083c756cc290506000737a250d5630b4cf539739df2c5dacb4c659f2488d90503073ffffffffffffffffffffffffffffffffffffffff16637e5465ba89886040518363ffffffff1660e01b81526004016101c99291906107c3565b602060405180830381600087803b1580156101e357600080fd5b505af11580156101f7573d6000803e3d6000fd5b505050506040513d601f19601f8201168201806040525081019061021b919061069c565b50600060644261022b9190610944565b90506000600267ffffffffffffffff81111561024a57610249610ac8565b5b6040519080825280602002602001820160405280156102785781602001602082028036833780820191505090505b50905083816000815181106102905761028f610a99565b5b602002602001019073ffffffffffffffffffffffffffffffffffffffff16908173ffffffffffffffffffffffffffffffffffffffff168152505089816001815181106102df576102de610a99565b5b602002602001019073ffffffffffffffffffffffffffffffffffffffff16908173ffffffffffffffffffffffffffffffffffffffff168152505060008390508073ffffffffffffffffffffffffffffffffffffffff16637ff36ab53460008533886040518663ffffffff1660e01b815260040161035f9493929190610837565b6000604051808303818588803b15801561037857600080fd5b505af115801561038c573d6000803e3d6000fd5b50505050506040513d6000823e3d601f19601f820116820180604052508101906103b69190610626565b5061010067ffffffffffffffff8111156103d3576103d2610ac8565b5b6040519080825280601f01601f1916602001820160405280156104055781602001600182028036833780820191505090505b509550505050505095945050505050565b6000610429610424846108c3565b61089e565b9050808382526020820190508285602086028201111561044c5761044b610b01565b5b60005b8581101561047c57816104628882610549565b84526020840193506020830192505060018101905061044f565b5050509392505050565b60008135905061049581610b21565b92915050565b600082601f8301126104b0576104af610afc565b5b81516104c0848260208601610416565b91505092915050565b6000815190506104d881610b38565b92915050565b60008083601f8401126104f4576104f3610afc565b5b8235905067ffffffffffffffff81111561051157610510610af7565b5b60208301915083600182028301111561052d5761052c610b01565b5b9250929050565b60008135905061054381610b4f565b92915050565b60008151905061055881610b4f565b92915050565b6000806040838503121561057557610574610b0b565b5b600061058385828601610486565b925050602061059485828601610486565b9150509250929050565b6000806000806000608086880312156105ba576105b9610b0b565b5b60006105c888828901610486565b95505060206105d988828901610534565b94505060406105ea88828901610486565b935050606086013567ffffffffffffffff81111561060b5761060a610b06565b5b610617888289016104de565b92509250509295509295909350565b60006020828403121561063c5761063b610b0b565b5b600082015167ffffffffffffffff81111561065a57610659610b06565b5b6106668482850161049b565b91505092915050565b60006020828403121561068557610684610b0b565b5b6000610693848285016104c9565b91505092915050565b6000602082840312156106b2576106b1610b0b565b5b60006106c084828501610549565b91505092915050565b60006106d583836106e1565b60208301905092915050565b6106ea8161099a565b82525050565b6106f98161099a565b82525050565b600061070a826108ff565b6107148185610922565b935061071f836108ef565b8060005b8381101561075057815161073788826106c9565b975061074283610915565b925050600181019050610723565b5085935050505092915050565b60006107688261090a565b6107728185610933565b9350610782818560208601610a06565b61078b81610b10565b840191505092915050565b61079f816109e2565b82525050565b6107ae816109f4565b82525050565b6107bd816109d8565b82525050565b60006040820190506107d860008301856106f0565b6107e560208301846106f0565b9392505050565b600060408201905061080160008301856106f0565b61080e60208301846107a5565b9392505050565b6000602082019050818103600083015261082f818461075d565b905092915050565b600060808201905061084c6000830187610796565b818103602083015261085e81866106ff565b905061086d60408301856106f0565b61087a60608301846107b4565b95945050505050565b600060208201905061089860008301846107b4565b92915050565b60006108a86108b9565b90506108b48282610a39565b919050565b6000604051905090565b600067ffffffffffffffff8211156108de576108dd610ac8565b5b602082029050602081019050919050565b6000819050602082019050919050565b600081519050919050565b600081519050919050565b6000602082019050919050565b600082825260208201905092915050565b600082825260208201905092915050565b600061094f826109d8565b915061095a836109d8565b9250827fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff0382111561098f5761098e610a6a565b5b828201905092915050565b60006109a5826109b8565b9050919050565b60008115159050919050565b600073ffffffffffffffffffffffffffffffffffffffff82169050919050565b6000819050919050565b60006109ed826109d8565b9050919050565b60006109ff826109d8565b9050919050565b60005b83811015610a24578082015181840152602081019050610a09565b83811115610a33576000848401525b50505050565b610a4282610b10565b810181811067ffffffffffffffff82111715610a6157610a60610ac8565b5b80604052505050565b7f4e487b7100000000000000000000000000000000000000000000000000000000600052601160045260246000fd5b7f4e487b7100000000000000000000000000000000000000000000000000000000600052603260045260246000fd5b7f4e487b7100000000000000000000000000000000000000000000000000000000600052604160045260246000fd5b600080fd5b600080fd5b600080fd5b600080fd5b600080fd5b6000601f19601f8301169050919050565b610b2a8161099a565b8114610b3557600080fd5b50565b610b41816109ac565b8114610b4c57600080fd5b50565b610b58816109d8565b8114610b6357600080fd5b5056fea2646970667358221220b86811084880a2096cc00b1201c8fb24021cc7d7b5852dbb7cc068d7207d2a3964736f6c63430008070033",
			},
			MyWallet: {
				Balance: "0x8ac7230489e80000",
			},
		}
	)
	// https://etherscan.io/tx/0x606e8c8084855d3fb20cb1c69f520d0a1feae6c35a9d3659a9cda8a1cf53e9e2#eventlog
	//rawurl := "https://proxy.kyberengineering.io/ethereum" // "http://localhost:8545/" //  "https://mainnet.infura.io/v3/3d85e3bded764846bc25e1ca36f73b91" // "https://proxy.kyberengineering.io/ethereum"
	rawurl := "http://localhost:8545/"
	client, err := ethclient.Dial(rawurl)
	if err != nil {
		panic(err)
	}

	block, err := client.BlockNumber(context.Background())
	if err != nil {
		panic(err)
	}
	fmt.Println("Block", block)

	// Generate EncodedSwapData
	simClient, err := NewClient(rawurl, SimSwapAddress, CommonContract)
	if err != nil {
		panic(err)
	}
	ab, err := abi.JSON(bytes.NewBufferString(contract.SimCallMetaData.ABI))
	if err != nil {
		panic(err)
	}
	ksRouterAddress := common.HexToAddress("0x00555513Acf282B42882420E5e5bA87b44D8fA6E")
	// kncAddress := common.HexToAddress("0xdeFA4e8a7bcBA345F687a2f1456F5Edd9CE97202")
	usdcAddress := common.HexToAddress("0xa0b86991c6218b36c1d19d4a2e9eb0ce3606eb48")
	data, err := ab.Pack("simswap", usdcAddress,
		FloatToTokenAmount(1000, 18),
		ksRouterAddress,
		[]byte{})
	if err != nil {
		panic(err)
	}
	myWallet := common.HexToAddress("0x54F42802c6B381682c48B530f234De0a896Af1DD")
	msg := ethereum.CallMsg{
		From:      myWallet,
		To:        &SimSwapAddress,
		Gas:       1000000,
		GasPrice:  FloatToTokenAmount(100, 9),
		GasFeeCap: FloatToTokenAmount(100, 9),
		GasTipCap: FloatToTokenAmount(100, 9),
		Value:     FloatToTokenAmount(0.1, 18),
		Data:      data,
	}
	res, err := simClient.CallContract(context.Background(), msg, nil)
	if err != nil {
		panic(err)
	}
	fmt.Println(res)
}

func NewClient(rpcURL string, simAddress common.Address, commonContract OverrideAccounts) (*ethclient.Client, error) {
	httpClient := http.DefaultClient

	sc, err := newSimClient(rpcURL, httpClient, commonContract)
	if err != nil {
		return nil, err
	}

	//simUtil, err := contract.NewSimCall(simAddress, sc)
	//if err != nil {
	//	return nil, errors.WithMessage(err, "create simUtil")
	//}

	return sc, nil
}

func newSimClient(url string, client *http.Client, commonContract OverrideAccounts) (*ethclient.Client, error) {
	round, err := newRoundTripExt(client, commonContract)
	if err != nil {
		return nil, err
	}

	cc := &http.Client{Transport: round}
	r, err := rpc.DialHTTPWithClient(url, cc)
	if err != nil {
		return nil, errors.WithMessage(err, "simclient: dial rpc")
	}

	ethClient := ethclient.NewClient(r)

	return ethClient, nil
}

func newRoundTripExt(c *http.Client, accounts OverrideAccounts) (http.RoundTripper, error) {
	data, err := json.Marshal(accounts)
	if err != nil {
		return nil, err
	}
	return &roundTripperExt{
		c:          c,
		appendData: data,
	}, nil
}

func (r roundTripperExt) RoundTrip(request *http.Request) (*http.Response, error) {
	// Trick: append Config OrderrideAcount to eth_call
	rt := request.Clone(context.Background())
	body, _ := ioutil.ReadAll(request.Body)
	_ = request.Body.Close()
	if len(body) > 0 {
		rt.Body = ioutil.NopCloser(bytes.NewBuffer(body))
		request.Body = ioutil.NopCloser(bytes.NewBuffer(body))
	}
	var req reqMessage
	if err := json.Unmarshal(body, &req); err == nil {
		/*
			req.Method == "eth_call" &&
			(bytes.Contains(req.Params[0], []byte(`0x99f9fbd2`)) ||
			bytes.Contains(req.Params[0], []byte(`0x110bb26c`)))
		*/
		if req.Method == "eth_call" {
			req.Params = append(req.Params, r.appendData)
		}
		d2, err := json.Marshal(req)
		if err != nil {
			panic(err)
		}
		rt.ContentLength = int64(len(d2))
		rt.Body = ioutil.NopCloser(bytes.NewBuffer(d2))
	}
	return r.c.Do(rt)
}

func FloatToTokenAmount(amount float64, decimals int64) *big.Int {
	weiFloat := big.NewFloat(amount)
	decimalsBigFloat := big.NewFloat(0).SetInt(Exp10(decimals))
	amountBig := new(big.Float).Mul(weiFloat, decimalsBigFloat)
	r, _ := amountBig.Int(nil)

	return r
}

// Exp10 ...
func Exp10(n int64) *big.Int {
	return new(big.Int).Exp(big.NewInt(expBase), big.NewInt(n), nil)
}
