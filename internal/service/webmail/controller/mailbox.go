package controller

import (
	"archive/zip"
	"bytes"
	"easymail/internal/model"
	"encoding/base64"
	"fmt"
	"github.com/gin-gonic/gin"
	"html/template"
	"io"
	"net/http"
	"net/mail"
	"strconv"
	"strings"
	"time"
)

type MailboxController struct{}

var FolderNameMapper = []string{
	"inbox",
	"draft",
	"trash",
	"spam",
	"quarantine",
}

func GetFolderName(folderID int) string {
	if folderID < len(FolderNameMapper) {
		return FolderNameMapper[folderID]
	}
	return "unKnown"
}

func GetFolderID(name string) int64 {
	for i, f := range FolderNameMapper {
		if name == f {
			return int64(i)
		}
	}
	return -1
}

func IsFolderName(folder string) bool {
	for _, f := range FolderNameMapper {
		if f == folder {
			return true
		}
	}
	return false
}

func (m *MailboxController) Index(c *gin.Context) {
	folder := c.Param("folder")
	folder = strings.ToLower(folder)
	// check FolderNameMapper contain folder
	if !IsFolderName(folder) {
		c.Redirect(http.StatusFound, "/mailbox/inbox")
		return
	}
	folderID := GetFolderID(folder)
	if folderID < 0 {
		c.Redirect(http.StatusFound, "/mailbox/inbox")
		return
	}

	c.HTML(200, "mailbox_index.html", gin.H{
		"mailbox":  GetMailbox(c),
		"folderID": folderID,
		"module":   "mailbox",
		"page":     folder,
	})

}

type FolderIndexRequest struct {
	DataTableRequest
}

type FolderIndexResponse struct {
	ID         int64     `json:"id"`
	Subject    string    `json:"subject"`
	Digest     string    `json:"digest"`
	Sender     string    `json:"sender"`
	Size       int64     `json:"size"`
	MailTime   time.Time `json:"mailTime"`
	ReadStatus uint8     `json:"readStatus"`
}

func (m *MailboxController) Folder(c *gin.Context) {
	folder := c.Param("folderID")
	folderID, err := strconv.ParseInt(folder, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"error": err.Error(),
		})
	}
	accID, err := GetAccountID(c)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"error": err.Error(),
		})
	}

	var req FolderIndexRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"error": err.Error()})
		return
	}
	orderField := ""
	orderDir := ""
	for _, o := range req.Orders {
		orderField = CamelFieldName(req.Columns[o.Column].Data)
		orderDir = o.Dir
		break
	}

	total, news, mails, err := localStorage.Query(accID, folderID, orderField, orderDir, req.Start, req.Length)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"error": err.Error()})
		return
	}
	data := make([]FolderIndexResponse, 0)
	for _, mail := range mails {
		digest := ""
		data = append(data, FolderIndexResponse{
			ID:         mail.ID,
			Subject:    mail.Subject,
			Digest:     digest,
			Sender:     mail.Sender,
			Size:       mail.Size,
			MailTime:   mail.MailTime,
			ReadStatus: uint8(mail.ReadStatus),
		})
	}
	c.JSON(200, gin.H{
		"folder":          GetFolderName(int(folderID)),
		"news":            news,
		"draw":            req.Draw,
		"recordsTotal":    total,
		"recordsFiltered": total,
		"data":            data,
	})
}

type BatchMailRequest struct {
	IDS []int64 `json:"ids"`
}

func (m *MailboxController) MarkRead(c *gin.Context) {
	accID, err := GetAccountID(c)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"error": err.Error(),
		})
	}

	var req BatchMailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"error": err.Error()})
		return
	}
	data := make([]string, 0)

	for _, id := range req.IDS {
		err = model.MarkRead(accID, id, model.WebRead)
		if err != nil {
			data = append(data, err.Error())
		}
		data = append(data, fmt.Sprintf("mark %d OK", id))
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": strings.Join(data, "\n")})
}

func (m *MailboxController) DeleteMails(c *gin.Context) {
	accID, err := GetAccountID(c)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"error": err.Error(),
		})
	}

	var req BatchMailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"error": err.Error()})
		return
	}
	data := make([]string, 0)

	for _, id := range req.IDS {
		err = model.MoveMail(accID, id, model.Trash)
		if err != nil {
			data = append(data, err.Error())
		}
		data = append(data, fmt.Sprintf("move mail %d to trash OK", id))
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": strings.Join(data, "\n")})
}

/*
AttachmentDisplay
for display in the read page
*/
type AttachmentDisplay struct {
	Name        string
	ContentType string
	Size        int64
	SizeAlias   string
	Icon        string
}

var AttachIconMapper = map[string]string{
	"image/*":                  "fa-file-image",
	"application/octet-stream": "fa-file-archive",
}

func getIcon(contentType string) string {
	d := strings.SplitN(contentType, "/", 2)
	if len(d) == 2 {
		if icon, ok := AttachIconMapper[d[0]+"/*"]; ok {
			return icon
		}
	}
	if icon, ok := AttachIconMapper[contentType]; ok {
		return icon
	}
	return "fa-file-archive"
}

func (m *MailboxController) Read(c *gin.Context) {
	mids := c.Param("mid")
	mid, err := strconv.ParseInt(mids, 10, 64)
	if err != nil {
		c.HTML(http.StatusOK, "single_error.html", gin.H{"error": err.Error()})
		return
	}

	accID, err := GetAccountID(c)
	if err != nil {
		c.HTML(http.StatusOK, "single_error.html", gin.H{"error": err.Error()})
		return
	}

	mail, err := model.GetMail(accID, mid)
	if err != nil {
		c.HTML(http.StatusOK, "single_error.html", gin.H{"error": err.Error()})
		return
	}

	mailContent, err := localStorage.Read(mail.SavePath)
	if err != nil {
		c.HTML(http.StatusOK, "single_error.html", gin.H{"error": err.Error()})
		return
	}
	html := ""
	if mailContent.Html != "" {
		html = mailContent.Html
	} else if mailContent.Text != "" {
		html = "<pre>" + mailContent.Text + "</pre>"
	}
	base64Content := base64.StdEncoding.EncodeToString([]byte(html))

	// mark read mail
	err = model.MarkRead(accID, mid, model.WebRead)
	if err != nil {
		c.HTML(http.StatusOK, "single_error.html", gin.H{"error": err.Error()})
		return
	}

	attachment := make([]AttachmentDisplay, 0)
	for _, a := range mailContent.Attaches {
		attachment = append(attachment, AttachmentDisplay{
			Name:        a.Name,
			ContentType: a.ContentType,
			Size:        a.Size,
			SizeAlias:   FormatNumber(float64(a.Size), 2),
			Icon:        getIcon(a.ContentType),
		})
	}

	c.HTML(http.StatusOK, "mailbox_read.html", gin.H{
		"mailbox":  GetMailbox(c),
		"mail":     mail,
		"sender":   mailContent.Sender,
		"receipts": mailContent.Recipient,
		"html":     template.HTML(mailContent.Html),
		"b64src":   base64Content,
		"attaches": attachment,
	})
}

type AttachRequest struct {
	FileName string `form:"fileName"`
	All      bool   `form:"all"`
}

func (m *MailboxController) DownloadAttach(c *gin.Context) {
	userID, err := GetAccountID(c)
	if err != nil {
		c.HTML(http.StatusOK, "single_error.html", gin.H{"error": err.Error()})
		return
	}

	mids := c.Param("mid")
	mid, err := strconv.ParseInt(mids, 10, 64)
	if err != nil {
		c.HTML(http.StatusOK, "single_error.html", gin.H{"error": err.Error()})
		return
	}
	var req AttachRequest
	if err := c.ShouldBind(&req); err != nil {
		c.HTML(http.StatusOK, "single_error.html", gin.H{"error": err.Error()})
		return
	}
	if err != nil {
		c.HTML(http.StatusOK, "single_error.html", gin.H{"error": err.Error()})
		return
	}
	fileName := req.FileName
	if !req.All && fileName == "" {
		c.HTML(http.StatusOK, "single_error.html", gin.H{"error": "file name is empty"})
		return
	}

	mail, err := model.GetMail(userID, mid)
	data, err := localStorage.GetAttach(mail.SavePath, fileName, req.All)
	if err != nil {
		c.HTML(http.StatusOK, "single_error.html", gin.H{"error": err.Error()})
		return
	}

	// write data to client as attachment
	if req.All {
		distName := "attach.zip"
		c.Writer.Header().Set("Content-type", "application/octet-stream")
		c.Writer.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", distName))
		ar := zip.NewWriter(c.Writer)
		for _, att := range data {
			f, _ := ar.Create(att.Name)
			io.Copy(f, bytes.NewReader(att.Data))
		}
		ar.Close()
	} else {
		if len(data) > 0 {
			c.Writer.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", data[0].Name))
			c.Data(http.StatusOK, "application/octet-stream", data[0].Data)
		}
	}
}

type WriteMailRequest struct {
	Receipt string `form:"receipt" binding:"required,email"`
	Subject string `form:"subject" binding:"required,min=2"`
	Content string `form:"hidContent"`
}

func (m *MailboxController) Write(c *gin.Context) {
	accID, err := GetAccountID(c)
	if err != nil {
		c.HTML(http.StatusOK, "single_error.html", gin.H{"error": err.Error()})
		return
	}
	mailbox := GetMailbox(c)

	if c.Request.Method == "GET" {
		c.HTML(http.StatusOK, "mailbox_write.html", gin.H{
			"accountID": accID,
			"mailbox":   GetMailbox(c),
			"module":    "mailbox",
			"page":      "write",
		})
	} else if c.Request.Method == "POST" {
		req := WriteMailRequest{}
		if err := c.ShouldBind(&req); err != nil {
			c.HTML(http.StatusOK, "single_error.html", gin.H{"error": err.Error()})
			return
		}

		form, _ := c.MultipartForm()
		files := form.File["attach"]
		attaches := make([]model.Attachment, 0)
		for _, file := range files {
			f, err := file.Open()
			if err != nil {
				c.AbortWithError(http.StatusInternalServerError, err)
				f.Close()
				return
			}

			buf := new(bytes.Buffer)
			_, err = buf.ReadFrom(f)
			if err != nil {
				c.AbortWithError(http.StatusInternalServerError, err)
				return
			}
			f.Close()

			attaches = append(attaches, model.Attachment{
				Name:        file.Filename,
				ContentType: file.Header.Get("Content-Type"),
				Data:        buf.Bytes(),
			})
		}

		senderAddress := mail.Address{
			Name:    "",
			Address: mailbox,
		}
		receiptAddress := mail.Address{
			Name:    "",
			Address: req.Receipt,
		}
		receiptAddresses := make([]mail.Address, 0)
		receiptAddresses = append(receiptAddresses, receiptAddress)

		data, err := model.CreateMail(senderAddress, receiptAddresses, req.Subject, "", req.Content, attaches)
		if err != nil {
			c.HTML(http.StatusOK, "single_error.html", gin.H{"error": err.Error()})
			return
		}
		err = SendMailOpenRelay("localhost", mailbox, []string{req.Receipt}, req.Subject, data)

		c.Redirect(http.StatusFound, "/mailbox/done")
	}
}

func (m *MailboxController) Done(c *gin.Context) {
	c.HTML(http.StatusOK, "mailbox_done.html", gin.H{})
}
