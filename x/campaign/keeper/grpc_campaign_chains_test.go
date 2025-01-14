package keeper_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	testkeeper "github.com/tendermint/spn/testutil/keeper"
	"github.com/tendermint/spn/x/campaign/types"
)

func TestCampaignChainsQuerySingle(t *testing.T) {
	ctx, tk, _ := testkeeper.NewTestSetup(t)
	wctx := sdk.WrapSDKContext(ctx)
	msgs := createNCampaignChains(tk.CampaignKeeper, ctx, 2)
	for _, tc := range []struct {
		desc     string
		request  *types.QueryGetCampaignChainsRequest
		response *types.QueryGetCampaignChainsResponse
		err      error
	}{
		{
			desc: "First",
			request: &types.QueryGetCampaignChainsRequest{
				CampaignID: msgs[0].CampaignID,
			},
			response: &types.QueryGetCampaignChainsResponse{CampaignChains: msgs[0]},
		},
		{
			desc: "Second",
			request: &types.QueryGetCampaignChainsRequest{
				CampaignID: msgs[1].CampaignID,
			},
			response: &types.QueryGetCampaignChainsResponse{CampaignChains: msgs[1]},
		},
		{
			desc: "KeyNotFound",
			request: &types.QueryGetCampaignChainsRequest{
				CampaignID: 100000,
			},
			err: status.Error(codes.NotFound, "not found"),
		},
		{
			desc: "InvalidRequest",
			err:  status.Error(codes.InvalidArgument, "invalid request"),
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			response, err := tk.CampaignKeeper.CampaignChains(wctx, tc.request)
			if tc.err != nil {
				require.ErrorIs(t, err, tc.err)
			} else {
				require.Equal(t, tc.response, response)
			}
		})
	}
}
