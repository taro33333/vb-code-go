package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

// vbDecode は Variable Byte Code でエンコードされたバイト列を整数列にデコードする
func vbDecode(data []byte) []uint64 {
	var nums []uint64
	var n uint64 = 0

	for _, b := range data {
		// 下位7ビットを取り出して値に追加
		n = (n << 7) | uint64(b&0x7F)

		// 最上位ビットが1なら値の終端
		if b&0x80 != 0 {
			nums = append(nums, n)
			n = 0
		}
	}

	return nums
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "usage: %s <binary file>\n", os.Args[0])
		os.Exit(1)
	}

	file, err := os.Open(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	writer := bufio.NewWriter(os.Stdout)
	defer writer.Flush()

	for {
		// タグ、VB符号部の長さを読み取る (8バイト = 32ビット + 32ビット)
		var tlen, vblen uint32
		if err := binary.Read(reader, binary.BigEndian, &tlen); err != nil {
			if err == io.EOF {
				break
			}
			fmt.Fprintf(os.Stderr, "Error reading header: %v\n", err)
			os.Exit(1)
		}
		if err := binary.Read(reader, binary.BigEndian, &vblen); err != nil {
			fmt.Fprintf(os.Stderr, "Error reading header: %v\n", err)
			os.Exit(1)
		}

		// 読み取った長さでタグ、VB符号部を読み取る
		tag := make([]byte, tlen)
		if _, err := io.ReadFull(reader, tag); err != nil {
			fmt.Fprintf(os.Stderr, "Error reading tag: %v\n", err)
			os.Exit(1)
		}

		vb := make([]byte, vblen)
		if _, err := io.ReadFull(reader, vb); err != nil {
			fmt.Fprintf(os.Stderr, "Error reading vb data: %v\n", err)
			os.Exit(1)
		}

		// VB Codeで復号し、差分だった値を元に戻す
		decoded := vbDecode(vb)
		nums := make([]string, len(decoded))
		var pre uint64 = 0
		for i, diff := range decoded {
			val := pre + diff
			nums[i] = strconv.FormatUint(val, 10)
			pre = val
		}

		// 当初のフォーマットに合わせて出力
		fmt.Fprintf(writer, "%s\t%s\n", string(tag), strings.Join(nums, ","))
	}
}
