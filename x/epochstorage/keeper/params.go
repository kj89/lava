package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/lavanet/lava/utils"
	"github.com/lavanet/lava/x/epochstorage/types"
)

// GetParams get all parameters as types.Params
func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	return types.NewParams(
		k.UnstakeHoldBlocksRaw(ctx),
		k.EpochBlocksRaw(ctx),
		k.EpochsToSaveRaw(ctx),
		k.LatestParamChange(ctx),
	)
}

// SetParams set the params
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramstore.SetParamSet(ctx, &params)
}

func (k Keeper) UnstakeHoldBlocks(ctx sdk.Context, block uint64) (res uint64) {
	// Unstake Hold Blocks is always used for the latest, but we want to use the fixated
	k.GetParamForBlock(ctx, string(types.KeyUnstakeHoldBlocks), block, &res)
	return
}

// UnstakeHoldBlocksRaw returns the UnstakeHoldBlocks param
func (k Keeper) UnstakeHoldBlocksRaw(ctx sdk.Context) (res uint64) {
	// Unstake Hold Blocks is always used for the latest, but we want to use the fixated
	k.paramstore.Get(ctx, types.KeyUnstakeHoldBlocks, &res)
	return
}

// EpochBlocks returns the EpochBlocks fixated param
func (k Keeper) EpochBlocks(ctx sdk.Context, block uint64) (res uint64, err error) {
	err = k.GetParamForBlock(ctx, string(types.KeyEpochBlocks), block, &res)
	return
}

// EpochBlocks returns the EpochBlocks param
func (k Keeper) EpochBlocksRaw(ctx sdk.Context) (res uint64) {
	k.paramstore.Get(ctx, types.KeyEpochBlocks, &res)
	return
}

// EpochsToSave returns the EpochsToSave fixated param
func (k Keeper) EpochsToSave(ctx sdk.Context, block uint64) (res uint64, err error) {
	err = k.GetParamForBlock(ctx, string(types.KeyEpochsToSave), block, &res)
	return
}

// EpochsToSaveRaw returns the EpochsToSave param
func (k Keeper) EpochsToSaveRaw(ctx sdk.Context) (res uint64) {
	// TODO: change to support param change
	k.paramstore.Get(ctx, types.KeyEpochsToSave, &res)
	return
}

// return if this block is an epoch start
func (k Keeper) IsEpochStart(ctx sdk.Context) (res bool) {
	currentBlock := uint64(ctx.BlockHeight())
	blockInEpoch, err := k.BlockInEpoch(ctx, currentBlock)
	if err != nil {
		utils.LavaError(ctx, k.Logger(ctx), "IsEpochStart", map[string]string{"error": err.Error()}, "can't get block in epoch")
		return false
	}
	return blockInEpoch == 0
}

func (k Keeper) BlocksToSave(ctx sdk.Context, block uint64) (res uint64, erro error) {
	epochsToSave, err := k.EpochsToSave(ctx, block)
	epochBlocks, err2 := k.EpochBlocks(ctx, block)
	blocksToSave := epochsToSave * epochBlocks
	if err != nil || err2 != nil {
		erro = fmt.Errorf("BlocksToSave param read errors %s, %s", err.Error(), err2.Error())
	}
	return blocksToSave, erro
}

func (k Keeper) BlockInEpoch(ctx sdk.Context, block uint64) (res uint64, err error) {
	// get epochBlocks directly because we also need an epoch start on the current grid and when fixation was saved is an epoch start
	fixtedParams, err := k.GetFixatedParamsForBlock(ctx, string(types.KeyEpochBlocks), block)
	var blocksCycle uint64
	utils.Deserialize(fixtedParams.Parameter, &blocksCycle)
	epochStartInGrid := fixtedParams.FixationBlock // fixation block is always <= block
	blockRelativeToGrid := block - epochStartInGrid
	return blockRelativeToGrid % blocksCycle, err
}

func (k Keeper) LatestParamChange(ctx sdk.Context) (res uint64) {
	k.paramstore.Get(ctx, types.KeyLatestParamChange, &res)
	return
}

// GetEpochStartForBlock gets a session start supports one param change
func (k Keeper) GetEpochStartForBlock(ctx sdk.Context, block uint64) (epochStart uint64, blockInEpoch uint64, err error) {
	blockInTargetEpoch, err := k.BlockInEpoch(ctx, block)
	targetEpochStart := block - blockInTargetEpoch
	return targetEpochStart, blockInTargetEpoch, err
}

func (k Keeper) GetNextEpoch(ctx sdk.Context, block uint64) (nextEpoch uint64, erro error) {
	epochBlocks, err := k.EpochBlocks(ctx, block)
	epochStart, _, err2 := k.GetEpochStartForBlock(ctx, block)
	nextEpoch = epochStart + epochBlocks
	if err != nil {
		erro = err
	} else if err2 != nil {
		erro = err2
	}
	return nextEpoch, erro
}

func (k Keeper) GetPreviousEpochStartForBlock(ctx sdk.Context, block uint64) (previousEpochStart uint64, erro error) {
	epochStart, _, err := k.GetEpochStartForBlock(ctx, block)
	if epochStart <= 0 {
		return 0, utils.LavaFormatError("GetPreviousEpochStartForBlock", fmt.Errorf("GetPreviousEpochStartForBlock tried to fetch epoch beyond zero"), nil)
	}
	previousEpochStart, _, err2 := k.GetEpochStartForBlock(ctx, epochStart-1) // we take one block before the target epoch so it belongs to the previous epoch
	if err != nil {
		erro = err
	} else if err2 != nil {
		erro = err2
	}
	return
}
