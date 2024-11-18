package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
)

type LineConverter struct {
}

func (li LineConverter) ConvertLine(line string, parseState *ParseState) (string, error) {
	// TODO: 実装
	return line, nil
}

type ParseState struct {
	// TODO: Order を処理している途中かどうかを判定するフィールドの追加
}

type Order string
type Converter struct {
	lineConverter *LineConverter
	parseState    *ParseState
}

func (ip Converter) ReadOrder(line string) Order {
	// TODO: なんかsplit とかで実装
	return ""
}
func (ip Converter) ExecOrder(line string, order Order, parseState *ParseState) {
	// TODO: parseState をoverwrite
	// TODO: そういえばline も書き換えないといけねえなあ
}
func (ip Converter) Convert(fileName string) ([]string, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return nil, fmt.Errorf("can not open file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var resultLines []string

	for scanner.Scan() {
		line := scanner.Text()
		order := ip.ReadOrder(line)
		if order != "" {
			// order あります！
			ip.ExecOrder(line, order, ip.parseState)
			resultLines = append(resultLines, line)
			continue
		}
		convertedLine, err := ip.lineConverter.ConvertLine(line, ip.parseState)
		if err != nil {
			return nil, fmt.Errorf("変換失敗！: %w", err)
		}
		resultLines = append(resultLines, convertedLine)
	}

	return resultLines, nil
}

func writeLines(fileName string, lines []string) error {
	file, err := os.Create(fileName)
	if err != nil {
		return fmt.Errorf("cannot open file %s: %w", fileName, err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, line := range lines {
		_, err = fmt.Fprintln(writer, line)
		if err != nil {
			return fmt.Errorf("cannot write to file %s: %w", fileName, err)
		}
	}
	err = writer.Flush()
	if err != nil {
		return fmt.Errorf("cannot flush to file %s: %w", fileName, err)
	}

	return nil
}

func main() {
	var lineConverter LineConverter
	var parseState ParseState
	converter := Converter{&lineConverter, &parseState}

	fileName := os.Args[1]
	lines, err := converter.Convert(fileName)
	if err != nil {
		log.Fatal(err)
	}

	outFileName := fileName + ".converted"
	err = writeLines(outFileName, lines)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Fprintf(os.Stderr, "successfully converted and wrote to %s \n", outFileName)
}
