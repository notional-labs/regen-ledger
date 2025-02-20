package server_test

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkmodule "github.com/cosmos/cosmos-sdk/types/module"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	disttypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	params "github.com/cosmos/cosmos-sdk/x/params/types/proposal"

	"github.com/regen-network/regen-ledger/types/v2/testutil/fixture"
	"github.com/regen-network/regen-ledger/x/ecocredit/v3"
	"github.com/regen-network/regen-ledger/x/ecocredit/v3/basket"
	"github.com/regen-network/regen-ledger/x/ecocredit/v3/module"
	"github.com/regen-network/regen-ledger/x/ecocredit/v3/server/testsuite"
)

func TestServer(t *testing.T) {
	ff, bankKeeper, accountKeeper := setup(t)
	s := testsuite.NewIntegrationTestSuite(ff, bankKeeper, accountKeeper)
	suite.Run(t, s)
}

func TestGenesis(t *testing.T) {
	ff, bankKeeper, _ := setup(t)
	s := testsuite.NewGenesisTestSuite(ff, bankKeeper)
	suite.Run(t, s)
}

func setup(t *testing.T) (fixture.Factory, bankkeeper.BaseKeeper, authkeeper.AccountKeeper) {
	ff := fixture.NewFixtureFactory(t, 8)
	baseApp := ff.BaseApp()
	cdc := ff.Codec()
	amino := codec.NewLegacyAmino()

	authtypes.RegisterInterfaces(cdc.InterfaceRegistry())
	params.RegisterInterfaces(cdc.InterfaceRegistry())

	authKey := sdk.NewKVStoreKey(authtypes.StoreKey)
	bankKey := sdk.NewKVStoreKey(banktypes.StoreKey)
	distKey := sdk.NewKVStoreKey(disttypes.StoreKey)
	paramsKey := sdk.NewKVStoreKey(paramstypes.StoreKey)
	ecoKey := sdk.NewKVStoreKey(ecocredit.ModuleName)
	tkey := sdk.NewTransientStoreKey(paramstypes.TStoreKey)

	baseApp.MountStore(authKey, storetypes.StoreTypeIAVL)
	baseApp.MountStore(ecoKey, storetypes.StoreTypeIAVL)
	baseApp.MountStore(bankKey, storetypes.StoreTypeIAVL)
	baseApp.MountStore(distKey, storetypes.StoreTypeIAVL)
	baseApp.MountStore(paramsKey, storetypes.StoreTypeIAVL)
	baseApp.MountStore(tkey, storetypes.StoreTypeTransient)

	authSubspace := paramstypes.NewSubspace(cdc, amino, paramsKey, tkey, authtypes.ModuleName)
	bankSubspace := paramstypes.NewSubspace(cdc, amino, paramsKey, tkey, banktypes.ModuleName)
	ecocreditSubspace := paramstypes.NewSubspace(cdc, amino, paramsKey, tkey, ecocredit.ModuleName)

	maccPerms := map[string][]string{
		minttypes.ModuleName:       {authtypes.Minter},
		ecocredit.ModuleName:       {authtypes.Burner},
		basket.BasketSubModuleName: {authtypes.Burner, authtypes.Minter},
	}

	accountKeeper := authkeeper.NewAccountKeeper(
		cdc, authKey, authSubspace, authtypes.ProtoBaseAccount, maccPerms, "regen",
	)

	bankKeeper := bankkeeper.NewBaseKeeper(
		cdc, bankKey, accountKeeper, bankSubspace, nil,
	)

	authority := authtypes.NewModuleAddress(govtypes.ModuleName)
	ecocreditModule := module.NewModule(ecoKey, authority, accountKeeper, bankKeeper, ecocreditSubspace, nil)
	ff.SetModules([]sdkmodule.AppModule{ecocreditModule})

	return ff, bankKeeper, accountKeeper
}
