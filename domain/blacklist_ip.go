package domain

import "github.com/kamva/mgm/v3"

type BlacklistIp struct {
	mgm.DefaultModel `bson:",inline"`
	IpAddress        string `bson:"ip_address" json:"ip_address"`
	Blocked          bool   `bson:"blocked" json:"blocked"`
}
