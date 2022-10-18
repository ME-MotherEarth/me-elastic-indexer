package accounts

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"time"

	"github.com/ME-MotherEarth/me-core/core"
	"github.com/ME-MotherEarth/me-core/core/check"
	coreData "github.com/ME-MotherEarth/me-core/data"
	"github.com/ME-MotherEarth/me-core/data/mect"
	"github.com/ME-MotherEarth/me-core/marshal"
	indexer "github.com/ME-MotherEarth/me-elastic-indexer"
	"github.com/ME-MotherEarth/me-elastic-indexer/converters"
	"github.com/ME-MotherEarth/me-elastic-indexer/data"
	logger "github.com/ME-MotherEarth/me-logger"
	vmcommon "github.com/ME-MotherEarth/me-vm-common"
)

var log = logger.GetOrCreate("indexer/process/accounts")

// accountsProcessor is a structure responsible for processing accounts
type accountsProcessor struct {
	internalMarshalizer    marshal.Marshalizer
	addressPubkeyConverter core.PubkeyConverter
	accountsDB             indexer.AccountsAdapter
	balanceConverter       indexer.BalanceConverter
	shardID                uint32
}

// NewAccountsProcessor will create a new instance of accounts processor
func NewAccountsProcessor(
	marshalizer marshal.Marshalizer,
	addressPubkeyConverter core.PubkeyConverter,
	accountsDB indexer.AccountsAdapter,
	balanceConverter indexer.BalanceConverter,
	shardID uint32,
) (*accountsProcessor, error) {
	if check.IfNil(marshalizer) {
		return nil, indexer.ErrNilMarshalizer
	}
	if check.IfNil(addressPubkeyConverter) {
		return nil, indexer.ErrNilPubkeyConverter
	}
	if check.IfNil(accountsDB) {
		return nil, indexer.ErrNilAccountsDB
	}
	if check.IfNil(balanceConverter) {
		return nil, indexer.ErrNilBalanceConverter
	}

	return &accountsProcessor{
		internalMarshalizer:    marshalizer,
		addressPubkeyConverter: addressPubkeyConverter,
		accountsDB:             accountsDB,
		balanceConverter:       balanceConverter,
		shardID:                shardID,
	}, nil
}

// GetAccounts will get accounts for regular operations and mect operations
func (ap *accountsProcessor) GetAccounts(alteredAccounts data.AlteredAccountsHandler) ([]*data.Account, []*data.AccountMECT) {
	regularAccountsToIndex := make([]*data.Account, 0)
	accountsToIndexMECT := make([]*data.AccountMECT, 0)

	if check.IfNil(alteredAccounts) {
		return regularAccountsToIndex, accountsToIndexMECT
	}

	allAlteredAccounts := alteredAccounts.GetAll()
	for address, altered := range allAlteredAccounts {
		userAccount, err := ap.getUserAccount(address)
		if err != nil || check.IfNil(userAccount) {
			log.Warn("cannot get user account", "address", address, "error", err)
			continue
		}

		regularAccounts, mectAccounts := splitAlteredAccounts(userAccount, altered)

		regularAccountsToIndex = append(regularAccountsToIndex, regularAccounts...)
		accountsToIndexMECT = append(accountsToIndexMECT, mectAccounts...)
	}

	return regularAccountsToIndex, accountsToIndexMECT
}

func splitAlteredAccounts(userAccount coreData.UserAccountHandler, altered []*data.AlteredAccount) ([]*data.Account, []*data.AccountMECT) {
	regularAccountsToIndex := make([]*data.Account, 0)
	accountsToIndexMECT := make([]*data.AccountMECT, 0)
	for _, info := range altered {
		if info.IsMECTOperation || info.IsNFTOperation {
			accountsToIndexMECT = append(accountsToIndexMECT, &data.AccountMECT{
				Account:         userAccount,
				TokenIdentifier: info.TokenIdentifier,
				IsSender:        info.IsSender,
				IsNFTOperation:  info.IsNFTOperation,
				NFTNonce:        info.NFTNonce,
				IsNFTCreate:     info.IsNFTCreate,
			})
		}

		// if the balance of the MECT receiver is 0 the receiver is a new account most probably, and we should index it
		ignoreReceiver := !info.BalanceChange && notZeroBalance(userAccount) && !info.IsSender
		if ignoreReceiver {
			continue
		}

		regularAccountsToIndex = append(regularAccountsToIndex, &data.Account{
			UserAccount: userAccount,
			IsSender:    info.IsSender,
		})
	}

	return regularAccountsToIndex, accountsToIndexMECT
}

func notZeroBalance(userAccount coreData.UserAccountHandler) bool {
	if userAccount.GetBalance() == nil {
		return false
	}

	return userAccount.GetBalance().Cmp(big.NewInt(0)) > 0
}

func (ap *accountsProcessor) getUserAccount(address string) (coreData.UserAccountHandler, error) {
	addressBytes, err := ap.addressPubkeyConverter.Decode(address)
	if err != nil {
		return nil, err
	}

	account, err := ap.accountsDB.LoadAccount(addressBytes)
	if err != nil {
		return nil, err
	}

	userAccount, ok := account.(coreData.UserAccountHandler)
	if !ok {
		return nil, indexer.ErrCannotCastAccountHandlerToUserAccount
	}

	return userAccount, nil
}

// PrepareRegularAccountsMap will prepare a map of regular accounts
func (ap *accountsProcessor) PrepareRegularAccountsMap(timestamp uint64, accounts []*data.Account) map[string]*data.AccountInfo {
	accountsMap := make(map[string]*data.AccountInfo)
	for _, userAccount := range accounts {
		address := ap.addressPubkeyConverter.Encode(userAccount.UserAccount.AddressBytes())
		balance := userAccount.UserAccount.GetBalance()
		balanceAsFloat := ap.balanceConverter.ComputeBalanceAsFloat(balance)
		acc := &data.AccountInfo{
			Address:                  address,
			Nonce:                    userAccount.UserAccount.GetNonce(),
			Balance:                  converters.BigIntToString(balance),
			BalanceNum:               balanceAsFloat,
			IsSender:                 userAccount.IsSender,
			IsSmartContract:          core.IsSmartContractAddress(userAccount.UserAccount.AddressBytes()),
			TotalBalanceWithStake:    converters.BigIntToString(balance),
			TotalBalanceWithStakeNum: balanceAsFloat,
			Timestamp:                time.Duration(timestamp),
			ShardID:                  ap.shardID,
		}

		accountsMap[address] = acc
	}

	return accountsMap
}

// PrepareAccountsMapMECT will prepare a map of accounts with MECT tokens
func (ap *accountsProcessor) PrepareAccountsMapMECT(
	timestamp uint64,
	accounts []*data.AccountMECT,
	tagsCount data.CountTags,
) (map[string]*data.AccountInfo, data.TokensHandler) {
	tokensData := data.NewTokensInfo()
	accountsMECTMap := make(map[string]*data.AccountInfo)
	for _, accountMECT := range accounts {
		address := ap.addressPubkeyConverter.Encode(accountMECT.Account.AddressBytes())
		balance, properties, tokenMetaData, err := ap.getMECTInfo(accountMECT)
		if err != nil {
			log.Warn("cannot get mect info from account",
				"address", address,
				"error", err.Error())
			continue
		}

		if tokenMetaData != nil && accountMECT.IsNFTCreate {
			tagsCount.ParseTags(tokenMetaData.Tags)
		}

		tokenIdentifier := converters.ComputeTokenIdentifier(accountMECT.TokenIdentifier, accountMECT.NFTNonce)
		acc := &data.AccountInfo{
			Address:         address,
			TokenName:       accountMECT.TokenIdentifier,
			TokenIdentifier: tokenIdentifier,
			TokenNonce:      accountMECT.NFTNonce,
			Balance:         balance.String(),
			BalanceNum:      ap.balanceConverter.ComputeMECTBalanceAsFloat(balance),
			Properties:      properties,
			IsSender:        accountMECT.IsSender,
			IsSmartContract: core.IsSmartContractAddress(accountMECT.Account.AddressBytes()),
			Data:            tokenMetaData,
			Timestamp:       time.Duration(timestamp),
			ShardID:         ap.shardID,
		}

		if acc.TokenNonce == 0 {
			acc.Type = core.FungibleMECT
		}

		keyInMap := fmt.Sprintf("%s-%s-%d", acc.Address, acc.TokenName, accountMECT.NFTNonce)
		accountsMECTMap[keyInMap] = acc

		if acc.Balance == "0" || acc.Balance == "" {
			continue
		}

		tokensData.Add(&data.TokenInfo{
			Token:      accountMECT.TokenIdentifier,
			Identifier: tokenIdentifier,
		})
	}

	return accountsMECTMap, tokensData
}

// PrepareAccountsHistory will prepare a map of accounts history balance from a map of accounts
func (ap *accountsProcessor) PrepareAccountsHistory(
	timestamp uint64,
	accounts map[string]*data.AccountInfo,
) map[string]*data.AccountBalanceHistory {
	accountsMap := make(map[string]*data.AccountBalanceHistory)
	for _, userAccount := range accounts {
		acc := &data.AccountBalanceHistory{
			Address:         userAccount.Address,
			Balance:         userAccount.Balance,
			Timestamp:       time.Duration(timestamp),
			Token:           userAccount.TokenName,
			TokenNonce:      userAccount.TokenNonce,
			IsSender:        userAccount.IsSender,
			IsSmartContract: userAccount.IsSmartContract,
			Identifier:      converters.ComputeTokenIdentifier(userAccount.TokenName, userAccount.TokenNonce),
			ShardID:         ap.shardID,
		}
		keyInMap := fmt.Sprintf("%s-%s-%d", acc.Address, acc.Token, acc.TokenNonce)
		accountsMap[keyInMap] = acc
	}

	return accountsMap
}

func (ap *accountsProcessor) getMECTInfo(accountMECT *data.AccountMECT) (*big.Int, string, *data.TokenMetaData, error) {
	if accountMECT.TokenIdentifier == "" {
		return big.NewInt(0), "", nil, nil
	}
	if accountMECT.IsNFTOperation && accountMECT.NFTNonce == 0 {
		return big.NewInt(0), "", nil, nil
	}

	tokenKey := computeTokenKey(accountMECT.TokenIdentifier, accountMECT.NFTNonce)
	valueBytes, err := accountMECT.Account.RetrieveValueFromDataTrieTracker(tokenKey)
	if err != nil {
		return nil, "", nil, err
	}

	mectToken := &mect.MECToken{}
	err = ap.internalMarshalizer.Unmarshal(mectToken, valueBytes)
	if err != nil {
		return nil, "", nil, err
	}

	if mectToken.Value == nil {
		return big.NewInt(0), "", nil, nil
	}

	if mectToken.TokenMetaData == nil && accountMECT.NFTNonce > 0 {
		metadata, errLoad := ap.loadMetadataFromSystemAccount(tokenKey)
		if errLoad != nil {
			return nil, "", nil, errLoad
		}

		mectToken.TokenMetaData = metadata
	}

	tokenMetaData := converters.PrepareTokenMetaData(ap.addressPubkeyConverter, mectToken)

	return mectToken.Value, hex.EncodeToString(mectToken.Properties), tokenMetaData, nil
}

// PutTokenMedataDataInTokens will put the TokenMedata in provided tokens data
func (ap *accountsProcessor) PutTokenMedataDataInTokens(tokensData []*data.TokenInfo) {
	for _, tokenData := range tokensData {
		if tokenData.Data != nil || tokenData.Nonce == 0 {
			continue
		}

		tokenKey := computeTokenKey(tokenData.Token, tokenData.Nonce)
		metadata, errLoad := ap.loadMetadataFromSystemAccount(tokenKey)
		if errLoad != nil {
			log.Warn("cannot load token metadata",
				"token identifier ", tokenData.Identifier,
				"error", errLoad.Error())

			continue
		}

		tokenData.Data = converters.PrepareTokenMetaData(ap.addressPubkeyConverter, &mect.MECToken{TokenMetaData: metadata})
	}
}

func (ap *accountsProcessor) loadMetadataFromSystemAccount(tokenKey []byte) (*mect.MetaData, error) {
	systemAccount, err := ap.accountsDB.LoadAccount(vmcommon.SystemAccountAddress)
	if err != nil {
		return nil, err
	}

	userAccount, ok := systemAccount.(coreData.UserAccountHandler)
	if !ok {
		return nil, indexer.ErrCannotCastAccountHandlerToUserAccount
	}

	marshaledData, err := userAccount.RetrieveValueFromDataTrieTracker(tokenKey)
	if err != nil {
		return nil, err
	}

	mectData := &mect.MECToken{}
	err = ap.internalMarshalizer.Unmarshal(mectData, marshaledData)
	if err != nil {
		return nil, err
	}

	return mectData.TokenMetaData, nil
}

func computeTokenKey(token string, nonce uint64) []byte {
	tokenKey := []byte(core.MotherEarthProtectedKeyPrefix + core.MECTKeyIdentifier + token)
	if nonce > 0 {
		nonceBig := big.NewInt(0).SetUint64(nonce)
		tokenKey = append(tokenKey, nonceBig.Bytes()...)
	}

	return tokenKey
}
