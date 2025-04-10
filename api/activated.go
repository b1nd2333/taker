package api

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum"
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

	// 获取当前网络 Gas 建议价
	gasTipCap, _ := bscClient.SuggestGasTipCap(context.Background())
	gasFeeCap, _ := bscClient.SuggestGasPrice(context.Background())
	//gasLimit, _ := bscClient.EstimateGas(context.Background())

	if gasFeeCap == nil || gasTipCap == nil {
		gasFeeCap = maxFeePerGas
		gasTipCap = maxPriorityFeePerGas
	}
	// 确保 MaxFee ≥ MaxPriorityFee
	if gasFeeCap.Cmp(gasTipCap) < 0 {
		gasFeeCap = new(big.Int).Mul(gasTipCap, big.NewInt(2))
	}

	gasLimit, err := estimateGasLimit(bscClient, address, toAddress, byteData)
	if err != nil {
		return errors.New(fmt.Sprintf("账号%dGas估算失败: %v，跳过", taker.ID, err))
	}

	// 构建 EIP-1559 类型的交易
	tx := types.NewTx(&types.DynamicFeeTx{
		ChainID:   big.NewInt(1125),
		Nonce:     nonce,
		GasFeeCap: gasFeeCap,
		GasTipCap: gasFeeCap,
		Gas:       gasLimit, //130224,63000, // 转账交易通常需要 21000 Gas
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
		if strings.Contains(err.Error(), "insufficient funds for gas * price + value") {
			color.Yellow("账号%d余额不足，跳过")
			return nil // 余额不足直接退出
		}
		color.Red("交易手续费失败7:%s", err)
		return err
	}

	// 新增：获取交易收据验证状态
	txHash := signedTx.Hash()
	var receipt *types.Receipt
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	color.Yellow("账号%d交易hash：%s，等待上链", taker.ID, signedTx.Hash().Hex())
	// 轮询获取收据（最多尝试36次，每次间隔5秒）
	for i := 0; i < 36; i++ {
		receipt, err = bscClient.TransactionReceipt(ctx, txHash)
		if err == nil {
			break
		}
		time.Sleep(5 * time.Second)
	}

	if err != nil || receipt == nil {
		color.Red("账号%d无法获取本次交易收据:%s", taker.ID, err)
		return fmt.Errorf("transaction receipt failed: %v", err)
	}

	if receipt.Status != 1 {
		color.Red("账号%d交易执行失败，区块高度：#%v", taker.ID, receipt.BlockNumber)
		return fmt.Errorf("transaction reverted")
	}

	fmt.Printf("账号%d，地址%sMint成功，交易哈希：%s\n", taker.ID, taker.Address, signedTx.Hash().Hex())
	return nil

}

func estimateGasLimit(client *ethclient.Client, from common.Address, to common.Address, data []byte) (uint64, error) {
	msg := ethereum.CallMsg{
		From:  from,
		To:    &to,
		Value: big.NewInt(0),
		Data:  data,
	}

	// 基础估算
	gas, err := client.EstimateGas(context.Background(), msg)
	if err != nil {
		return 0, fmt.Errorf("%v", err)
	}

	// 添加20%缓冲
	buffered := gas * 110 / 100
	return buffered, nil
}
