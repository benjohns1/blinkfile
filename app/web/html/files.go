package html

import (
	"fmt"
	domain "git.jfam.app/one-way-file-send"
	"git.jfam.app/one-way-file-send/app"
	"github.com/kataras/iris/v12"
	"io"
	"strings"
	"time"
)

type (
	FilesView struct {
		LayoutView
		Files []FileView
		MessageView
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
	for n := size / unit; n >= unit && exp < len(labels); n /= unit {
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
		return app.Error{Type: app.ErrBadRequest, Err: err}
	}
	password := ctx.FormValue("password")
	err = a.UploadFile(ctx, header.Filename, owner, file, header.Size, password)
	if err != nil {
		return err
	}
	ctx.Redirect("/")
	return nil
}

func downloadFile(ctx iris.Context, a App) error {
	fileID := domain.FileID(ctx.Params().Get("file_id"))
	user := loggedInUser(ctx)
	password := "" // TODO: capture password
	file, err := a.DownloadFile(ctx, user, fileID, password)
	if err != nil {
		return err
	}
	defer func() { _ = file.Data.Close() }()
	ctx.Header("Content-Type", "application/octet-stream")
	ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%q", file.Name))
	_, err = io.Copy(ctx.ResponseWriter(), file.Data)
	return err
}

func deleteFiles(ctx iris.Context, a App) error {
	owner := loggedInUser(ctx)
	req := ctx.Request()
	err := req.ParseForm()
	if err != nil {
		return err
	}
	deleteFileIDs := make([]domain.FileID, 0, len(req.Form))
	for name, values := range req.Form {
		if len(values) == 0 || values[0] != "on" {
			continue
		}
		deleteFileIDs = append(deleteFileIDs, domain.FileID(strings.TrimPrefix(name, "select-")))
	}
	if len(deleteFileIDs) > 0 {
		err = a.DeleteFiles(ctx, owner, deleteFileIDs)
		if err != nil {
			return err
		}
	}

	ctx.Redirect("/")
	return nil
}
