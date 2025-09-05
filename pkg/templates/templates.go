package htmlTemplates

import (
	"html/template"
	"path/filepath"
	"slices"
	"strings"

	"mytonstorage-gateway/pkg/models/private"
	"mytonstorage-gateway/pkg/utils"
)

const APIBase = "/api/v1/gateway"

var (
	imageFormats = []string{".jpg", ".jpeg", ".png", ".gif", ".webp", ".svg", ".bmp", ".tiff", ".ico", ".avif"}
	videoFormats = []string{".mp4", ".mkv", ".webm", ".avi", ".mov", ".flv", ".wmv"}
	audioFormats = []string{".mp3", ".wav", ".ogg", ".flac", ".aac", ".m4a"}
	textFormats  = []string{".txt", ".md", ".csv", ".json", ".xml", ".html", ".css", ".js", ".ts", ".py", ".go", ".java", ".c", ".cpp", ".h", ".php", ".rb", ".sh"}
)

type htmlTemplates struct {
	templates *template.Template
}

type Templates interface {
	ContentType(ext, filename string) (string, string)
	ErrorTemplate(err error) (string, error)
	HtmlFilesListWithTemplate(f private.FolderInfo, path string) (string, error)
}

func (t *htmlTemplates) ContentType(ext, filename string) (header, value string) {
	if slices.Contains(imageFormats, ext) {
		return "Content-Type", "image/" + strings.TrimPrefix(ext, ".")
	} else if slices.Contains(videoFormats, ext) {
		return "Content-Type", "video/" + strings.TrimPrefix(ext, ".")
	} else if slices.Contains(audioFormats, ext) {
		return "Content-Type", "audio/" + strings.TrimPrefix(ext, ".")
	} else if slices.Contains(textFormats, ext) {
		return "Content-Type", "text/" + strings.TrimPrefix(ext, ".")
	}

	return "Content-Disposition", `attachment; filename="` + filename + `"`
}

func (t *htmlTemplates) ErrorTemplate(err error) (string, error) {
	return t.renderTemplate("error.html", TemplateError{Error: err.Error()})
}

func (t *htmlTemplates) HtmlFilesListWithTemplate(f private.FolderInfo, path string) (string, error) {
	if !f.IsValid {
		return t.renderTemplate("error.html", TemplateError{Error: "Invalid path"})
	}

	f.BagID = strings.ToUpper(f.BagID)

	data := TemplateData{
		Title:       filepath.Join(f.BagID, path),
		FullPath:    filepath.Join(f.BagID, path),
		TotalSize:   utils.FormatSize(f.TotalSize),
		Description: f.Description,
		PeersCount:  f.PeersCount,
		FileCount:   len(f.Files),
		Files:       make([]FileData, 0, len(f.Files)),
	}

	if path != "" {
		parentPath := filepath.Dir(path)
		var parentHref string
		if parentPath == "." || parentPath == "" {
			parentHref = filepath.Join(APIBase, f.BagID)
		} else {
			parentHref = filepath.Join(APIBase, f.BagID, parentPath)
		}
		data.ParentDir = &ParentDirData{Href: parentHref}
	}

	for _, file := range f.Files {
		icon := "📄"
		ext := filepath.Ext(file.Name)
		if file.IsFolder {
			icon = "📁"
		} else if slices.Contains(imageFormats, ext) {
			icon = "🖼️"
		} else if slices.Contains(videoFormats, ext) {
			icon = "🎥"
		} else if slices.Contains(audioFormats, ext) {
			icon = "🎵"
		} else if slices.Contains(textFormats, ext) {
			icon = "📝"
		}

		var href string
		if path == "" {
			href = filepath.Join(APIBase, f.BagID, file.Name)
		} else {
			href = filepath.Join(APIBase, f.BagID, path, file.Name)
		}

		fileSize := utils.FormatSize(file.Size)
		if file.Size == 0 {
			fileSize = "folder"
		}

		data.Files = append(data.Files, FileData{
			Name:          file.Name,
			Href:          href,
			Icon:          icon,
			FormattedSize: fileSize,
		})
	}

	return t.renderTemplate("file_list.html", &data)
}

func (t *htmlTemplates) renderTemplate(templateName string, data any) (string, error) {
	var buf strings.Builder
	err := t.templates.ExecuteTemplate(&buf, templateName, data)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

func New(templatesPath string) (Templates, error) {
	templates, err := template.ParseGlob(filepath.Join(templatesPath, "*.html"))
	if err != nil {
		return nil, err
	}

	return &htmlTemplates{
		templates: templates,
	}, nil
}
