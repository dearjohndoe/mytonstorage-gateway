package htmlTemplates

import (
	"html/template"
	"path/filepath"
	"slices"
	"strings"

	"mytonstorage-gateway/pkg/models/private"
	"mytonstorage-gateway/pkg/utils"
)

const APIBase = "/gateway"

var (
	imageFormats      = []string{".jpg", ".jpeg", ".png", ".gif", ".webp", ".svg", ".bmp", ".tiff", ".ico", ".avif"}
	videoFormats      = []string{".mp4", ".mkv", ".webm", ".avi", ".mov", ".flv", ".wmv"}
	audioFormats      = []string{".mp3", ".wav", ".ogg", ".flac", ".aac", ".m4a"}
	executableFormats = []string{".exe", ".bat", ".sh", ".cmd"}
)

type htmlTemplates struct {
	templates *template.Template
}

type Templates interface {
	ErrorTemplate(err error) (string, error)
	HtmlFilesListWithTemplate(f private.FolderInfo, path string) (string, error)
}

func (t *htmlTemplates) ErrorTemplate(err error) (string, error) {
	return t.renderTemplate("error.html", TemplateError{Error: err.Error()})
}

func (t *htmlTemplates) HtmlFilesListWithTemplate(f private.FolderInfo, path string) (string, error) {
	if !f.IsValid {
		return t.renderTemplate("error.html", TemplateError{Error: "Invalid path"})
	}

	data := TemplateData{
		Title:    filepath.Join(f.BagID, path),
		FullPath: filepath.Join(f.BagID, path),
		Files:    make([]FileData, 0, len(f.Files)),
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
		} else if slices.Contains(executableFormats, ext) {
			icon = "⚙️"
		}

		var href string
		if path == "" {
			href = filepath.Join(APIBase, f.BagID, file.Name)
		} else {
			href = filepath.Join(APIBase, f.BagID, path, file.Name)
		}

		data.Files = append(data.Files, FileData{
			Name:          file.Name,
			Href:          href,
			Icon:          icon,
			FormattedSize: utils.FormatSize(file.Size),
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
