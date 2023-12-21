package common

import (
	"fmt"
	"hash/fnv"
	"math/rand"
	"time"
)

// 通过请求的方法信息，生成请求id
//
// @param	mothed 	方法信息
// @return 请求id
func GenRid(mothed string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(mothed))

	return h.Sum64()
}

func GenUid() string {
	seed := time.Now().UnixNano()
	src := rand.NewSource(seed)
	rng := rand.New(src)

	randomNumber := rng.Intn(1000)

	timestamp := time.Now().UnixNano() / int64(time.Millisecond)

	return fmt.Sprintf("%d%d", timestamp, randomNumber)
}
