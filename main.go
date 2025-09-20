package main

import (
	"fmt"
	"strings"

	"github.com/0xf0xx0/oigiki"
	"github.com/Delta456/box-cli-maker/v2"
	"github.com/gookit/color"
)

func main() {
	str := oigiki.ProcessTags(`
{/}    {blue}02{/}     {cyan}97 00     {blue}1a     {green}2f 70 6f 67 6f 6c 6f 20 2d 20 66 6f 73 73 20 69 73
{/}┌─────────┬─────┬──────────┬──────────────────────────────────────────────────┐
{/}│{blue}OP_PUSH_2{/}│ {cyan}151{/} │{blue}OP_PUSH_26{/}│            {green}/pogolo - foss is{/}                     │
{/}└─────────┴─────┴──────────┴──────────────────────────────────────────────────┘
{/} {green}20 66 72 65 65 64 6f 6d 2f   {blue}93   {red}59 12 12 00 00 00 00
{/}┌──────────────────────────┬──────┬────────────────────┐
{/}│         {green}freedom/{/}         │{blue}OP_ADD{/}│      {red}[error]{/}       │
{/}└──────────────────────────┴──────┴────────────────────┘
		`)
	str = str
	println(str)
	// script, _ := hex.DecodeString("0297001a2f706f676f6c6f202d20666f73732069732066726565646f6d2f9359121200000000")
	// tkzr := txscript.MakeScriptTokenizer(0, script)
	// for {
	// 	if tkzr.Done() {
	// 		break
	// 	}
	// 	tkzr.Next()
	// 	if err := tkzr.Err(); err != nil {
	// 		println(err.Error())
	// 		break
	// 	}
	// 	println(hex.EncodeToString([]byte{tkzr.Opcode()}),
	// 		hex.EncodeToString(tkzr.Data()))
	// }

	// println(fmt.Sprintf("%s", strings.Join(mergeBoxes([]string{
	// 	"OP_PUSH_2",
	// 	"151",
	// 	oigiki.ProcessTags("{cyan}OP_PUSH_26"),
	// 	"/pogolo - foss is",
	// }, []string{"02", oigiki.ProcessTags("{cyan}97 00"), "1a",
	// 	"2f 70 6f 67 6f 6c 6f 20 2d 20 66 6f 73 73 20 69 73",}), "\n")))
	// println(fmt.Sprintf("%s", strings.Join(mergeBoxes([]string{
	// 	" freedom/",
	// 	"OP_ADD",
	// }, []string{"20 66 72 65 65 64 6f 6d 2f", "93"}), "\n")))

	blocks := []string{
		"OP_PUSH_2",
		"151",
		"OP_PUSH_26",
		"/pogolo - foss is freedom/",
		"OP_ADD",
		"[error]",
	}
	b := make([][]string, 2)
	h := make([][]string, 2)
	headers := []string{
		"02",
		"97 00",
		"1a",
		"x2f 70 6f 67 6f 6c 6f 20 2d 20 66 6f 73 73 20 69 73 20 66 72 65 65 64 6f 6d 2f",
		"93",
		"59 12 12 00 00 00 00",
	}
	headerLen := 0
	currLine := 0
	for i, header := range headers {
		x := len(header)
		n := headerLen + x
		if n > 80 && header[0] == 'x' {
			x--
			n--
			header = header[1:]
			/// break it
			/// TODO: only if datatype is string
			diff := n - 80
			if x-diff < 0 {
				break
			}
			blockOffset := diff * 3
			headerOffset := x - blockOffset - diff - 3
			// headerOffset := (x-diff)/2

			blockChunk := blocks[i][blockOffset:]
			headerChunk := strings.TrimSpace(header[headerOffset:])
			blocks[i] = blocks[i][:blockOffset]
			headers[i] = header[:headerOffset]

			b[currLine] = blocks[:i+1]
			h[currLine] = headers[:i+1]
			currLine++
			b[currLine] = append(b[currLine], blockChunk)
			h[currLine] = append(h[currLine], headerChunk)
			headerLen = 0
		} else {
			b[currLine] = append(b[currLine], blocks[i])
			h[currLine] = append(h[currLine], headers[i])
		}
		headerLen += x
	}

	for x := range b {
		// println(fmt.Sprintf("%v", b[x]))
		println(fmt.Sprintf("%s", mergeBoxes(b[x], h[x])))
	}
}

func padString(str string, l int) string {
	x := len(color.ClearCode(str))
	if x < l {
		spaces := (l - x) / 2
		ret := strings.Repeat(" ", spaces) + str + strings.Repeat(" ", l-(spaces+x))
		return ret
	}
	return str
}

func mergeBoxes(boxes []string, headlines []string) string {
	b := box.New(
		box.Config{
			ContentAlign: "Center",
			Type:         "Single",
		},
	)
	ret := make([]string, 4)
	if len(boxes) == 1 {
		str := padString(boxes[0], len(color.ClearCode(headlines[0])))
		split := strings.Split(b.String("", str), "\n")
		if headlines[0] == "" {
			ret[0] += strings.Repeat(" ", len(split[1])-6)
		} else {
			/// ansi MUST be cleared here for proper len
			ret[0] += " " + padString(headlines[0], len(color.ClearCode(split[1]))-6)
		}
		ret[1] = split[0]
		ret[2] = split[1]
		ret[3] = split[2]
		return strings.Join(ret, "\n")
	}
	for i, str := range boxes {
		/// ansi MUST be cleared here for proper len
		str = padString(str, len(color.ClearCode(headlines[i])))
		str = b.String("", str)
		split := strings.Split(str, "\n")
		/// box height is always 3
		if i == 0 {
			/// furst box only needs the endcaps replaced
			ret[1] = split[0][:len(split[0])-3]
			ret[2] = split[1][:len(split[1])-3]
			ret[3] = split[2][:len(split[2])-3]
		} else if i == len(boxes)-1 {
			/// vice-versa for the last
			ret[1] += "┬" + split[0][3:]
			ret[2] += split[1]
			ret[3] += "┴" + split[2][3:]
		} else {
			/// replace both for the in-betweens, but only at the top and bottom
			temp := "┬" + split[0][3:]
			temp = temp[:len(temp)-3]
			ret[1] += temp

			ret[2] += split[1][:len(split[1])-3]

			temp = "┴" + split[2][3:]
			temp = temp[:len(temp)-3]
			ret[3] += temp
		}
		if headlines[i] == "" {
			ret[0] += strings.Repeat(" ", len(split[1])-6)
		} else {
			/// ansi MUST be cleared here for proper len
			ret[0] += " " + padString(headlines[i], len(color.ClearCode(split[1]))-6)
		}
	}
	return strings.Join(ret, "\n")
}

/// TODO: test file
//
// ret := b.String("", " freedom/")
// split := mergeBoxes([]string{ret})
// println(fmt.Sprintf("%s", strings.Join(split, "\n")))
// println(fmt.Sprintf("%s", strings.Join(mergeBoxes([]string{b.String("", "multi"),
// 	b.String("", "box")}), "\n")))

// println(fmt.Sprintf("%s", strings.Join(mergeBoxes([]string{b.String("", "multi"), b.String("", "box"), b.String("", "test")}), "\n")))
// println(fmt.Sprintf("%s", strings.Join(mergeBoxes([]string{
// 	b.String("", "multi"),
// 	b.String("", "box"),
// 	b.String("", "test"),
// 	b.String("", "quad"),
// }), "\n")))
// println(fmt.Sprintf("%s", strings.Join(mergeBoxes([]string{
// 	b.String("", "multi"),
// 	b.String("", "box"),
// 	b.String("", "test"),
// 	b.String("", "quadwfjkwfbwfvuwi"),
// 	b.String("", "quin"),
// 	b.String("", "sexbdugvebfuwu"),
// }), "\n")))
