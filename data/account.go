package data

import (
	"time"

	coreData "github.com/ME-MotherEarth/me-core/data"
)

// AccountInfo holds (serializable) data about an account
type AccountInfo struct {
	Address                  string         `json:"address,omitempty"`
	Nonce                    uint64         `json:"nonce,omitempty"`
	Balance                  string         `json:"balance"`
	BalanceNum               float64        `json:"balanceNum"`
	TokenName                string         `json:"token,omitempty"`
	TokenIdentifier          string         `json:"identifier,omitempty"`
	TokenNonce               uint64         `json:"tokenNonce,omitempty"`
	Properties               string         `json:"properties,omitempty"`
	TotalBalanceWithStake    string         `json:"totalBalanceWithStake,omitempty"`
	TotalBalanceWithStakeNum float64        `json:"totalBalanceWithStakeNum,omitempty"`
	Data                     *TokenMetaData `json:"data,omitempty"`
	Timestamp                time.Duration  `json:"timestamp,omitempty"`
	Type                     string         `json:"type,omitempty"`
	CurrentOwner             string         `json:"currentOwner,omitempty"`
	ShardID                  uint32         `json:"shardID"`
	IsSender                 bool           `json:"-"`
	IsSmartContract          bool           `json:"-"`
	IsNFTCreate              bool           `json:"-"`
}

// TokenMetaData holds data about a token metadata
type TokenMetaData struct {
	Name               string   `json:"name,omitempty"`
	Creator            string   `json:"creator,omitempty"`
	Royalties          uint32   `json:"royalties,omitempty"`
	Hash               []byte   `json:"hash,omitempty"`
	URIs               [][]byte `json:"uris,omitempty"`
	Tags               []string `json:"tags,omitempty"`
	Attributes         []byte   `json:"attributes,omitempty"`
	MetaData           string   `json:"metadata,omitempty"`
	NonEmptyURIs       bool     `json:"nonEmptyURIs"`
	WhiteListedStorage bool     `json:"whiteListedStorage"`
}

// AccountBalanceHistory represents an entry in the user accounts balances history
type AccountBalanceHistory struct {
	Address         string        `json:"address"`
	Timestamp       time.Duration `json:"timestamp"`
	Balance         string        `json:"balance"`
	Token           string        `json:"token,omitempty"`
	Identifier      string        `json:"identifier,omitempty"`
	TokenNonce      uint64        `json:"tokenNonce,omitempty"`
	IsSender        bool          `json:"isSender,omitempty"`
	IsSmartContract bool          `json:"isSmartContract,omitempty"`
	ShardID         uint32        `json:"shardID"`
}

// Account is a structure that is needed for regular accounts
type Account struct {
	UserAccount coreData.UserAccountHandler
	IsSender    bool
}

// AccountMECT is a structure that is needed for MECT accounts
type AccountMECT struct {
	Account         coreData.UserAccountHandler
	TokenIdentifier string
	NFTNonce        uint64
	IsSender        bool
	IsNFTOperation  bool
	IsNFTCreate     bool
}
