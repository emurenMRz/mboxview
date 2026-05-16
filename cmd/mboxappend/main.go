//go:build !windows

package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"syscall"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <mbox-file>\n", os.Args[0])
		os.Exit(75) // EX_TEMPFAIL
	}
	mboxPath := os.Args[1]

	// stdin から全量をメモリに読み込み
	input, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read error: %v\n", err)
		os.Exit(75)
	}

	// ヘッダーはそのまま、本文の "From " 行をエスケープしてから結合
	var buf bytes.Buffer
	scanner := bufio.NewScanner(bytes.NewReader(input))
	inHeader := true
	for scanner.Scan() {
		line := scanner.Text()

		if inHeader {
			buf.WriteString(line)
			buf.WriteByte('\n')
			if line == "" {
				inHeader = false
			}
			continue
		}

		if strings.HasPrefix(line, "From ") {
			line = ">" + line
		}
		buf.WriteString(line)
		buf.WriteByte('\n')
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "read error: %v\n", err)
		os.Exit(75)
	}

	// 末尾に空行を追加
	buf.WriteByte('\n')

	// mbox ファイルを排他モードで開き、flock 取得後一括書き込み
	f, err := os.OpenFile(mboxPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0660)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot open mbox: %v\n", err)
		os.Exit(75)
	}
	defer f.Close()

	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
		fmt.Fprintf(os.Stderr, "flock error: %v\n", err)
		os.Exit(75)
	}
	defer syscall.Flock(int(f.Fd()), syscall.LOCK_UN)

	if _, err := f.Write(buf.Bytes()); err != nil {
		fmt.Fprintf(os.Stderr, "write error: %v\n", err)
		os.Exit(75)
	}
}
