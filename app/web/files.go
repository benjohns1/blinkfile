package web

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/benjohns1/blinkfile"
	"github.com/benjohns1/blinkfile/app"
	"github.com/benjohns1/blinkfile/longduration"
	"github.com/kataras/iris/v12"
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
		ByteSize          int64
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
			ByteSize:          file.Size,
			Size:              formatFileSize(file.Size),
			PasswordProtected: file.PasswordHash != "",
		})
	}
	ctx.ViewData("content", FilesView{
		Files:       fileList,
		MessageView: flashMessageView(ctx),
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
		return app.ErrUser("Invalid file.", "We couldn't retrieve the uploaded file, please try again.", err)
	}

	var expiresIn longduration.LongDuration
	expireFutureAmount := ctx.FormValue("expire_future_amount")
	if expireFutureAmount != "" {
		expiresIn = longduration.LongDuration(fmt.Sprintf("%s%s", expireFutureAmount, ctx.FormValue("expire_future_unit")))
	}
	var expires time.Time
	expirationTime := ctx.FormValue("expiration_time")
	if expirationTime != "" {
		expires, err = time.Parse(time.RFC3339, expirationTime)
		if err != nil {
			err = fmt.Errorf("parsing expiration time %q: %w", expirationTime, err)
			return app.ErrUser("Invalid expiration time.", fmt.Sprintf("We couldn't understand the file expiration time %q, please make sure the date format is correct.", expirationTime), err)
		}
	}
	args := app.UploadFileArgs{
		Filename:  header.Filename,
		Owner:     loggedInUser(ctx),
		Reader:    file,
		Size:      header.Size,
		Password:  ctx.FormValue("password"),
		ExpiresIn: expiresIn,
		Expires:   expires,
	}
	err = a.UploadFile(ctx, args)
	if err != nil {
		return err
	}
	setFlashSuccess(ctx, fmt.Sprintf("Successfully uploaded %s", header.Filename))
	ctx.Redirect("/")
	return nil
}

func sanitizeFilename(in string) string {
	return strings.ReplaceAll(in, ";", "_")
}

func downloadFile(ctx iris.Context, a App) error {
	fileID := blinkfile.FileID(ctx.Params().Get("file_id"))
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
		ctx.Header("Content-Disposition", fmt.Sprintf("attachment; filename*=UTF-8''%s", sanitizeFilename(file.Name)))
		err = ctx.ServeFileWithRate(file.Location, 0, 0)
		if err != nil {
			return fmt.Errorf("sending file data: %w", err)
		}
		return nil
	}()
	if err != nil {
		if errors.Is(err, blinkfile.ErrFilePasswordRequired) {
			view.MessageView.SuccessMessage = "Password required"
		} else if errors.Is(err, blinkfile.ErrFilePasswordInvalid) {
			errView := ParseAppErr(ctx, a, err)
			errView.Detail = "Invalid password"
			view.ErrorView = errView
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
	deleteFileIDs := make([]blinkfile.FileID, 0, len(req.Form))
	for name, values := range req.Form {
		if len(values) == 0 || values[0] != "on" {
			continue
		}
		deleteFileIDs = append(deleteFileIDs, blinkfile.FileID(strings.TrimPrefix(name, "select-")))
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
