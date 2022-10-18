package transactions

import datafield "github.com/ME-MotherEarth/me-vm-common/parsers/dataField"

// DataFieldParser defines what a data field parser should be able to do
type DataFieldParser interface {
	Parse(dataField []byte, sender, receiver []byte) *datafield.ResponseParseData
}
