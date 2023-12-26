package html

import (
	"errors"
	"fmt"
	domain "git.jfam.app/one-way-file-send"
	"git.jfam.app/one-way-file-send/app"
	"git.jfam.app/one-way-file-send/app/web"
	"github.com/kataras/iris/v12"
	"io"
	"net/http"
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
		ID                string
		Name              string
		Uploaded          string
		Expires           string
		Size              string
		PasswordProtected bool
	}
	FileDownloadView struct {
		LayoutView
		ID string
		MessageView
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
		var expires string
		if file.Expires.IsZero() {
			expires = "Never"
		} else {
			expires = file.Expires.Format(time.RFC3339)
		}
		fileList = append(fileList, FileView{
			ID:                string(file.ID),
			Name:              file.Name,
			Uploaded:          file.Created.Format(time.RFC3339),
			Expires:           expires,
			Size:              formatFileSize(file.Size),
			PasswordProtected: file.PasswordHash != "",
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
	file, header, err := ctx.FormFile("file")
	if err != nil {
		return app.Error{Type: app.ErrBadRequest, Err: err}
	}

	var expiresIn domain.LongDuration
	expireFutureAmount := ctx.FormValue("expire_future_amount")
	if expireFutureAmount != "" {
		expiresIn = domain.LongDuration(fmt.Sprintf("%s%s", expireFutureAmount, ctx.FormValue("expire_future_unit")))
	}
	args := app.UploadFileArgs{
		Filename:  header.Filename,
		Owner:     loggedInUser(ctx),
		Reader:    file,
		Size:      header.Size,
		Password:  ctx.FormValue("password"),
		ExpiresIn: expiresIn,
	}
	err = a.UploadFile(ctx, args)
	if err != nil {
		return err
	}
	ctx.Redirect("/")
	return nil
}

func downloadFile(ctx iris.Context, a App) error {
	fileID := domain.FileID(ctx.Params().Get("file_id"))
	view := FileDownloadView{
		ID: string(fileID),
	}
	err := func() error {
		user := loggedInUser(ctx)
		password := ctx.FormValue("password")
		file, err := a.DownloadFile(ctx, user, fileID, password)
		if err != nil {
			return err
		}
		defer func() { _ = file.Data.Close() }()
		ctx.Header("Content-Type", "application/octet-stream")
		ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%q", file.Name))
		_, err = io.Copy(ctx.ResponseWriter(), file.Data)
		return err
	}()
	if err != nil {
		errID, _, _ := web.ParseAppErr(err)
		web.LogError(ctx, errID, err)
		if errors.Is(err, domain.ErrFilePasswordRequired) {
			view.MessageView.SuccessMessage = "Password required"
		} else if errors.Is(err, domain.ErrFilePasswordInvalid) {
			view.ErrorView = ErrorView{
				ID:      errID,
				Status:  http.StatusUnauthorized,
				Message: "Invalid password",
			}
		}
		ctx.ViewData("content", view)
		return ctx.View("file.html")
	}
	return nil
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
