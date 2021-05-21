package main

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
)

func prepare(s []byte) [][]byte {
	// append a bytes started with bit '1'
	tmp := append(s, 0x80)
	l := len(s)
	// zero-padding
	if l % 64 < 56 {
		tmp = append(tmp, make([]byte, 55 - l % 64)...)
	} else {
		tmp = append(tmp, make([]byte, 121 - l % 64)...)
	}
	// raw data length
	tmp2 := make([]byte, 8)
	binary.BigEndian.PutUint64(tmp2, uint64(l<<3))
	tmp = append(tmp, tmp2...)
	// chunk
	var prepared [][]byte
	for i := 0; i < len(tmp) / 64; i++ {
		prepared = append(prepared, tmp[i * 64 : i * 64 + 64])
	}
	return prepared
}

func sha256(s []byte) string {

	chunks := prepare([]byte(s))

	H := [...]uint32{0x6a09e667, 0xbb67ae85, 0x3c6ef372, 0xa54ff53a, 0x510e527f, 0x9b05688c, 0x1f83d9ab, 0x5be0cd19}
	k := [...]uint32{0x428a2f98, 0x71374491, 0xb5c0fbcf, 0xe9b5dba5, 0x3956c25b, 0x59f111f1, 0x923f82a4, 0xab1c5ed5, 0xd807aa98, 0x12835b01, 0x243185be, 0x550c7dc3, 0x72be5d74, 0x80deb1fe, 0x9bdc06a7, 0xc19bf174 ,0xe49b69c1, 0xefbe4786, 0x0fc19dc6, 0x240ca1cc, 0x2de92c6f, 0x4a7484aa, 0x5cb0a9dc, 0x76f988da ,0x983e5152, 0xa831c66d, 0xb00327c8, 0xbf597fc7, 0xc6e00bf3, 0xd5a79147, 0x06ca6351, 0x14292967 ,0x27b70a85, 0x2e1b2138, 0x4d2c6dfc, 0x53380d13, 0x650a7354, 0x766a0abb, 0x81c2c92e, 0x92722c85 ,0xa2bfe8a1, 0xa81a664b, 0xc24b8b70, 0xc76c51a3, 0xd192e819, 0xd6990624, 0xf40e3585, 0x106aa070 ,0x19a4c116, 0x1e376c08, 0x2748774c, 0x34b0bcb5, 0x391c0cb3, 0x4ed8aa4a, 0x5b9cca4f, 0x682e6ff3 ,0x748f82ee, 0x78a5636f, 0x84c87814, 0x8cc70208, 0x90befffa, 0xa4506ceb, 0xbef9a3f7, 0xc67178f2}

	for _, chunk := range chunks {
		w := make([]uint32, 64)
		for i := 0; i < 16; i++ {
			w[i] = binary.BigEndian.Uint32(chunk[i * 4: i * 4 + 4])
		}
		for i := 16; i < 64; i++ {
			s0 := (rightRotate(w[i-15], 7)) ^ (rightRotate(w[i-15], 18)) ^ (w[i-15] >> 3)
			s1 := (rightRotate(w[i-2], 17)) ^ (rightRotate(w[i-2], 19)) ^ (w[i-2] >> 10)
			w[i] = w[i-16] + s0 + w[i-7] + s1
		}

		a := H[0]
		b := H[1]
		c := H[2]
		d := H[3]
		e := H[4]
		f := H[5]
		g := H[6]
		h := H[7]

		for i := 0; i < 64; i++ {
			s1 := (rightRotate(e, 6)) ^ (rightRotate(e, 11)) ^ (rightRotate(e, 25))
			ch := (e & f) ^ (^e & g)
			temp1 := h + s1 + ch + k[i] + w[i]
			s0 := (rightRotate(a, 2)) ^ (rightRotate(a, 13)) ^ (rightRotate(a, 22))
			maj := (a & b) ^ (a & c) ^ (b & c)
			temp2 := s0 + maj
			h = g
			g = f
			f = e
			e = d + temp1
			d = c
			c = b
			b = a
			a = temp1 + temp2
		}

		H[0] = H[0] + a
		H[1] = H[1] + b
		H[2] = H[2] + c
		H[3] = H[3] + d
		H[4] = H[4] + e
		H[5] = H[5] + f
		H[6] = H[6] + g
		H[7] = H[7] + h

	}

	res := make([]byte, 32)
	for i := 0; i < 8; i++ {
		binary.BigEndian.PutUint32(res[4 * i:], H[i])
	}
	return hex.EncodeToString(res)
}

func rightRotate(n uint32, d uint) uint32 {
	return (n >> d) | (n << (32 - d))
}

func main()  {
	s := "hello world"
	res := sha256([]byte(s))
	fmt.Println(res)
}
