package main

import (
	"bufio"
	"encoding/gob"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// GapArray は差分エンコーディングで整数配列を圧縮
type GapArray struct {
	Gaps []int
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

// DocInfo はドキュメントの情報を保持
type DocInfo struct {
	Title string
	URL   string
}

func main() {
	if len(os.Args) != 3 {
		fmt.Fprintf(os.Stderr, "usage: %s titles_file text_dir\n", os.Args[0])
		os.Exit(1)
	}

	titleFile := os.Args[1]
	textDir := os.Args[2]

	// 基本データファイルからタイトル等を読み取る
	titles := make(map[string]*DocInfo)
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
			url := parts[2]
			title := parts[3]
			titles[docID] = &DocInfo{Title: title, URL: url}
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading title file: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "docs: %d\n", len(titles))
	fmt.Fprintf(os.Stderr, "now index loading\n")

	// 転置インデックスをメモリにロード
	indexFile, err := os.Open("index.data")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening index file: %v\n", err)
		os.Exit(1)
	}
	defer indexFile.Close()

	var index map[string]*GapArray
	decoder := gob.NewDecoder(indexFile)
	if err := decoder.Decode(&index); err != nil {
		fmt.Fprintf(os.Stderr, "Error decoding index: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "terms: %d\n", len(index))

	// 対話プロンプトを起動
	reader := bufio.NewReader(os.Stdin)
	limit := 5

	for {
		fmt.Print("> ")
		input, err := reader.ReadString('\n')
		if err != nil {
			break
		}

		input = strings.TrimSpace(input)
		if input == "" {
			continue
		}

		// 終了コマンド
		if input == "exit" || input == "quit" {
			break
		}

		// 転置インデックスから検索
		if gapArray, ok := index[input]; ok {
			// PostingsList を取得
			arr := gapArray.ToArray()
			resSize := len(arr)

			message := ""
			if resSize > limit {
				message = fmt.Sprintf("Results 1 - %d  of about %d for %s\n", limit, resSize, input)
				arr = arr[0:limit]
			}

			for _, docIDInt := range arr {
				docID := fmt.Sprintf("%d", docIDInt)

				if info, ok := titles[docID]; ok {
					snippet := getSnippet(input, docID, textDir)
					fmt.Printf("[%d] %s\n%s\n\n\"%s\"\n\n\n",
						docIDInt,
						info.Title,
						info.URL,
						snippet)
				}
			}

			if message != "" {
				fmt.Print(message)
			}
		} else {
			fmt.Println("No Match")
		}
	}
}

// getSnippet はスニペットを表示
func getSnippet(word, docID, textDir string) string {
	filePath := filepath.Join(textDir, docID)
	content, err := os.ReadFile(filePath)
	if err != nil {
		return ""
	}

	doc := string(content)
	pos := strings.Index(doc, word)
	if pos == -1 {
		// 見つからない場合は先頭から200文字
		if len(doc) > 200 {
			doc = doc[0:200]
		}
		doc = strings.ReplaceAll(doc, "\n", "")
		return doc
	}

	wlen := len(word)
	start := pos - 100
	if start < 0 {
		start = 0
	}

	end := start + wlen + 200
	if end > len(doc) {
		end = len(doc)
	}

	res := doc[start:end]
	res = strings.ReplaceAll(res, "\n", "")

	return res
}
