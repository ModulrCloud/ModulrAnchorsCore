package utils

import (
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/modulrcloud/modulr-anchors-core/databases"
	"github.com/modulrcloud/modulr-anchors-core/structures"

	"lukechampine.com/blake3"
)

// ANSI escape codes for text colors
const (
	RESET_COLOR       = "\033[0m"
	RED_COLOR         = "\033[31;1m"
	DEEP_GREEN_COLOR  = "\u001b[38;5;23m"
	DEEP_YELLOW       = "\u001b[38;5;214m"
	GREEN_COLOR       = "\033[32;1m"
	YELLOW_COLOR      = "\033[33m"
	MAGENTA_COLOR     = "\033[38;5;99m"
	CYAN_COLOR        = "\033[36;1m"
	WHITE_COLOR       = "\033[37;1m"
	TIMESTAMP_COLOR   = "\u001b[38;5;245m"
	TIMESTAMP_BRACKET = "\u001b[38;5;66m"
	PID_COLOR         = "\u001b[38;5;140m"
	DIVIDER_COLOR     = "\u001b[38;5;240m"
)

var SHUTDOWN_ONCE sync.Once

var kernelBanner = []string{
	"modulr anchors kernel bootstrap",
	"consensus mesh primed • transport warm",
	"telemetry uplink locked • shell glow ready",
}

func GracefulShutdown() {

	SHUTDOWN_ONCE.Do(func() {

		LogWithTime("Stop signal has been initiated.Keep waiting...", CYAN_COLOR)

		LogWithTime("Closing server connections...", CYAN_COLOR)

		if err := databases.CloseAll(); err != nil {
			LogWithTime(fmt.Sprintf("failed to close databases: %v", err), RED_COLOR)
		}

		LogWithTime("Node was gracefully stopped", GREEN_COLOR)

		os.Exit(0)

	})

}

func LogWithTime(msg, msgColor string) {

	formattedDate := time.Now().Format("02 January 2006 15:04:05")
	timestampLabel := fmt.Sprintf("%s[%s%s%s]%s", TIMESTAMP_BRACKET, TIMESTAMP_COLOR, formattedDate, TIMESTAMP_BRACKET, RESET_COLOR)
	pidLabel := fmt.Sprintf("%s(pid:%d)%s", PID_COLOR, os.Getpid(), RESET_COLOR)
	divider := fmt.Sprintf("%s ┇ %s", DIVIDER_COLOR, RESET_COLOR)

	fmt.Printf("%s %s%s%s%s\n", timestampLabel, pidLabel, divider, msgColor, msg+RESET_COLOR)

}

func PrintKernelBanner() {
	PrintShellDivider()
	textColors := []string{CYAN_COLOR, GREEN_COLOR, YELLOW_COLOR, MAGENTA_COLOR, WHITE_COLOR}
	accentPalette := []string{"\u001b[38;5;81m", "\u001b[38;5;117m", "\u001b[38;5;159m"}
	maxWidth := 0
	for _, line := range kernelBanner {
		if len(line) > maxWidth {
			maxWidth = len(line)
		}
	}
	innerWidth := maxWidth + 2
	LogWithTime(buildBannerBorder('╭', '╮', innerWidth, accentPalette), "")
	for idx, line := range kernelBanner {
		textColor := textColors[idx%len(textColors)]
		frameColor := accentPalette[idx%len(accentPalette)]
		LogWithTime(buildBannerLine(line, maxWidth, frameColor, textColor), "")
	}
	LogWithTime(buildBannerBorder('╰', '╯', innerWidth, accentPalette), "")
	PrintShellDivider()
}

func PrintShellDivider() {
	palette := []string{CYAN_COLOR, MAGENTA_COLOR, YELLOW_COLOR, GREEN_COLOR}
	glyphs := []rune{'╺', '━', '╸', '╾'}
	var builder strings.Builder
	width := 48
	for i := 0; i < width; i++ {
		builder.WriteString(palette[i%len(palette)])
		builder.WriteRune(glyphs[i%len(glyphs)])
	}
	builder.WriteString(RESET_COLOR)
	LogWithTime(builder.String(), "")
}

func buildBannerBorder(left, right rune, innerWidth int, palette []string) string {
	var builder strings.Builder
	paletteLen := len(palette)
	for idx := 0; idx < paletteLen; idx++ {
		if palette[idx] == "" {
			palette[idx] = WHITE_COLOR
		}
	}
	builder.WriteString(palette[0])
	builder.WriteRune(left)
	for i := 0; i < innerWidth; i++ {
		builder.WriteString(palette[(i+1)%paletteLen])
		builder.WriteRune('─')
	}
	builder.WriteString(palette[(innerWidth+1)%paletteLen])
	builder.WriteRune(right)
	builder.WriteString(RESET_COLOR)
	return builder.String()
}

func buildBannerLine(content string, maxWidth int, frameColor, textColor string) string {
	var builder strings.Builder
	padding := maxWidth - len(content)
	if padding < 0 {
		padding = 0
	}
	builder.WriteString(frameColor)
	builder.WriteRune('│')
	builder.WriteString(RESET_COLOR)
	builder.WriteString(" ")
	builder.WriteString(textColor)
	builder.WriteString(content)
	builder.WriteString(strings.Repeat(" ", padding))
	builder.WriteString(RESET_COLOR)
	builder.WriteString(" ")
	builder.WriteString(frameColor)
	builder.WriteRune('│')
	builder.WriteString(RESET_COLOR)
	return builder.String()
}

func Blake3(data string) string {

	blake3Hash := blake3.Sum256([]byte(data))

	return hex.EncodeToString(blake3Hash[:])

}

func GetUTCTimestampInMilliSeconds() int64 {

	return time.Now().UTC().UnixMilli()

}

func EpochStillFresh(epochHandler *structures.EpochDataHandler, networkParams *structures.NetworkParameters) bool {

	return (epochHandler.StartTimestamp + uint64(networkParams.EpochDuration)) > uint64(GetUTCTimestampInMilliSeconds())

}

func SignalAboutEpochRotationExists(epochIndex int) bool {

	keyValue := []byte("EPOCH_FINISH:" + strconv.Itoa(epochIndex))

	if readyToChangeEpochRaw, err := databases.FINALIZATION_VOTING_STATS.Get(keyValue, nil); err == nil && string(readyToChangeEpochRaw) == "TRUE" {

		return true

	}

	return false

}
