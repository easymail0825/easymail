package filter

import (
	"crypto/md5"
	"easymail/vender/ssdeep"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"strconv"
	"strings"
)

// computeAttachHash only compute hash for the first 10240 bytes + last 10240 bytes
func computeAttachHash(data []byte) (hash string, err error) {
	if len(data) < 20480 {
		return ssdeep.FuzzyBytes(data)
	}
	dist := make([]byte, 20480)
	copy(dist, data[:10240])
	copy(dist[10240:], data[len(data)-10240:])
	return ssdeep.FuzzyBytes(dist)
}

// computeAttachMD5 only compute hash for the first 10240 bytes + last 10240 bytes
func computeAttachMD5(data []byte) (hash string, err error) {
	if len(data) < 20480 {
		return ssdeep.FuzzyBytes(data)
	}
	dist := make([]byte, 20480)
	copy(dist, data[:10240])
	copy(dist[10240:], data[len(data)-10240:])
	s := md5.Sum(dist)
	return hex.EncodeToString(s[:]), nil
}

// getAll7CharChunks
func getAll7CharChunks(h string) []uint64 {
	chunks := make([]uint64, 0)
	for i := 0; i <= len(h)-7; i++ {
		chunk := h[i:i+7] + "="
		decoded, err := base64.StdEncoding.DecodeString(chunk)
		if err != nil {
			continue
		}
		decoded = append(decoded, byte(0), byte(0), byte(0))
		if len(decoded) != 8 {
			continue
		}
		num := binary.LittleEndian.Uint64(decoded)
		chunks = append(chunks, num)
	}
	return chunks
}

// preprocessHash cut ssdeep hash into continue int64 list
func preprocessHash(h string) (int, []uint64, []uint64) {
	parts := strings.SplitN(h, ":", 2)
	if len(parts) != 2 {
		return 0, nil, nil
	}
	blockSize, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, nil, nil
	}

	for _, c := range parts[1] {
		str := strings.Repeat(string(c), 4)
		for strings.Contains(parts[1], str) {
			parts[1] = strings.Replace(parts[1], str, strings.Repeat(string(c), 3), 1)
		}
	}

	parts = strings.SplitN(parts[1], ":", 2)
	if len(parts) != 2 {
		return 0, nil, nil
	}

	blockDataChunks := getAll7CharChunks(parts[0])
	doubleBlockDataChunks := getAll7CharChunks(parts[1])
	return blockSize, blockDataChunks, doubleBlockDataChunks
}
