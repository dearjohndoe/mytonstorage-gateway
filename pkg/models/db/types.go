package db

import "time"

type Report struct {
	BagID     string     `json:"bag_id"`
	Reason    string     `json:"reason"`
	Sender    string     `json:"sender"`
	Comment   string     `json:"comment"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
}

type BanStatus struct {
	BagID     string     `json:"bag_id"`
	Admin     string     `json:"admin"`
	Reason    string     `json:"reason"`
	Comment   string     `json:"comment"`
	Status    bool       `json:"status"`
	CreatedAt *time.Time `json:"created_at,omitempty"`
}
