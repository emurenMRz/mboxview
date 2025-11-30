package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <mbox-file>\n", os.Args[0])
		os.Exit(75) // EX_TEMPFAIL
	}
	mboxPath := os.Args[1]

	// ファイルを排他モードで開く（簡易版）
	f, err := os.OpenFile(mboxPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0660)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot open mbox: %v\n", err)
		os.Exit(75)
	}
	defer f.Close()

	scanner := bufio.NewScanner(os.Stdin)
	inHeader := true
	for scanner.Scan() {
		line := scanner.Text()

		// ヘッダ部は空行までそのまま書く
		if inHeader {
			_, err := f.WriteString(line + "\n")
			if err != nil {
				fmt.Fprintf(os.Stderr, "write error: %v\n", err)
				f.Close()
				os.Exit(75)
			}
			if line == "" {
				inHeader = false
			}
			continue
		}

		// 本文中の "From " 行だけエスケープ
		if strings.HasPrefix(line, "From ") {
			line = ">" + line
		}

		_, err := f.WriteString(line + "\n")
		if err != nil {
			fmt.Fprintf(os.Stderr, "write error: %v\n", err)
			f.Close()
			os.Exit(75)
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "read error: %v\n", err)
		f.Close()
		os.Exit(75)
	}

	// メッセージ末尾に空行を追加
	f.WriteString("\n")

	f.Close()
	os.Exit(0) // EX_OK
}
