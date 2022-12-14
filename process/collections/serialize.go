package collections

import (
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/ME-MotherEarth/me-core/core"
	"github.com/ME-MotherEarth/me-elastic-indexer/converters"
	"github.com/ME-MotherEarth/me-elastic-indexer/data"
)

// ExtractAndSerializeCollectionsData will extra the accounts with NFT/SFT and serialize
func ExtractAndSerializeCollectionsData(
	accountsMECT map[string]*data.AccountInfo,
	buffSlice *data.BufferSlice,
	index string,
) error {
	for _, acct := range accountsMECT {
		shouldIgnore := acct.Type != core.NonFungibleMECT && acct.Type != core.SemiFungibleMECT
		if shouldIgnore {
			if acct.Balance != "0" || acct.TokenNonce == 0 {
				continue
			}
		}

		nonceBig := big.NewInt(0).SetUint64(acct.TokenNonce)
		hexEncodedNonce := hex.EncodeToString(nonceBig.Bytes())

		meta := []byte(fmt.Sprintf(`{ "update" : {"_index":"%s", "_id" : "%s" } }%s`, index, converters.JsonEscape(acct.Address), "\n"))
		codeToExecute := `
			if (('create' == ctx.op) && ('0' == params.value)) {
				ctx.op = 'noop';
			} else if ('0' != params.value) {
				if (!ctx._source.containsKey(params.col)) {
					ctx._source[params.col] = new HashMap();
				}
				ctx._source[params.col][params.nonce] = params.value
			} else {
				if (ctx._source.containsKey(params.col)) {
					ctx._source[params.col].remove(params.nonce);
					if (ctx._source[params.col].size() == 0) {
						ctx._source.remove(params.col)
					}
					if (ctx._source.size() == 0) {
						ctx.op = 'delete';
					}
				}
			}
`

		tokenName := converters.JsonEscape(acct.TokenName)
		tokenNonceHex := converters.JsonEscape(hexEncodedNonce)
		balanceStr := converters.JsonEscape(acct.Balance)

		collection := fmt.Sprintf(`{"%s":{"%s": "%s"}}`,
			tokenName,
			tokenNonceHex,
			balanceStr,
		)
		serializedDataStr := fmt.Sprintf(`{"scripted_upsert": true, "script": {`+
			`"source": "%s",`+
			`"lang": "painless",`+
			`"params": { "col": "%s", "nonce": "%s", "value": "%s"}},`+
			`"upsert": %s}`,
			converters.FormatPainlessSource(codeToExecute), tokenName, tokenNonceHex, balanceStr, collection)

		err := buffSlice.PutData(meta, []byte(serializedDataStr))
		if err != nil {
			return err
		}
	}

	return nil
}
