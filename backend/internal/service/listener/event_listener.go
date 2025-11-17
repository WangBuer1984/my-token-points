package listener

import (
	"context"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/sirupsen/logrus"

	"my-token-points/config"
	"my-token-points/internal/model"
	"my-token-points/internal/repository"
	"my-token-points/internal/service/balance"
)

// EventListener 事件监听器
type EventListener struct {
	chainName       string
	chainConfig     *config.ChainConfig
	client          *ethclient.Client
	contractABI     abi.ABI
	syncRepo        repository.SyncRepository
	balanceService  *balance.BalanceService
	confirmBlocks   int64
	logger          *logrus.Logger
	stopChan        chan struct{}
}

// NewEventListener 创建事件监听器
func NewEventListener(
	chainName string,
	chainConfig *config.ChainConfig,
	confirmBlocks int,
	syncRepo repository.SyncRepository,
	balanceService *balance.BalanceService,
	logger *logrus.Logger,
) (*EventListener, error) {
	// 连接到区块链节点
	client, err := ethclient.Dial(chainConfig.RPCURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s: %w", chainName, err)
	}

	// 解析合约 ABI
	contractABI, err := abi.JSON(strings.NewReader(MyTokenABI))
	if err != nil {
		return nil, fmt.Errorf("failed to parse contract ABI: %w", err)
	}

	return &EventListener{
		chainName:      chainName,
		chainConfig:    chainConfig,
		client:         client,
		contractABI:    contractABI,
		syncRepo:       syncRepo,
		balanceService: balanceService,
		confirmBlocks:  int64(confirmBlocks),
		logger:         logger,
		stopChan:       make(chan struct{}),
	}, nil
}

// Start 启动事件监听
func (l *EventListener) Start(ctx context.Context) error {
	l.logger.Infof("Starting event listener for %s", l.chainName)

	// 初始化同步状态
	if err := l.syncRepo.InitSyncState(ctx, l.chainName, int64(l.chainConfig.StartBlock)); err != nil {
		return fmt.Errorf("failed to init sync state: %w", err)
	}

	// 启动主循环
	go l.run(ctx)

	return nil
}

// Stop 停止事件监听
func (l *EventListener) Stop() {
	l.logger.Infof("Stopping event listener for %s", l.chainName)
	close(l.stopChan)
}

// run 主循环
func (l *EventListener) run(ctx context.Context) {
	ticker := time.NewTicker(time.Duration(l.chainConfig.ScanInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			l.logger.Infof("Context cancelled, stopping listener for %s", l.chainName)
			return
		case <-l.stopChan:
			l.logger.Infof("Stop signal received, stopping listener for %s", l.chainName)
			return
		case <-ticker.C:
			if err := l.scanBlocks(ctx); err != nil {
				l.logger.Errorf("Error scanning blocks for %s: %v", l.chainName, err)
			}
		}
	}
}

// scanBlocks 扫描区块
func (l *EventListener) scanBlocks(ctx context.Context) error {
	// 获取当前链上最新区块
	latestBlock, err := l.client.BlockNumber(ctx)
	if err != nil {
		return fmt.Errorf("failed to get latest block: %w", err)
	}

	// 获取上次同步的区块
	syncState, err := l.syncRepo.GetSyncState(ctx, l.chainName)
	if err != nil {
		return fmt.Errorf("failed to get sync state: %w", err)
	}
	if syncState == nil {
		return fmt.Errorf("sync state not initialized for %s", l.chainName)
	}

	fromBlock := syncState.LastSyncedBlock + 1
	// 延迟确认：只扫描到 latestBlock - confirmBlocks
	toBlock := int64(latestBlock) - l.confirmBlocks

	if fromBlock > toBlock {
		// 没有新区块需要扫描
		return nil
	}

	// 限制每次扫描的区块数量
	batchSize := int64(l.chainConfig.BatchSize)
	if toBlock-fromBlock+1 > batchSize {
		toBlock = fromBlock + batchSize - 1
	}

	l.logger.Debugf("Scanning %s blocks from %d to %d (latest: %d, confirm delay: %d)",
		l.chainName, fromBlock, toBlock, latestBlock, l.confirmBlocks)

	// 查询事件日志
	logs, err := l.queryLogs(ctx, fromBlock, toBlock)
	if err != nil {
		return fmt.Errorf("failed to query logs: %w", err)
	}

	l.logger.Infof("Found %d events in blocks %d-%d on %s", len(logs), fromBlock, toBlock, l.chainName)

	// 处理事件
	for _, vLog := range logs {
		if err := l.processLog(ctx, vLog); err != nil {
			l.logger.Errorf("Failed to process log %s: %v", vLog.TxHash.Hex(), err)
			// 继续处理其他事件，不中断
		}
	}

	// 更新同步状态
	syncState.LastSyncedBlock = toBlock
	syncState.LastConfirmedBlock = toBlock
	syncState.LastSyncAt = time.Now()
	syncState.Status = model.StatusRunning
	if err := l.syncRepo.UpdateSyncState(ctx, syncState); err != nil {
		return fmt.Errorf("failed to update sync state: %w", err)
	}

	return nil
}

// queryLogs 查询事件日志
func (l *EventListener) queryLogs(ctx context.Context, fromBlock, toBlock int64) ([]types.Log, error) {
	contractAddress := common.HexToAddress(l.chainConfig.ContractAddress)

	// 构建过滤器
	query := ethereum.FilterQuery{
		FromBlock: big.NewInt(fromBlock),
		ToBlock:   big.NewInt(toBlock),
		Addresses: []common.Address{contractAddress},
	}

	return l.client.FilterLogs(ctx, query)
}

// processLog 处理单个事件日志
func (l *EventListener) processLog(ctx context.Context, vLog types.Log) error {
	// 解析事件
	event, err := l.contractABI.EventByID(vLog.Topics[0])
	if err != nil {
		// 不是我们关心的事件，忽略
		return nil
	}

	l.logger.Debugf("Processing event %s in tx %s", event.Name, vLog.TxHash.Hex())

	switch event.Name {
	case "TokenMinted":
		return l.handleTokenMinted(ctx, vLog)
	case "TokenBurned":
		return l.handleTokenBurned(ctx, vLog)
	case "Transfer":
		return l.handleTransfer(ctx, vLog)
	default:
		l.logger.Debugf("Ignoring event %s", event.Name)
		return nil
	}
}

// handleTokenMinted 处理 TokenMinted 事件
func (l *EventListener) handleTokenMinted(ctx context.Context, vLog types.Log) error {
	// 解析事件数据
	var event struct {
		To        common.Address
		Amount    *big.Int
		Timestamp *big.Int
	}

	err := l.contractABI.UnpackIntoInterface(&event, "TokenMinted", vLog.Data)
	if err != nil {
		return fmt.Errorf("failed to unpack TokenMinted event: %w", err)
	}

	// indexed 参数在 Topics 中
	if len(vLog.Topics) < 2 {
		return fmt.Errorf("invalid TokenMinted event: not enough topics")
	}
	event.To = common.HexToAddress(vLog.Topics[1].Hex())

	l.logger.Infof("TokenMinted: to=%s, amount=%s, block=%d",
		event.To.Hex(), event.Amount.String(), vLog.BlockNumber)

	// 获取区块时间
	blockTime, err := l.getBlockTime(ctx, vLog.BlockNumber)
	if err != nil {
		return fmt.Errorf("failed to get block time: %w", err)
	}

	// 更新余额
	return l.balanceService.UpdateBalance(ctx, &balance.BalanceUpdate{
		ChainName:   l.chainName,
		UserAddress: event.To.Hex(),
		TxHash:      vLog.TxHash.Hex(),
		BlockNumber: int64(vLog.BlockNumber),
		BlockTime:   blockTime,
		EventType:   model.EventTypeMint,
		AmountDelta: event.Amount.String(),
	})
}

// handleTokenBurned 处理 TokenBurned 事件
func (l *EventListener) handleTokenBurned(ctx context.Context, vLog types.Log) error {
	// 解析事件数据
	var event struct {
		From      common.Address
		Amount    *big.Int
		Timestamp *big.Int
	}

	err := l.contractABI.UnpackIntoInterface(&event, "TokenBurned", vLog.Data)
	if err != nil {
		return fmt.Errorf("failed to unpack TokenBurned event: %w", err)
	}

	// indexed 参数在 Topics 中
	if len(vLog.Topics) < 2 {
		return fmt.Errorf("invalid TokenBurned event: not enough topics")
	}
	event.From = common.HexToAddress(vLog.Topics[1].Hex())

	l.logger.Infof("TokenBurned: from=%s, amount=%s, block=%d",
		event.From.Hex(), event.Amount.String(), vLog.BlockNumber)

	// 获取区块时间
	blockTime, err := l.getBlockTime(ctx, vLog.BlockNumber)
	if err != nil {
		return fmt.Errorf("failed to get block time: %w", err)
	}

	// 更新余额（负数表示减少）
	amountDelta := new(big.Int).Neg(event.Amount)

	return l.balanceService.UpdateBalance(ctx, &balance.BalanceUpdate{
		ChainName:   l.chainName,
		UserAddress: event.From.Hex(),
		TxHash:      vLog.TxHash.Hex(),
		BlockNumber: int64(vLog.BlockNumber),
		BlockTime:   blockTime,
		EventType:   model.EventTypeBurn,
		AmountDelta: amountDelta.String(),
	})
}

// handleTransfer 处理 Transfer 事件
func (l *EventListener) handleTransfer(ctx context.Context, vLog types.Log) error {
	// 解析事件数据
	var event struct {
		From  common.Address
		To    common.Address
		Value *big.Int
	}

	// indexed 参数在 Topics 中
	if len(vLog.Topics) < 3 {
		return fmt.Errorf("invalid Transfer event: not enough topics")
	}
	event.From = common.HexToAddress(vLog.Topics[1].Hex())
	event.To = common.HexToAddress(vLog.Topics[2].Hex())

	err := l.contractABI.UnpackIntoInterface(&event, "Transfer", vLog.Data)
	if err != nil {
		return fmt.Errorf("failed to unpack Transfer event: %w", err)
	}

	l.logger.Infof("Transfer: from=%s, to=%s, amount=%s, block=%d",
		event.From.Hex(), event.To.Hex(), event.Value.String(), vLog.BlockNumber)

	// 获取区块时间
	blockTime, err := l.getBlockTime(ctx, vLog.BlockNumber)
	if err != nil {
		return fmt.Errorf("failed to get block time: %w", err)
	}

	zeroAddress := common.HexToAddress("0x0000000000000000000000000000000000000000")

	// 如果 from 是 0 地址，这是 mint 事件（已由 TokenMinted 处理，忽略）
	if event.From == zeroAddress {
		return nil
	}

	// 如果 to 是 0 地址，这是 burn 事件（已由 TokenBurned 处理，忽略）
	if event.To == zeroAddress {
		return nil
	}

	// 普通转账：减少 from 的余额
	amountDelta := new(big.Int).Neg(event.Value)
	if err := l.balanceService.UpdateBalance(ctx, &balance.BalanceUpdate{
		ChainName:   l.chainName,
		UserAddress: event.From.Hex(),
		TxHash:      vLog.TxHash.Hex(),
		BlockNumber: int64(vLog.BlockNumber),
		BlockTime:   blockTime,
		EventType:   model.EventTypeTransferOut,
		AmountDelta: amountDelta.String(),
	}); err != nil {
		return fmt.Errorf("failed to update from balance: %w", err)
	}

	// 普通转账：增加 to 的余额
	if err := l.balanceService.UpdateBalance(ctx, &balance.BalanceUpdate{
		ChainName:   l.chainName,
		UserAddress: event.To.Hex(),
		TxHash:      vLog.TxHash.Hex(),
		BlockNumber: int64(vLog.BlockNumber),
		BlockTime:   blockTime,
		EventType:   model.EventTypeTransferIn,
		AmountDelta: event.Value.String(),
	}); err != nil {
		return fmt.Errorf("failed to update to balance: %w", err)
	}

	return nil
}

// getBlockTime 获取区块时间
func (l *EventListener) getBlockTime(ctx context.Context, blockNumber uint64) (time.Time, error) {
	block, err := l.client.BlockByNumber(ctx, big.NewInt(int64(blockNumber)))
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(int64(block.Time()), 0), nil
}

