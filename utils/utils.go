package utils

import (
	"crypto/rand"
	"encoding/binary"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// waitTimeout waits for the waitgroup for the specified max timeout.
// Returns true if waiting timed out.
func WaitTimeout(wg *sync.WaitGroup, timeout time.Duration) (bool, error) {
	c := make(chan struct{})
	go func() {
		defer close(c)
		wg.Wait()
	}()
	select {
	case <-c:
		return true, nil // completed normally
	case <-time.After(timeout):
		return false, errors.New("time out") // timed out
	}
}

func GenerateRandom_ByteArray(count int) []byte {
	bytes := make([]byte, count)
	rand.Read(bytes)

	return bytes
}

func GenerateRandomUint() uint32 {
	bs := GenerateRandom_ByteArray(4)

	result := binary.LittleEndian.Uint32(bs)

	return result
}
func GetCurrentDir() (string, error) {
	// 找到该启动当前进程的可执行文件的路径名
	str, err := os.Executable()
	if err != nil {
		return "", err
	}
	str = filepath.Dir(str)

	return str, nil
}
