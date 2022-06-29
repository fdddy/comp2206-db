package db

import (
	"time"

	"github.com/fentec-project/gofe/abe"
	"github.com/plzfgme/mfast"
)

type Record struct {
	UserId   string
	Location string
	Time     time.Time
	Set      string
}

type FindAResult struct {
	Location string
	Time     time.Time
}

type FindBResult struct {
	UserId string
}

type DelegatedKeys struct {
	Set       string
	MFastKeys *mfast.DelegatedKeys
	ABEAttrK  *abe.FAMEAttribKeys
	ABEPK     *abe.FAMEPubKey
}
