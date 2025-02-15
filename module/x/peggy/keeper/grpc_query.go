package keeper

import (
	"context"

	"github.com/althea-net/peggy/module/x/peggy/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var _ types.QueryServer = Keeper{}

// Params queries the params of the peggy module
func (k Keeper) Params(c context.Context, req *types.QueryParamsRequest) (*types.QueryParamsResponse, error) {
	var params types.Params
	k.paramSpace.GetParamSet(sdk.UnwrapSDKContext(c), &params)
	return &types.QueryParamsResponse{Params: params}, nil

}

// CurrentValset queries the CurrentValset of the peggy module
func (k Keeper) CurrentValset(c context.Context, req *types.QueryCurrentValsetRequest) (*types.QueryCurrentValsetResponse, error) {
	return &types.QueryCurrentValsetResponse{Valset: k.GetCurrentValset(sdk.UnwrapSDKContext(c))}, nil
}

// ValsetRequest queries the ValsetRequest of the peggy module
func (k Keeper) ValsetRequest(c context.Context, req *types.QueryValsetRequestRequest) (*types.QueryValsetRequestResponse, error) {
	return &types.QueryValsetRequestResponse{Valset: k.GetValset(sdk.UnwrapSDKContext(c), req.Nonce)}, nil
}

// ValsetConfirm queries the ValsetConfirm of the peggy module
func (k Keeper) ValsetConfirm(c context.Context, req *types.QueryValsetConfirmRequest) (*types.QueryValsetConfirmResponse, error) {
	addr, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "address invalid")
	}
	return &types.QueryValsetConfirmResponse{Confirm: k.GetValsetConfirm(sdk.UnwrapSDKContext(c), req.Nonce, addr)}, nil
}

// ValsetConfirmsByNonce queries the ValsetConfirmsByNonce of the peggy module
func (k Keeper) ValsetConfirmsByNonce(c context.Context, req *types.QueryValsetConfirmsByNonceRequest) (*types.QueryValsetConfirmsByNonceResponse, error) {
	var confirms []*types.MsgValsetConfirm
	k.IterateValsetConfirmByNonce(sdk.UnwrapSDKContext(c), req.Nonce, func(_ []byte, c types.MsgValsetConfirm) bool {
		confirms = append(confirms, &c)
		return false
	})
	return &types.QueryValsetConfirmsByNonceResponse{Confirms: confirms}, nil
}

// LastValsetRequests queries the LastValsetRequests of the peggy module
func (k Keeper) LastValsetRequests(c context.Context, req *types.QueryLastValsetRequestsRequest) (*types.QueryLastValsetRequestsResponse, error) {
	valReq := k.GetValsets(sdk.UnwrapSDKContext(c))
	valReqLen := len(valReq)
	retLen := 0
	if valReqLen < maxValsetRequestsReturned {
		retLen = valReqLen
	} else {
		retLen = maxValsetRequestsReturned
	}
	return &types.QueryLastValsetRequestsResponse{Valsets: valReq[0:retLen]}, nil
}

// LastPendingValsetRequestByAddr queries the LastPendingValsetRequestByAddr of the peggy module
func (k Keeper) LastPendingValsetRequestByAddr(c context.Context, req *types.QueryLastPendingValsetRequestByAddrRequest) (*types.QueryLastPendingValsetRequestByAddrResponse, error) {
	addr, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "address invalid")
	}

	var pendingValsetReq []*types.Valset
	k.IterateValsets(sdk.UnwrapSDKContext(c), func(_ []byte, val *types.Valset) bool {
		// foundConfirm is true if the operatorAddr has signed the valset we are currently looking at
		foundConfirm := k.GetValsetConfirm(sdk.UnwrapSDKContext(c), val.Nonce, addr) != nil
		// if this valset has NOT been signed by operatorAddr, store it in pendingValsetReq
		// and exit the loop
		if !foundConfirm {
			pendingValsetReq = append(pendingValsetReq, val)
		}
		// if we have more than 100 unconfirmed requests in
		// our array we should exit, TODO pagination
		if len(pendingValsetReq) > 100 {
			return true
		}
		// return false to continue the loop
		return false
	})
	return &types.QueryLastPendingValsetRequestByAddrResponse{Valsets: pendingValsetReq}, nil
}

// BatchFees queries the batch fees from unbatched pool
func (k Keeper) BatchFees(c context.Context, req *types.QueryBatchFeeRequest) (*types.QueryBatchFeeResponse, error) {
	return &types.QueryBatchFeeResponse{BatchFees: k.CreateBatchFees(sdk.UnwrapSDKContext(c))}, nil
}

// LastPendingBatchRequestByAddr queries the LastPendingBatchRequestByAddr of the peggy module
func (k Keeper) LastPendingBatchRequestByAddr(c context.Context, req *types.QueryLastPendingBatchRequestByAddrRequest) (*types.QueryLastPendingBatchRequestByAddrResponse, error) {
	addr, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "address invalid")
	}

	var pendingBatchReq *types.OutgoingTxBatch
	k.IterateOutgoingTXBatches(sdk.UnwrapSDKContext(c), func(_ []byte, batch *types.OutgoingTxBatch) bool {
		foundConfirm := k.GetBatchConfirm(sdk.UnwrapSDKContext(c), batch.BatchNonce, batch.TokenContract, addr) != nil
		if !foundConfirm {
			pendingBatchReq = batch
			return true
		}
		return false
	})

	return &types.QueryLastPendingBatchRequestByAddrResponse{Batch: pendingBatchReq}, nil
}

func (k Keeper) LastPendingLogicCallByAddr(c context.Context, req *types.QueryLastPendingLogicCallByAddrRequest) (*types.QueryLastPendingLogicCallByAddrResponse, error) {
	addr, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidRequest, "address invalid")
	}

	var pendingLogicReq *types.OutgoingLogicCall
	k.IterateOutgoingLogicCalls(sdk.UnwrapSDKContext(c), func(_ []byte, logic *types.OutgoingLogicCall) bool {
		foundConfirm := k.GetLogicCallConfirm(sdk.UnwrapSDKContext(c), logic.InvalidationId, logic.InvalidationNonce, addr) != nil
		if !foundConfirm {
			pendingLogicReq = logic
			return true
		}
		return false
	})
	return &types.QueryLastPendingLogicCallByAddrResponse{Call: pendingLogicReq}, nil
}

// OutgoingTxBatches queries the OutgoingTxBatches of the peggy module
func (k Keeper) OutgoingTxBatches(c context.Context, req *types.QueryOutgoingTxBatchesRequest) (*types.QueryOutgoingTxBatchesResponse, error) {
	var batches []*types.OutgoingTxBatch
	k.IterateOutgoingTXBatches(sdk.UnwrapSDKContext(c), func(_ []byte, batch *types.OutgoingTxBatch) bool {
		batches = append(batches, batch)
		return len(batches) == MaxResults
	})
	return &types.QueryOutgoingTxBatchesResponse{Batches: batches}, nil
}

// OutgoingLogicCalls queries the OutgoingLogicCalls of the peggy module
func (k Keeper) OutgoingLogicCalls(c context.Context, req *types.QueryOutgoingLogicCallsRequest) (*types.QueryOutgoingLogicCallsResponse, error) {
	var calls []*types.OutgoingLogicCall
	k.IterateOutgoingLogicCalls(sdk.UnwrapSDKContext(c), func(_ []byte, call *types.OutgoingLogicCall) bool {
		calls = append(calls, call)
		return len(calls) == MaxResults
	})
	return &types.QueryOutgoingLogicCallsResponse{Calls: calls}, nil
}

// BatchRequestByNonce queries the BatchRequestByNonce of the peggy module
func (k Keeper) BatchRequestByNonce(c context.Context, req *types.QueryBatchRequestByNonceRequest) (*types.QueryBatchRequestByNonceResponse, error) {
	if err := types.ValidateEthAddress(req.ContractAddress); err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, err.Error())
	}
	foundBatch := k.GetOutgoingTXBatch(sdk.UnwrapSDKContext(c), req.ContractAddress, req.Nonce)
	if foundBatch == nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, "Can not find tx batch")
	}
	return &types.QueryBatchRequestByNonceResponse{Batch: foundBatch}, nil
}

// BatchConfirms returns the batch confirmations by nonce and token contract
func (k Keeper) BatchConfirms(c context.Context, req *types.QueryBatchConfirmsRequest) (*types.QueryBatchConfirmsResponse, error) {
	var confirms []*types.MsgConfirmBatch
	k.IterateBatchConfirmByNonceAndTokenContract(sdk.UnwrapSDKContext(c), req.Nonce, req.ContractAddress, func(_ []byte, c types.MsgConfirmBatch) bool {
		confirms = append(confirms, &c)
		return false
	})
	return &types.QueryBatchConfirmsResponse{Confirms: confirms}, nil
}

// LogicConfirms returns the Logic confirmations by nonce and token contract
func (k Keeper) LogicConfirms(c context.Context, req *types.QueryLogicConfirmsRequest) (*types.QueryLogicConfirmsResponse, error) {
	var confirms []*types.MsgConfirmLogicCall
	k.IterateLogicConfirmByInvalidationIdAndNonce(sdk.UnwrapSDKContext(c), req.InvalidationId, req.InvalidationNonce, func(_ []byte, c *types.MsgConfirmLogicCall) bool {
		confirms = append(confirms, c)
		return false
	})
	return &types.QueryLogicConfirmsResponse{Confirms: confirms}, nil
}

// LastEventNonceByAddr returns the last event nonce for the given validator address, this allows eth oracles to figure out where they left off
func (k Keeper) LastEventNonceByAddr(c context.Context, req *types.QueryLastEventNonceByAddrRequest) (*types.QueryLastEventNonceByAddrResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	var ret types.QueryLastEventNonceByAddrResponse
	addr, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrInvalidAddress, req.Address)
	}
	validator := k.GetOrchestratorValidator(ctx, addr)
	if validator == nil {
		return nil, sdkerrors.Wrap(types.ErrUnknown, "address")
	}
	lastEventNonce := k.GetLastEventNonceByValidator(ctx, validator)
	ret.EventNonce = lastEventNonce
	return &ret, nil
}

// DenomToERC20 queries the Cosmos Denom that maps to an Ethereum ERC20
func (k Keeper) DenomToERC20(c context.Context, req *types.QueryDenomToERC20Request) (*types.QueryDenomToERC20Response, error) {
	ctx := sdk.UnwrapSDKContext(c)
	cosmos_originated, erc20, err := k.DenomToERC20Lookup(ctx, req.Denom)
	var ret types.QueryDenomToERC20Response
	ret.Erc20 = erc20
	ret.CosmosOriginated = cosmos_originated
	return &ret, err
}

// ERC20ToDenom queries the ERC20 contract that maps to an Ethereum ERC20 if any
func (k Keeper) ERC20ToDenom(c context.Context, req *types.QueryERC20ToDenomRequest) (*types.QueryERC20ToDenomResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	cosmos_originated, name := k.ERC20ToDenomLookup(ctx, req.Erc20)
	var ret types.QueryERC20ToDenomResponse
	ret.Denom = name
	ret.CosmosOriginated = cosmos_originated
	return &ret, nil
}

func (k Keeper) GetDelegateKeyByValidator(c context.Context, req *types.QueryDelegateKeysByValidatorAddress) (*types.QueryDelegateKeysByValidatorAddressResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	keys := k.GetDelegateKeys(ctx)
	_, err := sdk.ValAddressFromBech32(req.ValidatorAddress)
	if err != nil {
		return nil, err
	}
	for _, key := range keys {
		if req.ValidatorAddress == key.Validator {
			return &types.QueryDelegateKeysByValidatorAddressResponse{EthAddress: key.EthAddress, OrchestratorAddress: key.Orchestrator}, nil
		}

	}
	return nil, sdkerrors.Wrap(types.ErrInvalid, "No validator")
}

func (k Keeper) GetDelegateKeyByOrchestrator(c context.Context, req *types.QueryDelegateKeysByOrchestratorAddress) (*types.QueryDelegateKeysByOrchestratorAddressResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	keys := k.GetDelegateKeys(ctx)
	_, err := sdk.AccAddressFromBech32(req.OrchestratorAddress)
	if err != nil {
		return nil, err
	}
	for _, key := range keys {
		if req.OrchestratorAddress == key.Orchestrator {
			return &types.QueryDelegateKeysByOrchestratorAddressResponse{ValidatorAddress: key.Validator, EthAddress: key.EthAddress}, nil
		}

	}
	return nil, sdkerrors.Wrap(types.ErrInvalid, "No validator")
}

func (k Keeper) GetDelegateKeyByEth(c context.Context, req *types.QueryDelegateKeysByEthAddress) (*types.QueryDelegateKeysByEthAddressResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	keys := k.GetDelegateKeys(ctx)
	if err := types.ValidateEthAddress(req.EthAddress); err != nil {
		return nil, sdkerrors.Wrap(err, "invalid eth address")
	}
	for _, key := range keys {
		if req.EthAddress == key.EthAddress {
			return &types.QueryDelegateKeysByEthAddressResponse{ValidatorAddress: key.Validator, OrchestratorAddress: key.Orchestrator}, nil
		}

	}
	return nil, sdkerrors.Wrap(types.ErrInvalid, "No validator")
}

func (k Keeper) GetPendingSendToEth(c context.Context, req *types.QueryPendingSendToEth) (*types.QueryPendingSendToEthResponse, error) {
	ctx := sdk.UnwrapSDKContext(c)
	batches := k.GetOutgoingTxBatches(ctx)
	unbatched_tx := k.GetPoolTransactions(ctx)
	sender_address := req.SenderAddress
	var res *types.QueryPendingSendToEthResponse
	for _, batch := range batches {
		for _, tx := range batch.Transactions {
			if tx.Sender == sender_address {
				res.TransfersInBatches = append(res.TransfersInBatches, tx)
			}
		}
	}
	for _, tx := range unbatched_tx {
		if tx.Sender == sender_address {
			res.UnbatchedTransfers = append(res.UnbatchedTransfers, tx)
		}
	}
	return res, nil
}
