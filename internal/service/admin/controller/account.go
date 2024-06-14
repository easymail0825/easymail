package controller

import (
	"easymail/internal/model"
	"easymail/internal/postfix/sync"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type AccountController struct{}

func (a *AccountController) IndexDomain(c *gin.Context) {
	if c.Request.Method == "GET" {
		sess := sessions.Default(c)
		username := sess.Get("userName")
		menu := createMenu()
		c.HTML(http.StatusOK, "domain_index.html", gin.H{
			"title":    "Domain Management - Easymail",
			"username": username,
			"module":   "account",
			"page":     "domain",
			"menu":     menu,
		})
		return
	} else if c.Request.Method == "POST" {
		var req model.IndexDomainRequest
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
		total, domains, err := model.IndexDomain(req.Keyword, orderField, orderDir, req.Start, req.Length)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "search domain error"})
			return
		}

		data := make([]model.IndexDomainResponse, 0)

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

			// mx
			mxes, err := resolver.LookupMX(domain.Name)
			mx := strings.Join(mxes, "<br>\n")

			// spf
			spf, _ := resolver.LookupSPF(domain.Name)

			// dmarc
			dmarcs, _ := resolver.LookupDMARC(domain.Name)
			dmarc := strings.Join(dmarcs, "<br>\n")

			data = append(data, model.IndexDomainResponse{
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

func (a *AccountController) CreateDomain(c *gin.Context) {
	if c.Request.Method == "POST" {
		var req model.CreateDomainRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusOK, gin.H{"error": err.Error()})
			return
		}

		err := model.CreateDomain(req)
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
		menu := createMenu()
		c.HTML(http.StatusOK, "account_index.html", gin.H{
			"title":      "Account Management - Easymail",
			"domainID":   domain.ID,
			"domainName": domain.Name,
			"domains":    domainPairs,
			"username":   username,
			"module":     "account",
			"page":       "account",
			"menu":       menu,
		})
		return
	} else if c.Request.Method == "POST" {
		var req model.IndexAccountRequest
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
			c.JSON(http.StatusInternalServerError, gin.H{"error": "search account error"})
			return
		}

		data := make([]model.IndexAccountResponse, 0)

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

			data = append(data, model.IndexAccountResponse{
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

func (a *AccountController) CreateAccount(c *gin.Context) {
	if c.Request.Method == "POST" {
		var err error
		var req model.CreateAccountRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusOK, gin.H{"error": err.Error()})
			return
		}

		if req.PasswordExpired == "" {
			req.PasswordExpired = "2099-12-31"
		}
		req.PasswordExpiredTime, err = time.Parse("2006-01-02", req.PasswordExpired)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"error": "invalid password expired time"})
			return
		}

		if req.Password != req.PasswordAgain {
			c.JSON(http.StatusOK, gin.H{"error": "password and password repeat not match"})
			return
		}

		err = model.CreateAccount(req)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"error": "create account failed:" + err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"success": "create account success"})
	} else {
		c.JSON(http.StatusOK, gin.H{"error": "method not allowed"})
	}
}

func (a *AccountController) ToggleAccount(c *gin.Context) {
	id := c.Query("id")
	if id == "" {
		c.JSON(http.StatusOK, gin.H{"error": "invalid account id"})
		return
	}

	accID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"error": "invalid account id"})
		return
	}

	err = model.ToggleAccount(accID)

	if err != nil {
		c.JSON(http.StatusOK, gin.H{"error": "toggle account active error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": "toggle account active success"})
}

func (a *AccountController) DeleteAccount(c *gin.Context) {
	id := c.Query("id")
	if id == "" {
		c.JSON(http.StatusOK, gin.H{"error": "invalid account id"})
		return
	}

	accID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"error": "invalid account id"})
		return
	}

	err = model.DeleteAccount(accID)

	if err != nil {
		c.JSON(http.StatusOK, gin.H{"error": "delete account error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": "delete account success"})
}

func (a *AccountController) EditAccount(c *gin.Context) {
	if c.Request.Method == "POST" {
		var err error
		var req model.EditAccountRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusOK, gin.H{"error": err.Error()})
			return
		}

		if req.PasswordExpired != "" {
			req.PasswordExpiredTime, err = time.Parse("2006-01-02", req.PasswordExpired)
			if err != nil {
				c.JSON(http.StatusOK, gin.H{"error": "invalid password expired time"})
				return
			}
		}

		if req.StorageQuota == "" {
			req.StorageQuotaNumber = -1
		} else {
			req.StorageQuotaNumber, err = strconv.ParseInt(req.StorageQuota, 10, 64)
			if err != nil {
				log.Println(err)
			}
		}
		err = model.EditAccount(req)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{"error": "edit account failed:" + err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"success": "edit account success"})
	} else {
		c.JSON(http.StatusOK, gin.H{"error": "method not allowed"})
	}
}
