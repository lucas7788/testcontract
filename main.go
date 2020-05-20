package main

import (
	"github.com/ontio/ontology-go-sdk"
	"fmt"
	"github.com/ontio/ontology/common"
	"github.com/ontio/ontology-go-sdk/examples/define"
	"github.com/ontio/ontology/core/utils"
	"time"
	common2 "github.com/ontio/ontology-go-sdk/common"
)

var resource_id = []byte("reso_1")    //test param can be changed

func main() {
	fmt.Println("==========================start============================")
	testUrl := "http://127.0.0.1:20336"
	//mainUrl := "http://dappnode2.ont.io:20336"
	testUrl = "http://polaris1.ont.io:20336"
	ontSdk := ontology_go_sdk.NewOntologySdk()
	ontSdk.NewRpcClient().SetAddress(testUrl)

	pwd := []byte("123456")

	var seller *ontology_go_sdk.Account
	var buyer *ontology_go_sdk.Account
	var agent *ontology_go_sdk.Account
	wallet, err := ontSdk.OpenWallet("./wallet.dat")
	if err != nil {
		fmt.Println("OpenWallet error: ", err)
		return
	}
	if false {
		seller, _ = wallet.NewDefaultSettingAccount(pwd)
		buyer, _ = wallet.NewDefaultSettingAccount(pwd)
		wallet.Save()
		fmt.Printf("seller:%s, buyer:%s\n", seller.Address.ToBase58(), buyer.Address.ToBase58())
		return
	} else {
		seller, _ = wallet.GetAccountByAddress("Aejfo7ZX5PVpenRj23yChnyH64nf8T1zbu", pwd)
		buyer, _ = wallet.GetAccountByAddress("AHhXa11suUgVLX1ZDFErqBd3gskKqLfa5N", pwd)
		agent, _ = wallet.GetAccountByAddress("ANb3bf1b67WP2ZPh5HQt4rkrmphMJmMCMK", pwd)
	}

	fmt.Printf("seller:%s, buyer:%s\n", seller.Address.ToBase58(), buyer.Address.ToBase58())


	contractAddr,_ := common.AddressFromHexString("")

	ddxf := NewDDXF(ontSdk, contractAddr)
	fmt.Printf("contractAddr:%s, contractAddr:%s\n", contractAddr.ToBase58(), contractAddr.ToHexString())

	showOngBalance(ontSdk, seller.Address, buyer.Address)

	//test ddxf contract
	if true {
		tokenHash := make([]byte, 32)    //test param can be changed, only the length is 32 can be success,other will be failed.
		template := define.TokenTemplate{
			DataIDs:   "",
			TokenHash: string(tokenHash),
		}

		//发布商品
		if false {
			param := getPublishParam(seller.Address, tokenHash, template)
			ddxf.invoke(seller, "dtokenSellerPublish", param)
			return
		}

		//购买 dtoken
		if false {
			// 第一个参数resource_id 用来标识购买的哪个商品，
			// 第二个参数 1是购买的数量,  单价乘以数量 = 该次会扣的钱
			// 第三个参数 购买的地址
			param := []interface{}{resource_id, 1, buyer.Address}
			ddxf.invoke(buyer, "buyDtoken", param)
			showOngBalance(ontSdk, seller.Address, buyer.Address)
			return
		}
		//消耗dtoken
		if false {
			// 第一个参数resource_id 用来标识购买的哪个商品，
			// 第二个参数 1是消耗的数量
			// 第三个参数 拥有该dtoken的地址
			param := []interface{}{resource_id, buyer, template.ToBytes(), 1}
			ddxf.invoke(buyer, "useToken", param)
			return
		}
		//添加代理
		if false {
			param := []interface{}{resource_id, buyer.Address, []interface{}{agent.Address}, 1}
			bs, _ := utils.BuildWasmContractParam(param)
			fmt.Println(common.ToHexString(bs))
			ddxf.invoke(buyer, "addAgents", param)
			return
		}
	}
}

func showOngBalance(ontSdk *ontology_go_sdk.OntologySdk, seller common.Address, buyer common.Address) {
	seller_ba, _ := ontSdk.Native.Ong.BalanceOf(seller)
	buyer_ba, _ := ontSdk.Native.Ong.BalanceOf(buyer)
	fmt.Printf("seller_ba:%d,buyer_ba:%d\n", seller_ba, buyer_ba)
}

func getPublishParam(seller common.Address, tokenHash []byte, template define.TokenTemplate) []interface{} {
	tokenResourceType := make(map[define.TokenTemplate]byte)
	tokenResourceType[template] = byte(0)
	tokenEndpoint := make(map[define.TokenTemplate]string)
	tokenEndpoint[template] = "endpoint2"
	ddo := define.ResourceDDO{
		ResourceType:      byte(1),
		TokenResourceType: tokenResourceType,    // RT for tokens
		Manager:           seller,               // data owner id
		Endpoint:          "endpoint",           // data service provider uri
		TokenEndpoint:     tokenEndpoint,        // endpoint for tokens
		DescHash:          "",                   // required if len(Templates) > 1
		DTC:               common.ADDRESS_EMPTY, // can be empty
		MP:                common.ADDRESS_EMPTY, // can be empty
		Split:             common.ADDRESS_EMPTY,
	}

	item := define.DTokenItem{
		Fee: define.Fee{
			ContractAddr: common.ADDRESS_EMPTY,
			ContractType: byte(1), //合约类型，0标识ont, 1标识ong，2标识oep4
			Count:        100,     //单价
		},
		ExpiredDate: uint64(time.Now().Unix()) + uint64(1000),
		Stocks:      100,
		Templates:   []define.TokenTemplate{template},
	}
	// 第一个参数 resource_id 用来标识发布的商品，
	// 第二个参数 ResourceDDO  商品信息
	// 第三个参数 DTokenItem  收费信息
	return []interface{}{resource_id, ddo.ToBytes(), item.ToBytes()}
}

type DDXF struct {
	sdk             *ontology_go_sdk.OntologySdk
	gasLimit        uint64
	gasPrice        uint64
	contractAddress common.Address
	timeoutSec      time.Duration
}

func NewDDXF(sdk *ontology_go_sdk.OntologySdk, contractAddress common.Address) *DDXF {
	return &DDXF{
		sdk:             sdk,
		gasLimit:        200000000,
		gasPrice:        500,
		contractAddress: contractAddress,
		timeoutSec:      30 * time.Second,
	}
}

func (this *DDXF) deploy(signer *ontology_go_sdk.Account, codeHash string) error {

	txHash, err := this.sdk.WasmVM.DeployWasmVMSmartContract(
		this.gasPrice,
		this.gasLimit,
		signer,
		codeHash,
		"ddxf wasm",
		"1.0",
		"author",
		"email",
		"desc",
	)

	if err != nil {
		fmt.Printf("error in DeployWasmVMSmartContract:%s\n", err)
		return err
	}
	_, err = this.sdk.WaitForGenerateBlock(this.timeoutSec)
	if err != nil {
		fmt.Printf("error in WaitForGenerateBlock:%s\n", err)

		return err
	}
	fmt.Printf("the deploy contract txhash is %s\n", txHash.ToHexString())
	return nil
}

func (this *DDXF) preInvoke(method string, param []interface{}) (*common2.ResultItem, error) {
	res, err := this.sdk.WasmVM.PreExecInvokeWasmVMContract(this.contractAddress, method, param)
	if err != nil {
		fmt.Println("InvokeWasmVMSmartContract error ", err)
		return nil, err
	}
	fmt.Printf("state:%d\n", res.State)
	return res.Result, nil
}

func (this *DDXF) invoke(signer *ontology_go_sdk.Account, method string, param []interface{}) error {
	txhash, err := this.sdk.WasmVM.InvokeWasmVMSmartContract(this.gasPrice, this.gasLimit, signer, signer, this.contractAddress, method, param)
	if err != nil {
		fmt.Println("InvokeWasmVMSmartContract error ", err)
		return err
	}

	timeoutSec := 30 * time.Second
	_, err = this.sdk.WaitForGenerateBlock(timeoutSec)
	if err != nil {
		fmt.Println("WaitForGenerateBlock error ", err)
		return err
	}
	fmt.Printf("method:%s, txHash:%s\n", method, txhash.ToHexString())
	event, err := this.sdk.GetSmartContractEvent(txhash.ToHexString())
	if err != nil {
		fmt.Println("GetSmartContractEvent error ", err)
		return err
	}
	if event != nil {
		for _, notify := range event.Notify {
			fmt.Printf("%+v\n", notify)
		}
	}
	return nil
}