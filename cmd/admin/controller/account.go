package controller

import (
	"context"
	"easymail/internal/account"
	"easymail/internal/database"
	"fmt"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type AccountController struct{}

type DomainIndexRequest struct {
	DataTableRequest
	Keyword string `json:"keyword"`
}

type DomainIndexResponse struct {
	ID           int64     `json:"id"`
	Name         string    `json:"name"`
	TotalAccount int       `json:"totalAccount"`
	MX           string    `json:"mx"`
	SPF          string    `json:"spf"`
	DMARC        string    `json:"dmarc"`
	Status       int       `json:"status"`
	CreateTime   time.Time `json:"createTime"`
}

func (a *AccountController) IndexDomain(ctx *gin.Context) {
	if ctx.Request.Method == "GET" {
		sess := sessions.Default(ctx)
		username := sess.Get("userName")

		ctx.HTML(http.StatusOK, "domain_index.html", gin.H{
			"title":    "Domain Management - Easymail",
			"username": username,
		})
		return
	} else if ctx.Request.Method == "POST" {
		var req DomainIndexRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusOK, gin.H{"error": err.Error()})
			return
		}

		orderField := ""
		orderDir := ""
		for _, o := range req.Orders {
			orderField = req.Columns[o.Column].Data
			orderDir = o.Dir
			break
		}

		// 执行数据库查询，获取数据列表
		total, domains, err := account.DomainIndex(req.Keyword, orderField, orderDir, req.Start, req.Length)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "search domain error"})
			return
		}

		data := make([]DomainIndexResponse, 0)

		for _, domain := range domains {
			status := 0
			if !domain.Active {
				if domain.Deleted {
					status = 2
				} else {
					status = 1
				}
			}
			totalAccount, err := account.CountDomainAccount(domain.ID)
			if err != nil {
				log.Println("count domain account error:", err)
			}

			// dns resolve
			rdb := database.GetRedisClient()
			c := context.Background()
			expireTime := time.Duration(30 * 24 * time.Hour)

			// mx
			mxKey := fmt.Sprintf("mx:%s", domain.Name)
			mx, err := rdb.Get(c, mxKey).Result()
			if err == redis.Nil {
				s := make([]string, 0)
				s, err = resolver.LookupMX(domain.Name)
				if err != nil {
					log.Println(err)
				} else {
					mx = strings.Join(s, "<br>\n")
					rdb.Set(c, mxKey, mx, expireTime)
				}
			} else if err != nil {
				log.Println(err)
			}

			// spf
			spfKey := fmt.Sprintf("spf:%s", domain.Name)
			spf, err := rdb.Get(c, spfKey).Result()
			if err == redis.Nil {
				s := make([]string, 0)
				s, err = resolver.LookupSPF(domain.Name)
				if err != nil {
					log.Println(err)
				} else {
					spf = strings.Join(s, "<br>\n")
					rdb.Set(c, spfKey, spf, expireTime)
				}
			} else if err != nil {
				log.Println(err)
			}

			// dmarc
			dmarcKey := fmt.Sprintf("dmarc:%s", domain.Name)
			dmarc, err := rdb.Get(c, dmarcKey).Result()
			if err == redis.Nil {
				s := make([]string, 0)
				s, err = resolver.LookupDMARC(domain.Name)
				if err != nil {
					log.Println(err)
				} else {
					dmarc = strings.Join(s, "<br>\n")
					rdb.Set(c, dmarcKey, dmarc, expireTime)
				}
			} else if err != nil {
				log.Println(err)
			}

			data = append(data, DomainIndexResponse{
				ID:           domain.ID,
				Name:         domain.Name,
				TotalAccount: int(totalAccount),
				MX:           mx,
				SPF:          spf,
				DMARC:        dmarc,
				Status:       status,
				CreateTime:   domain.CreateTime,
			})
		}

		ctx.JSON(http.StatusOK, gin.H{
			"draw":            req.Draw,
			"recordsTotal":    total,
			"recordsFiltered": total,
			"data":            data,
		})
	}
}

func (a *AccountController) ToggleDomainActive(ctx *gin.Context) {
	id := ctx.Query("did")
	if id == "" {
		ctx.JSON(http.StatusOK, gin.H{"error": "invalid domain id"})
		return
	}

	domainID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{"error": "invalid domain id"})
		return
	}

	err = account.ToggleDomainActive(domainID)

	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{"error": "toggle domain active error"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"success": "toggle domain active success"})
}

func (a *AccountController) DeleteDomain(ctx *gin.Context) {
	id := ctx.Query("did")
	if id == "" {
		ctx.JSON(http.StatusOK, gin.H{"error": "invalid domain id"})
		return
	}

	domainID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{"error": "invalid domain id"})
		return
	}

	err = account.DeleteDomain(domainID)

	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{"error": "delete domain error"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"success": "delete domain success"})
}

type CreateDomainRequest struct {
	//ID          int    `json:"id"`
	Name        string `json:"domainName" binding:"required,min=3,max=255"`
	Description string `json:"description" binding:"required,min=3,max=255"`
}

func (a *AccountController) CreateDomain(ctx *gin.Context) {
	if ctx.Request.Method == "POST" {
		var req CreateDomainRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusOK, gin.H{"error": err.Error()})
			return
		}

		err := account.CreateDomain(req.Name, req.Description)
		if err != nil {
			ctx.JSON(http.StatusOK, gin.H{"error": "create domain failed:" + err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, gin.H{"success": "create domain success"})
	} else {
		ctx.JSON(http.StatusOK, gin.H{"error": "method not allowed"})
	}
}

type AccountIndexRequest struct {
	DataTableRequest
	DomainID int    `json:"domainID"`
	Keyword  string `json:"keyword"`
	Status   int    `json:"status"`
}

type AccountIndexResponse struct {
	ID           int64     `json:"id"`
	Username     string    `json:"username"`
	Status       int       `json:"status"`
	CreateTime   time.Time `json:"createTime"`
	StorageQuota int64     `json:"storageQuota"`
	StorageUsage int64     `json:"storageUsage"`
	MailQuantity int64     `json:"mailQuantity"`
	MailUsage    int64     `json:"mailUsage"`
	ExpiredTime  time.Time `json:"expiredTime"`
}

type domainPair struct {
	ID   int64  `json:"ID"`
	Name string `json:"name"`
}

func (a *AccountController) IndexAccount(ctx *gin.Context) {
	if ctx.Request.Method == "GET" {
		did, err := strconv.Atoi(ctx.Query("did"))
		if err != nil {
			ctx.HTML(http.StatusBadRequest, "single_error.html", gin.H{"error": "invalid did"})
			return
		}
		domain, err := account.FindDomainByID(int64(did))
		if err != nil {
			ctx.HTML(http.StatusBadRequest, "single_error.html", gin.H{"error": "domain not found"})
			return
		}
		sess := sessions.Default(ctx)
		username := sess.Get("userName")
		domainPairs := make([]domainPair, 0)
		if domains, err := account.FindAllValidateDomain(); err == nil {
			for _, d := range domains {
				domainPairs = append(domainPairs, domainPair{ID: d.ID, Name: d.Name})
			}
		}

		ctx.HTML(http.StatusOK, "account_index.html", gin.H{
			"title":      "Account Management - Easymail",
			"domainID":   domain.ID,
			"domainName": domain.Name,
			"domains":    domainPairs,
			"username":   username,
		})
		return
	} else if ctx.Request.Method == "POST" {
		var req AccountIndexRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusOK, gin.H{"error": err.Error()})
			return
		}

		orderField := ""
		orderDir := ""
		for _, o := range req.Orders {
			orderField = req.Columns[o.Column].Data
			orderDir = o.Dir
			break
		}

		total, accounts, err := account.Index(req.DomainID, req.Status, req.Keyword, orderField, orderDir, req.Start, req.Length)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "search account error"})
			return
		}

		data := make([]AccountIndexResponse, 0)

		for _, acc := range accounts {
			status := 0
			if !acc.Active {
				if acc.Deleted {
					status = 2
				} else {
					status = 1
				}
			}
			if acc.Deleted {
				status = 2
			}

			// mail usage
			mailQuantity, err := account.GetMailQuantity(acc.ID)
			if err != nil {
				log.Println(err)
			}
			mailUsage, err := account.GetMailUsage(acc.ID)
			if err != nil {
				log.Println(err)
			}

			data = append(data, AccountIndexResponse{
				ID:           acc.ID,
				Username:     acc.Username,
				Status:       status,
				CreateTime:   acc.CreateTime,
				ExpiredTime:  acc.PasswordExpireTime,
				StorageQuota: acc.StorageQuota,
				MailQuantity: mailQuantity,
				MailUsage:    mailUsage,
			})
		}

		ctx.JSON(http.StatusOK, gin.H{
			"draw":            req.Draw,
			"recordsTotal":    total,
			"recordsFiltered": total,
			"data":            data,
		})
	}
}

type CreateAccountRequest struct {
	Name            string `json:"accountName" binding:"required,min=3,max=64"`
	DomainID        string `json:"domainID" binding:"required"`
	Password        string `json:"password" binding:"required,min=3,max=64"`
	PasswordAgain   string `json:"passwordRepeat" binding:"required,min=3,max=64"`
	StorageQuota    string `json:"storageQuota" binding:"required,min=-1,max=100000"`
	PasswordExpired string `json:"passwordExpired"`
}

func (a *AccountController) CreateAccount(ctx *gin.Context) {
	if ctx.Request.Method == "POST" {
		var req CreateAccountRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusOK, gin.H{"error": err.Error()})
			return
		}

		did, err := strconv.ParseInt(req.DomainID, 10, 64)
		if err != nil {
			ctx.JSON(http.StatusOK, gin.H{"error": "invalid domain id"})
			return
		}

		if req.PasswordExpired == "" {
			req.PasswordExpired = "2099-12-31"
		}
		passwordExpiredTime, err := time.Parse("2006-01-02", req.PasswordExpired)
		if err != nil {
			ctx.JSON(http.StatusOK, gin.H{"error": "invalid password expired time"})
			return
		}

		if req.Password != req.PasswordAgain {
			ctx.JSON(http.StatusOK, gin.H{"error": "password and password repeat not match"})
			return
		}
		var storageQuota int64
		if req.StorageQuota == "" {
			storageQuota = -1
		} else {
			storageQuota, err = strconv.ParseInt(req.StorageQuota, 10, 64)
			if err != nil {
				log.Println(err)
			}
		}
		err = account.CreateAccount(did, req.Name, req.Password, storageQuota, passwordExpiredTime)
		if err != nil {
			ctx.JSON(http.StatusOK, gin.H{"error": "create account failed:" + err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, gin.H{"success": "create account success"})
	} else {
		ctx.JSON(http.StatusOK, gin.H{"error": "method not allowed"})
	}
}

func (a *AccountController) ToggleAccountActive(ctx *gin.Context) {
	id := ctx.Query("id")
	if id == "" {
		ctx.JSON(http.StatusOK, gin.H{"error": "invalid account id"})
		return
	}

	accountID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{"error": "invalid account id"})
		return
	}

	err = account.ToggleAccountActive(accountID)

	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{"error": "toggle account active error"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"success": "toggle account active success"})
}

func (a *AccountController) DeleteAccount(ctx *gin.Context) {
	id := ctx.Query("id")
	if id == "" {
		ctx.JSON(http.StatusOK, gin.H{"error": "invalid account id"})
		return
	}

	accountID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{"error": "invalid account id"})
		return
	}

	err = account.DeleteAccount(accountID)

	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{"error": "delete account error"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"success": "delete account success"})
}

type EditAccountRequest struct {
	ID              string `json:"accountID" binding:"required"`
	Password        string `json:"editPassword"`
	StorageQuota    string `json:"editStorageQuota"`
	PasswordExpired string `json:"editPasswordExpired"`
}

func (a *AccountController) EditAccount(ctx *gin.Context) {
	if ctx.Request.Method == "POST" {
		var err error
		var req EditAccountRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusOK, gin.H{"error": err.Error()})
			return
		}

		accountID, err := strconv.ParseInt(req.ID, 10, 64)
		if err != nil {
			ctx.JSON(http.StatusOK, gin.H{"error": "invalid account id"})
			return
		}

		passwordExpiredTime := time.Time{}
		if req.PasswordExpired != "" {
			passwordExpiredTime, err = time.Parse("2006-01-02", req.PasswordExpired)
			if err != nil {
				ctx.JSON(http.StatusOK, gin.H{"error": "invalid password expired time"})
				return
			}
		}

		var storageQuota int64
		if req.StorageQuota == "" {
			storageQuota = -1
		} else {
			storageQuota, err = strconv.ParseInt(req.StorageQuota, 10, 64)
			if err != nil {
				log.Println(err)
			}
		}
		err = account.EditAccount(accountID, req.Password, storageQuota, passwordExpiredTime)
		if err != nil {
			ctx.JSON(http.StatusOK, gin.H{"error": "edit account failed:" + err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, gin.H{"success": "edit account success"})
	} else {
		ctx.JSON(http.StatusOK, gin.H{"error": "method not allowed"})
	}
}
