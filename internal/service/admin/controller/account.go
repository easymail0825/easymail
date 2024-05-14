package controller

import (
	"context"
	"easymail/internal/database"
	"easymail/internal/model"
	"easymail/internal/postfix/sync"
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

func (a *AccountController) IndexDomain(c *gin.Context) {
	if c.Request.Method == "GET" {
		sess := sessions.Default(c)
		username := sess.Get("userName")

		c.HTML(http.StatusOK, "domain_index.html", gin.H{
			"title":    "Domain Management - Easymail",
			"username": username,
			"module":   "model",
		})
		return
	} else if c.Request.Method == "POST" {
		var req DomainIndexRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusOK, gin.H{"error": err.Error()})
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
		total, domains, err := model.DomainIndex(req.Keyword, orderField, orderDir, req.Start, req.Length)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "search domain error"})
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
			totalAccount, err := model.CountDomainAccount(domain.ID)
			if err != nil {
				log.Println("count domain model error:", err)
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

		c.JSON(http.StatusOK, gin.H{
			"draw":            req.Draw,
			"recordsTotal":    total,
			"recordsFiltered": total,
			"data":            data,
		})
	}
}

func (a *AccountController) ToggleDomainActive(c *gin.Context) {
	id := c.Query("did")
	if id == "" {
		c.JSON(http.StatusOK, gin.H{"error": "invalid domain id"})
		return
	}

	domainID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"error": "invalid domain id"})
		return
	}

	err = model.ToggleDomainActive(domainID)

	if err != nil {
		c.JSON(http.StatusOK, gin.H{"error": "toggle domain active error"})
		return
	}
	err = sync.SynchronizeVirtualDomain()
	if err != nil {
		log.Println("synchronize virtual domains failed:", err.Error())
	}
	c.JSON(http.StatusOK, gin.H{"success": "toggle domain active success"})
}

func (a *AccountController) DeleteDomain(c *gin.Context) {
	id := c.Query("did")
	if id == "" {
		c.JSON(http.StatusOK, gin.H{"error": "invalid domain id"})
		return
	}

	domainID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"error": "invalid domain id"})
		return
	}

	err = model.DeleteDomain(domainID)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"error": "delete domain error"})
		return
	}
	err = sync.SynchronizeVirtualDomain()
	if err != nil {
		log.Println("synchronize virtual domains failed:", err.Error())
	}

	c.JSON(http.StatusOK, gin.H{"success": "delete domain success"})
}

type CreateDomainRequest struct {
	//ID          int    `json:"id"`
	Name        string `json:"domainName" binding:"required,min=3,max=255"`
	Description string `json:"description" binding:"required,min=3,max=255"`
}

func (a *AccountController) CreateDomain(c *gin.Context) {
	if c.Request.Method == "POST" {
		var req CreateDomainRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusOK, gin.H{"error": err.Error()})
			return
		}

		err := model.CreateDomain(req.Name, req.Description)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"error": "create domain failed:" + err.Error()})
			return
		}
		err = sync.SynchronizeVirtualDomain()
		if err != nil {
			log.Println("synchronize virtual domains failed:", err.Error())
		}
		c.JSON(http.StatusOK, gin.H{"success": "create domain success"})
	} else {
		c.JSON(http.StatusOK, gin.H{"error": "method not allowed"})
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

func (a *AccountController) IndexAccount(c *gin.Context) {
	if c.Request.Method == "GET" {
		did, err := strconv.Atoi(c.Query("did"))
		if err != nil {
			c.HTML(http.StatusBadRequest, "single_error.html", gin.H{"error": "invalid did"})
			return
		}
		domain, err := model.FindDomainByID(int64(did))
		if err != nil {
			c.HTML(http.StatusBadRequest, "single_error.html", gin.H{"error": "domain not found"})
			return
		}
		sess := sessions.Default(c)
		username := sess.Get("userName")
		domainPairs := make([]domainPair, 0)
		if domains, err := model.FindAllValidateDomain(); err == nil {
			for _, d := range domains {
				domainPairs = append(domainPairs, domainPair{ID: d.ID, Name: d.Name})
			}
		}

		c.HTML(http.StatusOK, "account_index.html", gin.H{
			"title":      "Account Management - Easymail",
			"domainID":   domain.ID,
			"domainName": domain.Name,
			"domains":    domainPairs,
			"username":   username,
			"module":     "model",
		})
		return
	} else if c.Request.Method == "POST" {
		var req AccountIndexRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusOK, gin.H{"error": err.Error()})
			return
		}

		orderField := ""
		orderDir := ""
		for _, o := range req.Orders {
			orderField = req.Columns[o.Column].Data
			orderDir = o.Dir
			break
		}

		total, accounts, err := model.Index(req.DomainID, req.Status, req.Keyword, orderField, orderDir, req.Start, req.Length)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "search model error"})
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
			mailQuantity, err := model.GetMailQuantity(acc.ID)
			if err != nil {
				log.Println(err)
			}
			mailUsage, err := model.GetMailUsage(acc.ID)
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

		c.JSON(http.StatusOK, gin.H{
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

func (a *AccountController) CreateAccount(c *gin.Context) {
	if c.Request.Method == "POST" {
		var req CreateAccountRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusOK, gin.H{"error": err.Error()})
			return
		}

		did, err := strconv.ParseInt(req.DomainID, 10, 64)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"error": "invalid domain id"})
			return
		}

		if req.PasswordExpired == "" {
			req.PasswordExpired = "2099-12-31"
		}
		passwordExpiredTime, err := time.Parse("2006-01-02", req.PasswordExpired)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"error": "invalid password expired time"})
			return
		}

		if req.Password != req.PasswordAgain {
			c.JSON(http.StatusOK, gin.H{"error": "password and password repeat not match"})
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
		err = model.CreateAccount(did, req.Name, req.Password, storageQuota, passwordExpiredTime)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"error": "create model failed:" + err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"success": "create model success"})
	} else {
		c.JSON(http.StatusOK, gin.H{"error": "method not allowed"})
	}
}

func (a *AccountController) ToggleAccountActive(c *gin.Context) {
	id := c.Query("id")
	if id == "" {
		c.JSON(http.StatusOK, gin.H{"error": "invalid model id"})
		return
	}

	accID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"error": "invalid model id"})
		return
	}

	err = model.ToggleAccountActive(accID)

	if err != nil {
		c.JSON(http.StatusOK, gin.H{"error": "toggle model active error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": "toggle model active success"})
}

func (a *AccountController) DeleteAccount(c *gin.Context) {
	id := c.Query("id")
	if id == "" {
		c.JSON(http.StatusOK, gin.H{"error": "invalid model id"})
		return
	}

	accID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"error": "invalid model id"})
		return
	}

	err = model.DeleteAccount(accID)

	if err != nil {
		c.JSON(http.StatusOK, gin.H{"error": "delete model error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": "delete model success"})
}

type EditAccountRequest struct {
	ID              string `json:"accID" binding:"required"`
	Password        string `json:"editPassword"`
	StorageQuota    string `json:"editStorageQuota"`
	PasswordExpired string `json:"editPasswordExpired"`
}

func (a *AccountController) EditAccount(c *gin.Context) {
	if c.Request.Method == "POST" {
		var err error
		var req EditAccountRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusOK, gin.H{"error": err.Error()})
			return
		}

		accID, err := strconv.ParseInt(req.ID, 10, 64)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"error": "invalid model id"})
			return
		}

		passwordExpiredTime := time.Time{}
		if req.PasswordExpired != "" {
			passwordExpiredTime, err = time.Parse("2006-01-02", req.PasswordExpired)
			if err != nil {
				c.JSON(http.StatusOK, gin.H{"error": "invalid password expired time"})
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
		err = model.EditAccount(accID, req.Password, storageQuota, passwordExpiredTime)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"error": "edit model failed:" + err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"success": "edit model success"})
	} else {
		c.JSON(http.StatusOK, gin.H{"error": "method not allowed"})
	}
}
