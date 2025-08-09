package v1

type PathInfo struct {
	IsValid    bool   `json:"is_valid"`
	CommonPath string `json:"common_path"`
	Files      []File `json:"files"`
}

type File struct {
	Name     string `json:"name"`
	Size     uint64 `json:"size"`
	IsFolder bool   `json:"is_folder,omitempty"`
}

type Report struct {
	BagID     string `json:"bag_id"`
	Reason    string `json:"reason"`
	Sender    string `json:"sender"`
	Comment   string `json:"comment"`
	CreatedAt uint64 `json:"created_at"`
}

type BanInfo struct {
	BagID   string `json:"bag_id"`
	Admin   string `json:"admin"`
	Reason  string `json:"reason"`
	Comment string `json:"comment"`
}

type BanStatus struct {
	BagID   string `json:"bag_id"`
	Admin   string `json:"admin"`
	Reason  string `json:"reason"`
	Comment string `json:"comment"`
	Status  bool   `json:"status"`
}
