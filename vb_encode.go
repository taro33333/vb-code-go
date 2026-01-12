package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// vbEncode は整数を Variable Byte Code でエンコードする
func vbEncode(n uint64) []byte {
	if n == 0 {
		return []byte{0x80}
	}

	var bytes []byte
	for n > 0 {
		bytes = append([]byte{byte(n & 0x7F)}, bytes...)
		n >>= 7
	}
	// 最後のバイトに終端マーカー(0x80)を付ける
	bytes[len(bytes)-1] |= 0x80
	return bytes
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: %s <data file>\n", os.Args[0])
		os.Exit(1)
	}

	file, err := os.Open(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	// 長い行に対応するためバッファサイズを増やす (4MB)
	scanner.Buffer(make([]byte, 4*1024*1024), 4*1024*1024)
	writer := bufio.NewWriter(os.Stdout)
	defer writer.Flush()

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, "\t", 2)
		if len(parts) != 2 {
			continue
		}

		tag := parts[0]
		nums := parts[1]

		// 整数列の差分を取って VB Code で符号化
		var vb []byte
		var pre uint64 = 0
		for _, s := range strings.Split(nums, ",") {
			n, err := strconv.ParseUint(strings.TrimSpace(s), 10, 64)
			if err != nil {
				continue
			}
			vb = append(vb, vbEncode(n-pre)...)
			pre = n
		}

		// タグ、VB符号の長さを付与しつつ出力 (ビッグエンディアン32ビット)
		tagBytes := []byte(tag)
		binary.Write(writer, binary.BigEndian, uint32(len(tagBytes)))
		binary.Write(writer, binary.BigEndian, uint32(len(vb)))
		writer.Write(tagBytes)
		writer.Write(vb)
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading file: %v\n", err)
		os.Exit(1)
	}
}
