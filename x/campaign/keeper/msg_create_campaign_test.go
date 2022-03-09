package keeper_test

import (
	"testing"

	testkeeper "github.com/tendermint/spn/testutil/keeper"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	"github.com/stretchr/testify/require"

	"github.com/tendermint/spn/testutil/sample"
	"github.com/tendermint/spn/x/campaign/keeper"
	"github.com/tendermint/spn/x/campaign/types"
	profiletypes "github.com/tendermint/spn/x/profile/types"
)

func initCreationFeeAndFundCoordAccounts(
	t *testing.T,
	keeper *keeper.Keeper,
	bk bankkeeper.Keeper,
	sdkCtx sdk.Context,
	fee sdk.Coins,
	numCreations int64,
	addrs ...string,
) {
	// set fee param to `coins`
	params := keeper.GetParams(sdkCtx)
	params.CampaignCreationFee = fee
	keeper.SetParams(sdkCtx, params)

	coins := sdk.NewCoins()
	for _, coin := range fee {
		coin.Amount = coin.Amount.MulRaw(numCreations)
		coins = coins.Add(coin)
	}

	// add `coins` to balance of each coordinator address
	for _, addr := range addrs {
		accAddr, err := sdk.AccAddressFromBech32(addr)
		require.NoError(t, err)
		err = bk.MintCoins(sdkCtx, types.ModuleName, coins)
		require.NoError(t, err)
		err = bk.SendCoinsFromModuleToAccount(sdkCtx, types.ModuleName, accAddr, coins)
		require.NoError(t, err)
	}
}

func TestMsgCreateCampaign(t *testing.T) {
	var (
		coordAddrs          = make([]string, 3)
		coordMap            = make(map[string]uint64)
		sdkCtx, tk, ts      = testkeeper.NewTestSetup(t)
		ctx                 = sdk.WrapSDKContext(sdkCtx)
		campaignCreationFee = sample.Coins()
	)

	// Create coordinators
	for i := range coordAddrs {
		addr := sample.Address()
		coordAddrs[i] = addr
		msgCreateCoordinator := sample.MsgCreateCoordinator(addr)
		resCoord, err := ts.ProfileSrv.CreateCoordinator(ctx, &msgCreateCoordinator)
		require.NoError(t, err)
		coordMap[addr] = resCoord.CoordinatorID
	}

	// assign random sdk.Coins to `campaignCreationFee` param and provide balance to coordinators
	// coordAddrs[2] is not funded
	initCreationFeeAndFundCoordAccounts(t, tk.CampaignKeeper, tk.BankKeeper, sdkCtx, campaignCreationFee, 1, coordAddrs[:2]...)

	for _, tc := range []struct {
		name       string
		msg        types.MsgCreateCampaign
		expectedID uint64
		err        error
	}{
		{
			name: "create a campaign 1",
			msg: types.MsgCreateCampaign{
				CampaignName: sample.CampaignName(),
				Coordinator:  coordAddrs[0],
				TotalSupply:  sample.TotalSupply(),
				Metadata:     sample.Metadata(20),
			},
			expectedID: uint64(0),
		},
		{
			name: "create a campaign from a different coordinator",
			msg: types.MsgCreateCampaign{
				CampaignName: sample.CampaignName(),
				Coordinator:  coordAddrs[1],
				TotalSupply:  sample.TotalSupply(),
				Metadata:     sample.Metadata(20),
			},
			expectedID: uint64(1),
		},
		{
			name: "create a campaign from a non existing coordinator",
			msg: types.MsgCreateCampaign{
				CampaignName: sample.CampaignName(),
				Coordinator:  sample.Address(),
				TotalSupply:  sample.TotalSupply(),
				Metadata:     sample.Metadata(20),
			},
			err: profiletypes.ErrCoordAddressNotFound,
		},
		{
			name: "create a campaign with an invalid token supply",
			msg: types.MsgCreateCampaign{
				CampaignName: sample.CampaignName(),
				Coordinator:  coordAddrs[0],
				TotalSupply:  sample.CoinsWithRange(10, 20),
				Metadata:     sample.Metadata(20),
			},
			err: types.ErrInvalidTotalSupply,
		},
		{
			name: "insufficient balance to cover creation fee",
			msg: types.MsgCreateCampaign{
				CampaignName: sample.CampaignName(),
				Coordinator:  coordAddrs[2],
				TotalSupply:  sample.TotalSupply(),
				Metadata:     sample.Metadata(20),
			},
			err: sdkerrors.ErrInsufficientFunds,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			// get account initial balance
			accAddr, err := sdk.AccAddressFromBech32(tc.msg.Coordinator)
			require.NoError(t, err)
			preBalance := tk.BankKeeper.SpendableCoins(sdkCtx, accAddr)

			got, err := ts.CampaignSrv.CreateCampaign(ctx, &tc.msg)
			if tc.err != nil {
				require.ErrorIs(t, err, tc.err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.expectedID, got.CampaignID)
			campaign, found := tk.CampaignKeeper.GetCampaign(sdkCtx, got.CampaignID)
			require.True(t, found)
			require.EqualValues(t, got.CampaignID, campaign.CampaignID)
			require.EqualValues(t, tc.msg.CampaignName, campaign.CampaignName)
			require.EqualValues(t, coordMap[tc.msg.Coordinator], campaign.CoordinatorID)
			require.False(t, campaign.MainnetInitialized)
			require.True(t, tc.msg.TotalSupply.IsEqual(campaign.TotalSupply))
			require.EqualValues(t, types.EmptyShares(), campaign.AllocatedShares)
			require.EqualValues(t, types.EmptyShares(), campaign.TotalShares)

			// dynamic share is always disabled
			require.EqualValues(t, false, campaign.DynamicShares)

			// Empty list of campaign chains
			campaignChains, found := tk.CampaignKeeper.GetCampaignChains(sdkCtx, got.CampaignID)
			require.True(t, found)
			require.EqualValues(t, got.CampaignID, campaignChains.CampaignID)
			require.Empty(t, campaignChains.Chains)

			// check fee deduction
			postBalance := tk.BankKeeper.SpendableCoins(sdkCtx, accAddr)
			require.True(t, preBalance.Sub(campaignCreationFee).IsEqual(postBalance))
		})
	}
}