package api

import (
	"context"
	"encoding/hex"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/fatih/color"
	"math/big"
	"strings"
	"time"
)

func activated(taker Taker) error {
	// 创建交易
	bscClient, err := ethclient.Dial("https://rpc-mainnet.taker.xyz")
	if err != nil {
		color.Red("交易手续费失败1:%s", err)
		return err
	}

	// 你的BSC账户地址
	address := common.HexToAddress(taker.Address)

	// 获取账户的nonce
	data := ""
	nonce, err := bscClient.PendingNonceAt(context.Background(), address)
	if err != nil {
		color.Red("交易手续费失败2:%s", err)
		return err
	}
	// 构造交易数据
	data = "0x02fb0c5e"

	// 打印交易数据
	//fmt.Println(data)
	toAddress := common.HexToAddress("0xb3efe5105b835e5dd9d206445dbd66df24b912ab")

	// 构建交易
	//fmt.Println(data)
	byteData, err := hex.DecodeString(data[2:])
	if err != nil {
		color.Red("交易手续费失败3:%s", err)
		return err
	}
	maxFeePerGas := new(big.Int).SetUint64(1500000001)         // 1.51 Gwei
	maxPriorityFeePerGas := new(big.Int).SetUint64(1500000000) // 1.5 Gwei

	// 构建 EIP-1559 类型的交易
	tx := types.NewTx(&types.DynamicFeeTx{
		ChainID:   big.NewInt(1125),
		Nonce:     nonce,
		GasFeeCap: maxFeePerGas,
		GasTipCap: maxPriorityFeePerGas,
		Gas:       130224, //130224,63000, // 转账交易通常需要 21000 Gas
		To:        &toAddress,
		Value:     big.NewInt(0),
		Data:      byteData,
	})
	//fmt.Println(tx)
	//tx := types.NewTransaction(nonce, toAddress, new(big.Int).SetUint64(0), uint64(153608), new(big.Int).SetUint64(1500000001), byteData)

	// 签名交易
	privateKeyHex := taker.PrivateKey
	if strings.Contains(taker.PrivateKey, "0x") {
		privateKeyHex = taker.PrivateKey[2:]
	}
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		color.Red("交易手续费失败4:%s", err)
		return err
	}

	// 签名交易
	chainID, err := bscClient.ChainID(context.Background())
	if err != nil {
		color.Red("交易手续费失败5:%s", err)
		return err
	}

	signedTx, err := types.SignTx(tx, types.NewLondonSigner(chainID), privateKey)
	if err != nil {
		color.Red("交易手续费失败6:%s", err)
		return err
	}

	// 发送交易
	err = bscClient.SendTransaction(context.Background(), signedTx)
	if err != nil {
		color.Red("交易手续费失败7:%s", err)
		return err
	}

	color.Green("第%d行，地址%s交易成功，交易哈希：%s\n", taker.ID, taker.Address, signedTx.Hash().Hex())
	// 网站上确定hash
	time.Sleep(5 * time.Second)
	return nil
}
