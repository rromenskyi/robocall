package main

import (
//	"crypto/rand"
	"math/rand"
	"time"
//"fmt"
)


func init() {
    rand.Seed(time.Now().UnixNano())
}



func generateRandomID() string {
    const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
    result := make([]byte, 32)
    for i := range result {
	result[i] = charset[rand.Intn(len(charset))]
    }
    return string(result)
}

