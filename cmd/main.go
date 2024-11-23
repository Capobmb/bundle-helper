package main

import (
	"bufio"
	"fmt"
	"log"
	"log/slog"
	"os"
	"strings"
)

type LineConverter struct {
}

type Line struct {
	content    string
	lineNumber int
}

func (li Line) String() string {
	return fmt.Sprintf("line %d: %s", li.lineNumber, li.content)
}

func (li Line) SetContent(content string) Line {
	li.content = content
	return li
}

func (li LineConverter) ConvertLine(line Line, parseState *ParseState) (Line, error) {
	const COMMENT_PREFIX = "//"
	makeCommentLine := func(line string) string {
		return COMMENT_PREFIX + line
	}
	switch parseState.Current {
	case undefined:
		return line, fmt.Errorf("undefined parse state detected: %v", parseState)
	case identity:
		return line, nil
	case isSingleLineCommenting:
		return line.SetContent(makeCommentLine(line.content)), nil
	case isSingleLineUncommenting:
		return line.SetContent(strings.TrimPrefix(line.content, COMMENT_PREFIX)), nil
	case isBlockCommenting:
		return line.SetContent(makeCommentLine(line.content)), nil
	case isBlockUncommenting:
		return line.SetContent(strings.TrimPrefix(line.content, COMMENT_PREFIX)), nil
	}
	return line, nil
}

type ParseState struct {
	Current AtomicParseState
	Next    AtomicParseState
}

type AtomicParseState int

const (
	undefined = iota
	identity
	isSingleLineCommenting
	isSingleLineUncommenting
	isBlockCommenting
	isBlockUncommenting
	// TODO: add isErasingLine, isProducingCommentingCommand, isProducingUncommentingCommand
)

type CommandType string

const (
	CommentSingleLine   CommandType = "comment_single_line"
	UncommentSingleLine CommandType = "uncomment_single_line"
	CommentBlockBegin   CommandType = "comment_block_begin"
	CommentBlockEnd     CommandType = "comment_block_end"
	UncommentBlockBegin CommandType = "uncomment_block_begin"
	UncommentBlockEnd   CommandType = "uncomment_block_end"
)

type CommandHandler func(state *ParseState) error

type CommandProcessor struct {
	handlers map[CommandType]CommandHandler
}

func NewCommandProcessor() *CommandProcessor {
	return &CommandProcessor{
		handlers: make(map[CommandType]CommandHandler),
	}
}

func (p *CommandProcessor) Register(cmdType CommandType, handler CommandHandler) {
	p.handlers[cmdType] = handler
}

type Converter struct {
	lineConverter    *LineConverter
	parseState       *ParseState
	commandProcessor *CommandProcessor
}

func (c Converter) ReadOrder(line Line, cmdProcessor *CommandProcessor) error {
	trimmedLine := strings.Trim(line.content, " ")
	const COMMENTLINEPREFIX = "//"
	if !strings.HasPrefix(trimmedLine, COMMENTLINEPREFIX) {
		// not an comment, returning
		slog.Debug("not a comment line", "line", line)
		return nil
	} else {
		slog.Debug("comment line", "line", line)
	}
	trimmedLine = strings.TrimPrefix(trimmedLine, COMMENTLINEPREFIX)
	trimmedLine = strings.Trim(trimmedLine, " ")

	slog.Debug("trimmedLine", "trimmedLine", trimmedLine)
	splittedLine := strings.Split(trimmedLine, " ")
	for i := range splittedLine {
		splittedLine[i] = strings.Trim(splittedLine[i], " ")
	}
	// debug print splittedLine
	slog.Debug("splittedLine", "splittedLine", splittedLine, "size", len(splittedLine))

	const COMMANDPREFIX = "@bundle-helper"
	if len(splittedLine) < 1 || splittedLine[0] != COMMANDPREFIX {
		// not a command, returning
		return nil
	}
	// this line is a command

	if len(splittedLine) < 2 {
		return fmt.Errorf("command argument not enough. line content: %s", line)
	}
	commandTypeStr := splittedLine[1]
	var commandType CommandType
	for cmdType := range cmdProcessor.handlers {
		if string(cmdType) == commandTypeStr {
			commandType = cmdType
		}
	}
	if commandType == "" {
		return fmt.Errorf("unknown command type `%s` in line `%s`", commandType, line)
	}
	slog.Debug("command type", "commandType", commandType)

	handler, found := cmdProcessor.handlers[commandType]
	if !found {
		return fmt.Errorf("unknown command type `%s` in line `%s`", commandType, line)
	}

	// handle
	err := handler(c.parseState)
	if err != nil {
		return err
	}

	return nil
}

func (lc LineConverter) NotifyConvertLineFinish(parseState *ParseState) error {
	parseState.Current = parseState.Next
	parseState.Next = identity
	if parseState.Current == isBlockCommenting {
		parseState.Next = isBlockCommenting
	}
	if parseState.Current == isBlockUncommenting {
		parseState.Next = isBlockUncommenting
	}
	return nil
}

func (c Converter) Convert(fileName string) ([]Line, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return nil, fmt.Errorf("can not open file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var resultLines []Line

	for i := 0; scanner.Scan(); i++ {
		lineContent := scanner.Text()
		currentLine := Line{lineContent, i}
		slog.Debug("line", "lineNumber", i, "current parse state", c.parseState)
		err := c.ReadOrder(currentLine, c.commandProcessor)
		if err != nil {
			return nil, fmt.Errorf("command read error: %w", err)
		}
		slog.Debug("line", "lineNumber", i, "after read order", "current parse state", c.parseState)
		convertedLine, err := c.lineConverter.ConvertLine(currentLine, c.parseState)
		if err != nil {
			return nil, fmt.Errorf("変換失敗！: %w", err)
		}
		resultLines = append(resultLines, convertedLine)
		c.lineConverter.NotifyConvertLineFinish(c.parseState)
	}

	return resultLines, nil
}

func writeLines(fileName string, lines []Line) error {
	file, err := os.Create(fileName)
	if err != nil {
		return fmt.Errorf("cannot open file %s: %w", fileName, err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, line := range lines {
		_, err = fmt.Fprintln(writer, line.content)
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
	// change default logger to debug level
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug})))

	lineConverter := LineConverter{}
	parseState := ParseState{identity, identity}
	commandProcessor := NewCommandProcessor()

	// register handlers
	commandProcessor.Register(CommentSingleLine, func(state *ParseState) error {
		state.Current = identity
		state.Next = isSingleLineCommenting
		return nil
	})
	commandProcessor.Register(UncommentSingleLine, func(state *ParseState) error {
		state.Current = identity
		state.Next = isSingleLineUncommenting
		return nil
	})
	commandProcessor.Register(CommentBlockBegin, func(state *ParseState) error {
		if state.Current == isBlockCommenting {
			return fmt.Errorf("nested block commenting is not implemented for now")
		}
		state.Current = identity
		state.Next = isBlockCommenting
		return nil
	})
	commandProcessor.Register(UncommentBlockBegin, func(state *ParseState) error {
		if state.Current == isBlockUncommenting {
			return fmt.Errorf("nested block uncommenting is not implemented for now")
		}
		state.Current = identity
		state.Next = isBlockUncommenting
		return nil
	})
	commandProcessor.Register(CommentBlockEnd, func(state *ParseState) error {
		if state.Current != isBlockCommenting {
			return fmt.Errorf("unpaired comment block end detected")
		}
		state.Current = identity
		state.Next = identity
		return nil
	})
	commandProcessor.Register(UncommentBlockEnd, func(state *ParseState) error {
		if state.Current != isBlockUncommenting {
			return fmt.Errorf("unpaired uncomment block end detected")
		}
		state.Current = identity
		state.Next = identity
		return nil
	})

	converter := Converter{&lineConverter, &parseState, commandProcessor}

	fileName := os.Args[1]
	lines, err := converter.Convert(fileName)
	if err != nil {
		log.Fatal(err)
	}

	outFileName := fileName + ".converted.cpp"
	err = writeLines(outFileName, lines)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Fprintf(os.Stderr, "successfully converted and wrote to %s \n", outFileName)
}
