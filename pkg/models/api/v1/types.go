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
