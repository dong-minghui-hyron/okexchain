package keeper

import (
	"fmt"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/okex/okexchain/x/gov/types"
)

// NewQuerier returns all query handlers
func NewQuerier(keeper Keeper) sdk.Querier {
	return func(ctx sdk.Context, path []string, req abci.RequestQuery) ([]byte, sdk.Error) {
		switch path[0] {
		case types.QueryParams:
			return queryParams(ctx, path[1:], req, keeper)
		case types.QueryProposals:
			return queryProposals(ctx, path[1:], req, keeper)
		case types.QueryProposal:
			return queryProposal(ctx, path[1:], req, keeper)
		case types.QueryDeposits:
			return queryDeposits(ctx, path[1:], req, keeper)
		case types.QueryDeposit:
			return queryDeposit(ctx, path[1:], req, keeper)
		case types.QueryVotes:
			return queryVotes(ctx, path[1:], req, keeper)
		case types.QueryVote:
			return queryVote(ctx, path[1:], req, keeper)
		case types.QueryTally:
			return queryTally(ctx, path[1:], req, keeper)
		default:
			return nil, sdk.ErrUnknownRequest("unknown gov query endpoint")
		}
	}
}

func queryParams(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {
	switch path[0] {
	case types.ParamDeposit:
		bz, err := codec.MarshalJSONIndent(keeper.cdc, keeper.GetDepositParams(ctx))
		if err != nil {
			return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, fmt.Sprintf("could not marshal result to JSON %s", err.Error()))
		}
		return bz, nil
	case types.ParamVoting:
		bz, err := codec.MarshalJSONIndent(keeper.cdc, keeper.GetVotingParams(ctx))
		if err != nil {
			return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, fmt.Sprintf("could not marshal result to JSON %s", err.Error()))
		}
		return bz, nil
	case types.ParamTallying:
		bz, err := codec.MarshalJSONIndent(keeper.cdc, keeper.GetTallyParams(ctx))
		if err != nil {
			return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, fmt.Sprintf("could not marshal result to JSON %s", err.Error()))
		}
		return bz, nil
	default:
		return nil, sdk.ErrUnknownRequest(fmt.Sprintf("%s is not a valid query request path", req.Path))
	}
}

// nolint: unparam
func queryProposal(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {
	var params types.QueryProposalParams
	err := keeper.cdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, fmt.Sprintf("incorrectly formatted request data %s", err.Error()))
	}

	proposal, ok := keeper.GetProposal(ctx, params.ProposalID)
	if !ok {
		return nil, types.ErrUnknownProposal(types.DefaultCodespace, params.ProposalID)
	}

	bz, err := codec.MarshalJSONIndent(keeper.cdc, proposal)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, fmt.Sprintf("could not marshal result to JSON %s", err.Error()))
	}
	return bz, nil
}

// nolint: unparam
func queryDeposit(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {
	var params types.QueryDepositParams
	err := keeper.cdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, fmt.Sprintf("incorrectly formatted request data %s", err.Error()))
	}

	deposit, _ := keeper.GetDeposit(ctx, params.ProposalID, params.Depositor)
	bz, err := codec.MarshalJSONIndent(keeper.cdc, deposit)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, fmt.Sprintf("could not marshal result to JSON %s", err.Error()))
	}
	return bz, nil
}

// nolint: unparam
func queryVote(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {
	var params types.QueryVoteParams
	err := keeper.cdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, fmt.Sprintf("incorrectly formatted request data %s", err.Error()))
	}

	vote, _ := keeper.GetVote(ctx, params.ProposalID, params.Voter)
	bz, err := codec.MarshalJSONIndent(keeper.cdc, vote)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, fmt.Sprintf("could not marshal result to JSON %s", err.Error()))
	}
	return bz, nil
}

// nolint: unparam
func queryDeposits(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {
	var params types.QueryProposalParams
	err := keeper.cdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, fmt.Sprintf("incorrectly formatted request data %s", err.Error()))
	}

	deposits := keeper.GetDeposits(ctx, params.ProposalID)

	bz, err := codec.MarshalJSONIndent(keeper.cdc, deposits)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, fmt.Sprintf("could not marshal result to JSON %s", err.Error()))
	}
	return bz, nil
}

// nolint: unparam
func queryTally(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {
	var params types.QueryProposalParams
	err := keeper.cdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, fmt.Sprintf("incorrectly formatted request data %s", err.Error()))
	}

	proposalID := params.ProposalID

	proposal, ok := keeper.GetProposal(ctx, proposalID)
	if !ok {
		return nil, types.ErrUnknownProposal(types.DefaultCodespace, proposalID)
	}

	var tallyResult types.TallyResult

	switch proposal.Status {
	case types.StatusDepositPeriod:
		tallyResult = types.EmptyTallyResult(keeper.totalPower(ctx))
	case types.StatusPassed, types.StatusRejected, types.StatusFailed:
		tallyResult = proposal.FinalTallyResult
	default:
		// proposal is in voting period
		_, _, tallyResult = Tally(ctx, keeper, proposal, true)
	}

	bz, err := codec.MarshalJSONIndent(keeper.cdc, tallyResult)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, fmt.Sprintf("could not marshal result to JSON %s", err.Error()))
	}
	return bz, nil
}

// nolint: unparam
func queryVotes(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {
	var params types.QueryProposalParams
	err := keeper.cdc.UnmarshalJSON(req.Data, &params)

	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, fmt.Sprintf("incorrectly formatted request data %s", err.Error()))
	}

	votes := keeper.GetVotes(ctx, params.ProposalID)

	bz, err := codec.MarshalJSONIndent(keeper.cdc, votes)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, fmt.Sprintf("could not marshal result to JSON %s", err.Error()))
	}
	return bz, nil
}

// nolint: unparam
func queryProposals(ctx sdk.Context, path []string, req abci.RequestQuery, keeper Keeper) ([]byte, sdk.Error) {
	var params types.QueryProposalsParams
	err := keeper.cdc.UnmarshalJSON(req.Data, &params)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrUnknownRequest, fmt.Sprintf("incorrectly formatted request data %s", err.Error()))
	}

	proposals := keeper.GetProposalsFiltered(ctx, params.Voter, params.Depositor, params.ProposalStatus, params.Limit)

	bz, err := codec.MarshalJSONIndent(keeper.cdc, proposals)
	if err != nil {
		return nil, sdkerrors.Wrap(sdkerrors.ErrJSONMarshal, fmt.Sprintf("could not marshal result to JSON %s", err.Error()))
	}
	return bz, nil
}
