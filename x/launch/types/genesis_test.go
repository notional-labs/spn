package types_test

import (
	"github.com/stretchr/testify/require"
	"github.com/tendermint/spn/testutil/sample"
	"github.com/tendermint/spn/x/launch/types"
	"testing"
)

func TestGenesisState_Validate(t *testing.T) {
	chainId1 := sample.AlphaString(5)
	chainId2 := sample.AlphaString(5)
	addr1 := sample.AccAddress()
	addr2 := sample.AccAddress()
	vestedAddress := sample.AccAddress()

	for _, tc := range []struct {
		desc          string
		genState      *types.GenesisState
		shouldBeValid bool
	}{
		{
			desc:          "default is valid",
			genState:      types.DefaultGenesis(),
			shouldBeValid: true,
		},
		{
			desc: "valid genesis state",
			genState: &types.GenesisState{
				ChainList: []*types.Chain{
					{
						ChainID: chainId1,
					},
					{
						ChainID: chainId2,
					},
				},
				GenesisAccountList: []*types.GenesisAccount{
					{
						ChainID: chainId1,
						Address: addr1,
					},
					{
						ChainID: chainId1,
						Address: addr2,
					},
					{
						ChainID: chainId2,
						Address: addr1,
					},
					{
						ChainID: chainId2,
						Address: addr2,
					},
				},
				VestedAccountList: []*types.VestedAccount{
					{
						ChainID: chainId1,
						Address: vestedAddress,
					},
					{
						ChainID: chainId2,
						Address: vestedAddress,
					},
				},
			},
			shouldBeValid: true,
		},
		{
			desc: "duplicated chains",
			genState: &types.GenesisState{
				ChainList: []*types.Chain{
					{
						ChainID: chainId1,
					},
					{
						ChainID: chainId1,
					},
				},
			},
			shouldBeValid: false,
		},
		{
			desc: "duplicated accounts",
			genState: &types.GenesisState{
				GenesisAccountList: []*types.GenesisAccount{
					{
						ChainID: chainId1,
						Address: addr1,
					},
					{
						ChainID: chainId1,
						Address: addr1,
					},
				},
			},
			shouldBeValid: false,
		},
		{
			desc: "account not associated with chain",
			genState: &types.GenesisState{
				ChainList: []*types.Chain{
					{
						ChainID: chainId1,
					},
				},
				GenesisAccountList: []*types.GenesisAccount{
					{
						ChainID: chainId2,
						Address: addr1,
					},
				},
			},
			shouldBeValid: false,
		},
		{
			desc: "duplicated vested accounts",
			genState: &types.GenesisState{
				VestedAccountList: []*types.VestedAccount{
					{
						ChainID: chainId1,
						Address: vestedAddress,
					},
					{
						ChainID: chainId1,
						Address: vestedAddress,
					},
				},
			},
			shouldBeValid: false,
		},
		{
			desc: "vested account not associated with chain",
			genState: &types.GenesisState{
				ChainList: []*types.Chain{
					{
						ChainID: chainId1,
					},
				},
				VestedAccountList: []*types.VestedAccount{
					{
						ChainID: chainId2,
						Address: vestedAddress,
					},
				},
			},
			shouldBeValid: false,
		},
		{
			desc: "address as genesis account and vested account",
			genState: &types.GenesisState{
				ChainList: []*types.Chain{
					{
						ChainID: chainId1,
					},
				},
				GenesisAccountList: []*types.GenesisAccount{
					{
						ChainID: chainId1,
						Address: addr1,
					},
				},
				VestedAccountList: []*types.VestedAccount{
					{
						ChainID: chainId1,
						Address: addr1,
					},
				},
			},
			shouldBeValid: false,
		},
	} {
		tc := tc
		t.Run(tc.desc, func(t *testing.T) {
			err := tc.genState.Validate()
			if tc.shouldBeValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}