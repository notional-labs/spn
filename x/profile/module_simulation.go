package profile

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/baseapp"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	"github.com/tendermint/spn/testutil/sample"
	"github.com/tendermint/spn/x/profile/keeper"
	"github.com/tendermint/spn/x/profile/types"
)

const (
	weightMsgUpdateValidatorDescription   = 50
	weightMsgDeleteValidator              = 10
	weightMsgCreateCoordinator            = 50
	weightMsgUpdateCoordinatorDescription = 20
	weightMsgUpdateCoordinatorAddress     = 20
	weightMsgDeleteCoordinator            = 10
)

// GenerateGenesisState creates a randomized GenState of the module
func (AppModule) GenerateGenesisState(simState *module.SimulationState) {
	profileGenesis := sample.ProfileGenesisState()
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(&profileGenesis)
}

// ProposalContents doesn't return any content functions for governance proposals
func (AppModule) ProposalContents(_ module.SimulationState) []simtypes.WeightedProposalContent {
	return nil
}

// RandomizedParams creates randomized  param changes for the simulator
func (am AppModule) RandomizedParams(r *rand.Rand) []simtypes.ParamChange {
	return []simtypes.ParamChange{}
}

// RegisterStoreDecoder registers a decoder
func (am AppModule) RegisterStoreDecoder(_ sdk.StoreDecoderRegistry) {}

// WeightedOperations returns the all the gov module operations with their respective weights.
func (am AppModule) WeightedOperations(simState module.SimulationState) []simtypes.WeightedOperation {
	return []simtypes.WeightedOperation{
		simulation.NewWeightedOperation(
			weightMsgUpdateValidatorDescription,
			SimulateMsgUpdateValidatorDescription(am.accountKeeper, am.bankKeeper, am.keeper),
		),
		simulation.NewWeightedOperation(
			weightMsgDeleteValidator,
			SimulateMsgDeleteValidator(am.accountKeeper, am.bankKeeper, am.keeper),
		),
		simulation.NewWeightedOperation(
			weightMsgCreateCoordinator,
			SimulateMsgCreateCoordinator(am.accountKeeper, am.bankKeeper, am.keeper),
		),
		simulation.NewWeightedOperation(
			weightMsgUpdateCoordinatorDescription,
			SimulateMsgUpdateCoordinatorDescription(am.accountKeeper, am.bankKeeper, am.keeper),
		),
		simulation.NewWeightedOperation(
			weightMsgUpdateCoordinatorAddress,
			SimulateMsgUpdateCoordinatorAddress(am.accountKeeper, am.bankKeeper, am.keeper),
		),
		simulation.NewWeightedOperation(
			weightMsgDeleteCoordinator,
			SimulateMsgDeleteCoordinator(am.accountKeeper, am.bankKeeper, am.keeper),
		),
	}
}

// SimulateMsgUpdateValidatorDescription simulates a MsgUpdateValidatorDescription message
func SimulateMsgUpdateValidatorDescription(ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		// Select a random account
		accountNb := r.Intn(len(accs))

		desc := sample.ValidatorDescription(sample.String(50))
		msg := types.NewMsgUpdateValidatorDescription(
			accs[accountNb].Address.String(),
			desc.Identity,
			desc.Moniker,
			desc.Website,
			desc.SecurityContact,
			desc.Details,
		)

		txCtx := simulation.OperationInput{
			R:               r,
			App:             app,
			TxGen:           simappparams.MakeTestEncodingConfig().TxConfig,
			Cdc:             nil,
			Msg:             msg,
			MsgType:         msg.Type(),
			Context:         ctx,
			SimAccount:      accs[accountNb],
			AccountKeeper:   ak,
			Bankkeeper:      bk,
			ModuleName:      types.ModuleName,
			CoinsSpentInMsg: sdk.NewCoins(),
		}
		return simulation.GenAndDeliverTxWithRandFees(txCtx)
	}
}

// SimulateMsgDeleteValidator simulates a MsgUpdateValidatorDescription message
func SimulateMsgDeleteValidator(ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		var simAccount simtypes.Account

		// Find an account with validator description
		var found bool
		for _, acc := range accs {
			_, found = k.GetValidator(ctx, acc.Address.String())
			if found {
				simAccount = acc
				break
			}
		}
		if !found {
			// No message if no validator description
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgDeleteValidator, "skip validator delete"), nil, nil
		}

		msg := types.NewMsgDeleteValidator(simAccount.Address.String())
		txCtx := simulation.OperationInput{
			R:               r,
			App:             app,
			TxGen:           simappparams.MakeTestEncodingConfig().TxConfig,
			Cdc:             nil,
			Msg:             msg,
			MsgType:         msg.Type(),
			Context:         ctx,
			SimAccount:      simAccount,
			AccountKeeper:   ak,
			Bankkeeper:      bk,
			ModuleName:      types.ModuleName,
			CoinsSpentInMsg: sdk.NewCoins(),
		}
		return simulation.GenAndDeliverTxWithRandFees(txCtx)
	}
}

// SimulateMsgCreateCoordinator simulates a MsgCreateCoordinator message
func SimulateMsgCreateCoordinator(ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		var simAccount simtypes.Account

		// Find an account with no coordinator
		allCreated := true
		for _, acc := range accs {
			_, found := k.GetCoordinatorByAddress(ctx, acc.Address.String())
			if !found {
				simAccount = acc
				allCreated = false
				break
			}
		}
		if allCreated {
			// No message if all coordinator created
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgCreateCoordinator, "skip coordinator creation"), nil, nil
		}

		msg := types.NewMsgCreateCoordinator(
			simAccount.Address.String(),
			sample.String(30),
			sample.String(30),
			sample.String(30),
		)

		txCtx := simulation.OperationInput{
			R:               r,
			App:             app,
			TxGen:           simappparams.MakeTestEncodingConfig().TxConfig,
			Cdc:             nil,
			Msg:             msg,
			MsgType:         msg.Type(),
			Context:         ctx,
			SimAccount:      simAccount,
			AccountKeeper:   ak,
			Bankkeeper:      bk,
			ModuleName:      types.ModuleName,
			CoinsSpentInMsg: sdk.NewCoins(),
		}
		return simulation.GenAndDeliverTxWithRandFees(txCtx)
	}
}

// SimulateMsgUpdateCoordinatorDescription simulates a MsgUpdateCoordinatorDescription message
func SimulateMsgUpdateCoordinatorDescription(ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		var simAccount simtypes.Account

		// Find an account with coordinator associated
		var found bool
		for _, acc := range accs {
			_, found = k.GetCoordinatorByAddress(ctx, acc.Address.String())
			if found {
				simAccount = acc
				break
			}
		}
		if !found {
			// No message if no coordinator
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgUpdateCoordinatorDescription, "skip update coordinator description"), nil, nil
		}

		desc := sample.CoordinatorDescription()
		msg := types.NewMsgUpdateCoordinatorDescription(
			simAccount.Address.String(),
			desc.Identity,
			desc.Website,
			desc.Details,
		)

		txCtx := simulation.OperationInput{
			R:               r,
			App:             app,
			TxGen:           simappparams.MakeTestEncodingConfig().TxConfig,
			Cdc:             nil,
			Msg:             msg,
			MsgType:         msg.Type(),
			Context:         ctx,
			SimAccount:      simAccount,
			AccountKeeper:   ak,
			Bankkeeper:      bk,
			ModuleName:      types.ModuleName,
			CoinsSpentInMsg: sdk.NewCoins(),
		}
		return simulation.GenAndDeliverTxWithRandFees(txCtx)
	}
}

// SimulateMsgUpdateCoordinatorAddress simulates a MsgUpdateCoordinatorAddress message
func SimulateMsgUpdateCoordinatorAddress(ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		var simAccount simtypes.Account

		// Find an account with coordinator associated
		var found bool
		for _, acc := range accs {
			_, found = k.GetCoordinatorByAddress(ctx, acc.Address.String())
			if found {
				simAccount = acc
				break
			}
		}
		if !found {
			// No message if no coordinator
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgUpdateCoordinatorAddress, "skip update coordinator address"), nil, nil
		}

		msg := types.NewMsgUpdateCoordinatorAddress(simAccount.Address.String(), sample.Address())
		txCtx := simulation.OperationInput{
			R:               r,
			App:             app,
			TxGen:           simappparams.MakeTestEncodingConfig().TxConfig,
			Cdc:             nil,
			Msg:             msg,
			MsgType:         msg.Type(),
			Context:         ctx,
			SimAccount:      simAccount,
			AccountKeeper:   ak,
			Bankkeeper:      bk,
			ModuleName:      types.ModuleName,
			CoinsSpentInMsg: sdk.NewCoins(),
		}
		return simulation.GenAndDeliverTxWithRandFees(txCtx)
	}
}

// SimulateMsgDeleteCoordinator simulates a MsgDeleteCoordinator message
func SimulateMsgDeleteCoordinator(ak types.AccountKeeper, bk types.BankKeeper, k keeper.Keeper) simtypes.Operation {
	return func(
		r *rand.Rand, app *baseapp.BaseApp, ctx sdk.Context, accs []simtypes.Account, chainID string,
	) (simtypes.OperationMsg, []simtypes.FutureOperation, error) {
		var simAccount simtypes.Account

		// Find an account with coordinator associated
		var found bool
		for _, acc := range accs {
			_, found = k.GetCoordinatorByAddress(ctx, acc.Address.String())
			if found {
				simAccount = acc
				break
			}
		}
		if !found {
			// No message if no coordinator
			return simtypes.NoOpMsg(types.ModuleName, types.TypeMsgDeleteCoordinator, "skip update coordinator delete"), nil, nil
		}

		msg := types.NewMsgDeleteCoordinator(simAccount.Address.String())
		txCtx := simulation.OperationInput{
			R:               r,
			App:             app,
			TxGen:           simappparams.MakeTestEncodingConfig().TxConfig,
			Cdc:             nil,
			Msg:             msg,
			MsgType:         msg.Type(),
			Context:         ctx,
			SimAccount:      simAccount,
			AccountKeeper:   ak,
			Bankkeeper:      bk,
			ModuleName:      types.ModuleName,
			CoinsSpentInMsg: sdk.NewCoins(),
		}
		return simulation.GenAndDeliverTxWithRandFees(txCtx)
	}
}