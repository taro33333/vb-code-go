package main

import (
	"bufio"
	"encoding/gob"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ikawaha/kagome-dict/ipa"
	"github.com/ikawaha/kagome/v2/tokenizer"
)

// GapArray は差分エンコーディングで整数配列を圧縮
type GapArray struct {
	Gaps []int
}

// NewGapArray は整数配列から差分配列を作成
func NewGapArray(arr []int) *GapArray {
	if len(arr) == 0 {
		return &GapArray{Gaps: []int{}}
	}

	gaps := make([]int, len(arr))
	gaps[0] = arr[0]
	for i := 1; i < len(arr); i++ {
		gaps[i] = arr[i] - arr[i-1]
	}
	return &GapArray{Gaps: gaps}
}

// ToArray は差分配列から元の整数配列を復元
func (g *GapArray) ToArray() []int {
	if len(g.Gaps) == 0 {
		return []int{}
	}

	arr := make([]int, len(g.Gaps))
	arr[0] = g.Gaps[0]
	for i := 1; i < len(g.Gaps); i++ {
		arr[i] = arr[i-1] + g.Gaps[i]
	}
	return arr
}

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintf(os.Stderr, "usage: %s titles_file text_dir\n", os.Args[0])
		os.Exit(1)
	}

	titleFile := os.Args[1]
	textDir := os.Args[2]

	// 基本データファイルからタイトル、URLなどを読み取る
	titles := make(map[string]string)
	file, err := os.Open(titleFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening title file: %v\n", err)
		os.Exit(1)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, "\t")
		if len(parts) >= 4 {
			docID := parts[0]
			title := parts[3]
			titles[docID] = title
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading title file: %v\n", err)
		os.Exit(1)
	}

	// Kagome (MeCabの代替) で全文書を形態素解析、転置インデックスを作る
	t, err := tokenizer.New(ipa.Dict(), tokenizer.OmitBosEos())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating tokenizer: %v\n", err)
		os.Exit(1)
	}

	mainIndex := make(map[string][]int)

	entries, err := os.ReadDir(textDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading text directory: %v\n", err)
		os.Exit(1)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		docID := entry.Name()
		filePath := filepath.Join(textDir, docID)

		content, err := os.ReadFile(filePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error reading file %s: %v\n", filePath, err)
			continue
		}

		// タイトルと本文を結合
		doc := ""
		if title, ok := titles[docID]; ok {
			doc = title + string(content)
		} else {
			doc = string(content)
		}

		// 形態素解析
		tokens := t.Tokenize(doc)
		for _, token := range tokens {
			key := token.Surface

			if key == "" {
				continue
			}

			// 文書IDを数値に変換（ファイル名をそのまま整数として扱う）
			var docIDInt int
			fmt.Sscanf(docID, "%d", &docIDInt)

			mainIndex[key] = append(mainIndex[key], docIDInt)
		}
	}

	// 転置インデックスの PostingsList を整理（重複削除、ソート、圧縮）
	compressedIndex := make(map[string]*GapArray)
	for key, postings := range mainIndex {
		// 重複削除とソート
		uniquePostings := uniqueAndSort(postings)
		// 差分エンコーディングで圧縮
		compressedIndex[key] = NewGapArray(uniquePostings)
	}

	// インデックスをgobでディスクに書き出し
	outFile, err := os.Create("index.data")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating index file: %v\n", err)
		os.Exit(1)
	}
	defer outFile.Close()

	encoder := gob.NewEncoder(outFile)
	if err := encoder.Encode(compressedIndex); err != nil {
		fmt.Fprintf(os.Stderr, "Error encoding index: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Index created successfully with %d terms\n", len(compressedIndex))
}

// uniqueAndSort は整数配列の重複を削除してソート
func uniqueAndSort(arr []int) []int {
	if len(arr) == 0 {
		return arr
	}

	sort.Ints(arr)

	result := []int{arr[0]}
	for i := 1; i < len(arr); i++ {
		if arr[i] != arr[i-1] {
			result = append(result, arr[i])
		}
	}
	return result
}
