package model

import (
	"easymail/internal/database"
	"time"
)

type SsdeepHash struct {
	ID         int64     `gorm:"primaryKey;AUTO_INCREMENT" json:"id"`
	SessionID  string    `gorm:"type:varchar(128);index:idx_session,unique" json:"session_id"`
	ChunkSize  int       `gorm:"type:int(11);default(0);index:idx_chunk_size" json:"chunk_size"`
	Hash       string    `gorm:"type:varchar(255);index:idx_hash,unique" json:"hash"`
	IsAttach   bool      `gorm:"type:tinyint(1);default(0)" json:"is_attach"`
	CreateTime time.Time `json:"create_time"`
}

func CreateSsdeepHash(hash string, sessionID string, chunkSize int, isAttach bool) (*SsdeepHash, error) {
	db := database.GetDB()

	// check if hash exists
	var h SsdeepHash
	err := db.Where("hash = ?", hash).First(&h).Error
	if err == nil {
		return &h, nil
	}
	h = SsdeepHash{
		Hash:       hash,
		SessionID:  sessionID,
		ChunkSize:  chunkSize,
		IsAttach:   isAttach,
		CreateTime: time.Now(),
	}

	err = db.Create(&h).Error

	return &h, err
}
