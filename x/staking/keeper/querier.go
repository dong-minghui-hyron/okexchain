package keeper

import (
	"fmt"
	"strings"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/tendermint/tendermint/crypto"

	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/okex/okexchain/x/staking/types"
)

// NewQuerier creates a querier for staking REST endpoints
func NewQuerier(k Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) (res []byte, err error) {
		switch path[0] {
		case types.QueryValidators:
			return queryValidators(ctx, req, k)
		case types.QueryValidator:
			return queryValidator(ctx, req, k)
		case types.QueryPool:
			return queryPool(ctx, k)
		case types.QueryParameters:
			return queryParameters(ctx, k)
			// required by okexchain
		case types.QueryUnbondingDelegation:
			return queryUndelegation(ctx, req, k)
		case types.QueryValidatorAllShares:
			return queryValidatorAllShares(ctx, req, k)
		case types.QueryAddress:
			return queryAddress(ctx, k)
		case types.QueryForAddress:
			return queryForAddress(ctx, req, k)
		case types.QueryForAccAddress:
			return queryForAccAddress(ctx, req)
		case types.QueryProxy:
			return queryProxy(ctx, req, k)
		case types.QueryDelegator:
			return queryDelegator(ctx, req, k)
		default:
			return nil, sdkerrors.ErrUnknownRequest
		}
	}
}

func queryDelegator(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, error) {
	var params types.QueryDelegatorParams

	if err := types.ModuleCdc.UnmarshalJSON(req.Data, &params); err != nil {
		return nil, defaultQueryErrParseParams(err)
	}

	delegator, found := k.GetDelegator(ctx, params.DelegatorAddr)
	if !found {
		return nil, types.ErrNoDelegatorExisted(types.DefaultCodespace, params.DelegatorAddr.String())
	}

	res, err := codec.MarshalJSONIndent(types.ModuleCdc, delegator)
	if err != nil {
		return nil, defaultQueryErrJSONMarshal(err)
	}

	return res, nil
}

func queryValidators(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, error) {
	var params types.QueryValidatorsParams

	if err := types.ModuleCdc.UnmarshalJSON(req.Data, &params); err != nil {
		return nil, defaultQueryErrParseParams(err)
	}

	validators := k.GetAllValidators(ctx)

	var filteredVals []types.Validator
	if params.Status == "all" {
		filteredVals = validators
	} else {
		filteredVals = make([]types.Validator, 0, len(validators))
		for _, val := range validators {
			if strings.EqualFold(val.GetStatus().String(), params.Status) {
				filteredVals = append(filteredVals, val)
			}
		}

		start, end := client.Paginate(len(filteredVals), params.Page, params.Limit, int(k.GetParams(ctx).MaxValidators))
		if start < 0 || end < 0 {
			filteredVals = []types.Validator{}
		} else {
			filteredVals = filteredVals[start:end]
		}
	}

	res, err := codec.MarshalJSONIndent(types.ModuleCdc, filteredVals)
	if err != nil {
		return nil, defaultQueryErrJSONMarshal(err)
	}

	return res, nil
}

func queryValidator(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, error) {
	var params types.QueryValidatorParams

	err := types.ModuleCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, defaultQueryErrParseParams(err)
	}

	validator, found := k.GetValidator(ctx, params.ValidatorAddr)
	if !found {
		return nil, types.ErrNoValidatorFound(types.DefaultCodespace, params.ValidatorAddr.String())
	}

	res, err := codec.MarshalJSONIndent(types.ModuleCdc, validator)
	if err != nil {
		return nil, defaultQueryErrJSONMarshal(err)
	}

	return res, nil
}

func queryPool(ctx sdk.Context, k Keeper) ([]byte, error) {
	bondDenom := k.BondDenom(ctx)
	bondedPool := k.GetBondedPool(ctx)
	notBondedPool := k.GetNotBondedPool(ctx)
	if bondedPool == nil || notBondedPool == nil {
		return nil, sdkerrors.New(types.ModuleName, types.CodeInternalError, "pool accounts haven't been set")
	}

	pool := types.NewPool(
		notBondedPool.GetCoins().AmountOf(bondDenom),
		bondedPool.GetCoins().AmountOf(bondDenom),
	)

	res, err := codec.MarshalJSONIndent(types.ModuleCdc, pool)
	if err != nil {
		return nil, defaultQueryErrJSONMarshal(err)
	}

	return res, nil
}

func queryParameters(ctx sdk.Context, k Keeper) ([]byte, error) {
	params := k.GetParams(ctx)

	res, err := codec.MarshalJSONIndent(types.ModuleCdc, params)
	if err != nil {
		return nil, defaultQueryErrJSONMarshal(err)
	}

	return res, nil
}

func queryProxy(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, error) {
	var params types.QueryDelegatorParams
	if err := types.ModuleCdc.UnmarshalJSON(req.Data, &params); err != nil {
		return nil, defaultQueryErrParseParams(err)
	}

	delAddrs := k.GetDelegatorsByProxy(ctx, params.DelegatorAddr)
	resp, err := codec.MarshalJSONIndent(types.ModuleCdc, delAddrs)
	if err != nil {
		return nil, defaultQueryErrJSONMarshal(err)
	}

	return resp, nil
}

func queryValidatorAllShares(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, error) {
	var params types.QueryValidatorParams

	if err := types.ModuleCdc.UnmarshalJSON(req.Data, &params); err != nil {
		return nil, sdkerrors.New(types.ModuleName, types.CodeInternalError, fmt.Sprintf("failed to parse validator params. %s", err))
	}

	sharesResponses := k.GetValidatorAllShares(ctx, params.ValidatorAddr)
	resp, err := codec.MarshalJSONIndent(types.ModuleCdc, sharesResponses)
	if err != nil {
		return nil, defaultQueryErrJSONMarshal(err)
	}

	return resp, nil
}

func queryUndelegation(ctx sdk.Context, req abci.RequestQuery, k Keeper) ([]byte, error) {
	var params types.QueryDelegatorParams
	err := types.ModuleCdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, defaultQueryErrParseParams(err)
	}

	undelegation, found := k.GetUndelegating(ctx, params.DelegatorAddr)
	if !found {
		return nil, types.ErrNoUnbondingDelegation(types.DefaultCodespace)
	}

	res, err := codec.MarshalJSONIndent(types.ModuleCdc, undelegation)
	if err != nil {
		return nil, defaultQueryErrJSONMarshal(err)
	}

	return res, nil
}

func queryAddress(ctx sdk.Context, k Keeper) (res []byte, err error) {

	ovPairs := k.GetOperAndValidatorAddr(ctx)
	res, errRes := codec.MarshalJSONIndent(types.ModuleCdc, ovPairs)
	if errRes != nil {
		return nil, defaultQueryErrJSONMarshal(errRes)
	}
	return res, nil
}

func queryForAddress(ctx sdk.Context, req abci.RequestQuery, k Keeper) (res []byte, err error) {
	validatorAddr := string(req.Data)
	if len(validatorAddr) != crypto.AddressSize*2 {
		return nil, types.ErrBadValidatorAddr(types.DefaultCodespace)
	}

	operAddr, found := k.GetOperAddrFromValidatorAddr(ctx, validatorAddr)
	if !found {
		return nil, types.ErrNoValidatorFound(types.DefaultCodespace, validatorAddr)
	}

	res, errRes := codec.MarshalJSONIndent(types.ModuleCdc, operAddr)
	if errRes != nil {
		return nil, defaultQueryErrJSONMarshal(errRes)
	}
	return res, nil
}

func queryForAccAddress(ctx sdk.Context, req abci.RequestQuery) (res []byte, err error) {

	valAddr, errBech32 := sdk.ValAddressFromBech32(string(req.Data))
	if errBech32 != nil {
		return nil, sdkerrors.New(types.ModuleName, types.CodeInternalError, "failed to get operator address"+errBech32.Error())
	}

	accAddr := sdk.AccAddress(valAddr)

	res, errMarshal := codec.MarshalJSONIndent(types.ModuleCdc, accAddr)
	if errMarshal != nil {
		return nil, defaultQueryErrJSONMarshal(errMarshal)
	}
	return res, nil
}

func defaultQueryErrJSONMarshal(err error) error {
	return sdkerrors.New(types.ModuleName, types.CodeInternalError, "failed to marshal result to JSON"+err.Error())
}

func defaultQueryErrParseParams(err error) error {
	return sdkerrors.New(types.ModuleName, types.CodeInternalError, fmt.Sprintf("failed to parse params. %s", err))
}
