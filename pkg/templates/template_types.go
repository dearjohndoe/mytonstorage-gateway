package htmlTemplates

type TemplateData struct {
	Title       string
	FullPath    string
	TotalSize   string
	Description string
	PeersCount  int
	FileCount   int
	ParentDir   *ParentDirData
	Files       []FileData
}

type ContentType struct {
	Header     string
	Value      string
	IsDownload bool
	IsHtml     bool
}

type ParentDirData struct {
	Href string
}

type FileData struct {
	Name          string
	Href          string
	Icon          string
	FormattedSize string
}
