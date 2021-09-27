package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/spn/testutil/sample"
	"github.com/tendermint/spn/x/campaign/types"
	profiletypes "github.com/tendermint/spn/x/profile/types"
)

func TestMsgUpdateTotalShares(t *testing.T) {
	var (
		coordAddr1                                         = sample.AccAddress()
		coordAddr2                                         = sample.AccAddress()
		campaignKeeper, _, campaignSrv, profileSrv, sdkCtx = setupMsgServer(t)
		ctx                                                = sdk.WrapSDKContext(sdkCtx)
	)

	// Create coordinators
	res, err := profileSrv.CreateCoordinator(ctx, &profiletypes.MsgCreateCoordinator{
		Address:     coordAddr1,
		Description: sample.CoordinatorDescription(),
	})
	require.NoError(t, err)
	coordID := res.CoordinatorId
	res, err = profileSrv.CreateCoordinator(ctx, &profiletypes.MsgCreateCoordinator{
		Address:     coordAddr2,
		Description: sample.CoordinatorDescription(),
	})
	require.NoError(t, err)

	// Set different campaigns
	campaign := sample.Campaign(0)
	campaign.CoordinatorID = coordID
	campaign.DynamicShares = true
	campaignKeeper.SetCampaign(sdkCtx, campaign)

	campaignMainnetInitialized := sample.Campaign(1)
	campaignMainnetInitialized.CoordinatorID = coordID
	campaignMainnetInitialized.DynamicShares = true
	campaignMainnetInitialized.MainnetInitialized = true
	campaignKeeper.SetCampaign(sdkCtx, campaignMainnetInitialized)

	campaignNoDynamicShares := sample.Campaign(2)
	campaignNoDynamicShares.CoordinatorID = coordID
	campaignKeeper.SetCampaign(sdkCtx, campaignNoDynamicShares)

	campaignWithAllocatedShares := sample.Campaign(3)
	campaignWithAllocatedShares.CoordinatorID = coordID
	campaignWithAllocatedShares.DynamicShares = true
	campaignWithAllocatedShares.AllocatedShares, _ = types.NewShares("100foo")
	campaignKeeper.SetCampaign(sdkCtx, campaignWithAllocatedShares)
	smallerTotalShares, _ := types.NewShares("50foo")

	for _, tc := range []struct {
		name string
		msg  types.MsgUpdateTotalShares
		err  error
	}{
		{
			name: "update total shares",
			msg: types.MsgUpdateTotalShares{
				CampaignID:  0,
				Coordinator: coordAddr1,
				TotalShares: sample.Shares(),
			},
		},
		{
			name: "can update total shares again",
			msg: types.MsgUpdateTotalShares{
				CampaignID:  0,
				Coordinator: coordAddr1,
				TotalShares: sample.Shares(),
			},
		},
		{
			name: "campaign not found",
			msg: types.MsgUpdateTotalShares{
				CampaignID:  100,
				Coordinator: coordAddr1,
				TotalShares: sample.Shares(),
			},
			err: types.ErrCampaignNotFound,
		},
		{
			name: "non existing coordinator",
			msg: types.MsgUpdateTotalShares{
				CampaignID:  0,
				Coordinator: sample.AccAddress(),
				TotalShares: sample.Shares(),
			},
			err: profiletypes.ErrCoordAddressNotFound,
		},
		{
			name: "not the coordinator of the campaign",
			msg: types.MsgUpdateTotalShares{
				CampaignID:  0,
				Coordinator: coordAddr2,
				TotalShares: sample.Shares(),
			},
			err: profiletypes.ErrCoordInvalid,
		},
		{
			name: "cannot update total shares when mainnet is initialized",
			msg: types.MsgUpdateTotalShares{
				CampaignID:  1,
				Coordinator: coordAddr1,
				TotalShares: sample.Shares(),
			},
			err: types.ErrMainnetInitialized,
		},
		{
			name: "cannot update total shares when dynamic shares option not set",
			msg: types.MsgUpdateTotalShares{
				CampaignID:  2,
				Coordinator: coordAddr1,
				TotalShares: sample.Shares(),
			},
			err: types.ErrNoDynamicShares,
		},
		{
			name: "cannot update total shares when below allocated shares",
			msg: types.MsgUpdateTotalShares{
				CampaignID:  3,
				Coordinator: coordAddr1,
				TotalShares: smallerTotalShares,
			},
			err: types.ErrInvalidShares,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := campaignSrv.UpdateTotalShares(ctx, &tc.msg)
			if tc.err != nil {
				require.ErrorIs(t, err, tc.err)
				return
			}
			require.NoError(t, err)
			campaign, found := campaignKeeper.GetCampaign(sdkCtx, tc.msg.CampaignID)
			require.True(t, found)
			require.True(t, sdk.Coins(tc.msg.TotalShares).IsEqual(sdk.Coins(campaign.TotalShares)))
		})
	}
}