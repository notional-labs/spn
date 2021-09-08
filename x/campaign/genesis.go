package campaign

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/spn/x/campaign/keeper"
	"github.com/tendermint/spn/x/campaign/types"
)

// InitGenesis initializes the capability module's state from a provided genesis
// state.
func InitGenesis(ctx sdk.Context, k keeper.Keeper, genState types.GenesisState) {
	// this line is used by starport scaffolding # genesis/module/init
	// Set all the campaign
	for _, elem := range genState.CampaignList {
		k.SetCampaign(ctx, elem)
	}

	// Set campaign count
	k.SetCampaignCount(ctx, genState.CampaignCount)

}

// ExportGenesis returns the capability module's exported genesis.
func ExportGenesis(ctx sdk.Context, k keeper.Keeper) *types.GenesisState {
	genesis := types.DefaultGenesis()

	// this line is used by starport scaffolding # genesis/module/export
	genesis.CampaignList = k.GetAllCampaign(ctx)
	genesis.CampaignCount = k.GetCampaignCount(ctx)

	return genesis
}