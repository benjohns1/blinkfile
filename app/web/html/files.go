package html

import (
	"fmt"
	domain "git.jfam.app/one-way-file-send"
	"github.com/kataras/iris/v12"
	"io"
	"time"
)

type (
	FilesView struct {
		LayoutView
		Files []FileView
	}
	FileView struct {
		ID       string
		Name     string
		Uploaded string
		Size     string
	}
)

func showFiles(ctx iris.Context, a App) error {
	owner := loggedInUser(ctx)
	files, err := a.ListFiles(ctx, owner)
	if err != nil {
		return err
	}
	fileList := make([]FileView, 0, len(files))
	for _, file := range files {
		fileList = append(fileList, FileView{
			ID:       string(file.ID),
			Name:     file.Name,
			Uploaded: file.Created.Format(time.RFC3339),
			Size:     formatFileSize(file.Size),
		})
	}
	ctx.ViewData("content", FilesView{
		Files: fileList,
	})
	return ctx.View("files.html")
}

func formatFileSize(size int64) string {
	const unit = 1024
	const labels = "KMGTPE"
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := unit, 0
	for n := size / unit; n >= unit || exp >= len(labels); n /= unit {
		div *= unit
		exp++
	}
	amount := float64(size) / float64(div)
	return fmt.Sprintf("%.2f %ciB", amount, labels[exp])
}

func uploadFile(ctx iris.Context, a App) error {
	owner := loggedInUser(ctx)
	file, header, err := ctx.FormFile("file")
	if err != nil {
		return err
	}
	err = a.UploadFile(ctx, header.Filename, owner, file, header.Size)
	if err != nil {
		return err
	}
	ctx.Redirect("/")
	return nil
}

func downloadFile(ctx iris.Context, a App) error {
	fileID := domain.FileID(ctx.Params().Get("file_id"))
	user := loggedInUser(ctx)
	file, err := a.DownloadFile(ctx, user, fileID)
	if err != nil {
		return err
	}
	defer func() { _ = file.Data.Close() }()
	ctx.Header("Content-Type", "application/octet-stream")
	ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%q", file.Name))
	_, err = io.Copy(ctx.ResponseWriter(), file.Data)
	return err
}
