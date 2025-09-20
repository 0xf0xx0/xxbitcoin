package main

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/0xf0xx0/oigiki"
	"github.com/Delta456/box-cli-maker/v2"
	"github.com/btcsuite/btcd/txscript"
	"github.com/gookit/color"
)

const (
	DATATYPE_UNK = iota
	DATATYPE_OPCODE
	DATATYPE_STRING
	DATATYPE_INT
	DATATYPE_FLOAT
)

type block struct {
	Header, Body string
	Type         int
}

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
	///strings.ToValidUTF8(s string, replacement string)
	println(str)
	script, _ := hex.DecodeString("0297001a2f706f676f6c6f202d20666f73732069732066726565646f6d2f9359121200000000")
	tkzr := txscript.MakeScriptTokenizer(0, script)
	reverseOpcodeMap := make(map[byte]string)
	for k, v := range txscript.OpcodeByName {
		reverseOpcodeMap[v] = k
	}
	blks := make([]block, 0, 20)
	for {
		if tkzr.Done() {
			break
		}
		tkzr.Next()
		if err := tkzr.Err(); err != nil {
			println(err.Error())
			break
		}
		x := reverseOpcodeMap[tkzr.Opcode()]
		blk := block{
			Header: hex.EncodeToString([]byte{tkzr.Opcode()}),
			Body:   x,
			Type:   DATATYPE_OPCODE,
		}
		blks = append(blks, blk)
		// println(fmt.Sprintf("%+v", blk))
		if len(tkzr.Data()) > 0 {
			hdr := prettyprintHex(tkzr.Data())
			blk2 := block{
				Header: hdr,
			}
			if strings.Contains(x, "DATA") {
				blk2.Body = strings.ToValidUTF8(string(tkzr.Data()), "..")
			}
			// println(fmt.Sprintf("%+v", blk2))
			blks = append(blks, blk2)
		}
	}
	/// furst pushed bytes are the block height
	blks[1].Type = DATATYPE_INT
	/// next is the miner tag
	blks[3].Type = DATATYPE_STRING

	for _, line := range chunkData(blks) {
		println(fmt.Sprintf("%s", mergeBoxes(line)))
	}
}

func prettyprintHex(b []byte) string {
	hex := hex.EncodeToString(b)
	ret := ""
	// println(len(hex))
	for x := 0; x < len(hex); x += 2 {
		ret += string([]byte{hex[x], hex[x+1]})
		ret += " "
	}
	return strings.TrimSpace(ret)
}
func chunkData(blocks []block) [][]block {
	ret := make([][]block, 1, 2)
	headerLen := 0
	currLine := 0
	for _, header := range blocks {
		x := len(header.Header)
		n := headerLen + x
		if n > 80 {
			if header.Type == DATATYPE_STRING {
				/// break it
				for {
					if n <= 80 {
						break
					}
					diff := n - 80
					if x-diff < 0 {
						break
					}
					newBlk := block{Type: header.Type}
					blockOffset := diff * 3
					headerOffset := x - blockOffset - diff - 2
					// headerOffset := (x-diff)/2
					newBlk.Body = header.Body[:blockOffset]
					newBlk.Header = header.Header[:headerOffset]
					ret[currLine] = append(ret[currLine], newBlk)

					bodyChunk := header.Body[blockOffset:]
					headerChunk := strings.TrimSpace(header.Header[headerOffset:])

					currLine++
					ret = append(ret, make([]block, 0, 3))
					newBlk = block{Type: header.Type}
					newBlk.Body = bodyChunk
					newBlk.Header = headerChunk
					ret[currLine] = append(ret[currLine], newBlk)

					headerLen = 0
					x = len(headerChunk)
					n = diff
				}
			} else {
				currLine++
				ret = append(ret, make([]block, 0, 3))
				ret[currLine] = append(ret[currLine], header)
				headerLen = 0
			}
		} else {
			ret[currLine] = append(ret[currLine], header)
		}
		headerLen += x
	}
	return ret
}

func padString(str string, l int) string {
	x := len(str)
	if x < l {
		spaces := (l - x) / 2
		ret := strings.Repeat(" ", spaces) + str + strings.Repeat(" ", l-(spaces+x))
		return ret
	}
	// println("str: ", str)
	return str
}

func mergeBoxes(boxes []block) string {
	b := box.New(
		box.Config{
			ContentAlign: "Center",
			Type:         "Single",
		},
	)
	ret := make([]string, 4)
	if len(boxes) == 1 {
		/// TODO: figure out how to add ansi here???
		str := padString(boxes[0].Body, len(color.ClearCode(boxes[0].Header)))
		split := strings.Split(b.String("", str), "\n")
		if boxes[0].Header == "" {
			ret[0] += strings.Repeat(" ", len(split[1])-6)
		} else {
			/// ansi MUST be cleared here for proper len
			ret[0] += " " + padString(boxes[0].Header, len(color.ClearCode(split[1]))-6)
		}
		ret[1] = split[0]
		ret[2] = split[1]
		ret[3] = split[2]
		return strings.Join(ret, "\n")
	}
	for i, blk := range boxes {
		/// ansi MUST be cleared here for proper len
		str := padString(blk.Body, len(color.ClearCode(boxes[i].Header)))
		str = b.String("", colorText(blk.Type, str))
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
		if boxes[i].Header == "" {
			ret[0] += strings.Repeat(" ", len(split[1])-6)
		} else {
			/// ansi MUST be cleared here for proper len
			ret[0] += " " + padString(boxes[i].Header, len(color.ClearCode(split[1]))-6)
		}
	}
	return strings.Join(ret, "\n")
}
func colorText(datatype int, text string) string {
	color := ""
	switch datatype {
	case DATATYPE_INT:
		fallthrough
	case DATATYPE_FLOAT:
		{
			color = "blue"
		}
	case DATATYPE_STRING:
		{
			color = "green"
		}
	case DATATYPE_OPCODE:
		{
			color = "yellow"
		}
	case DATATYPE_UNK:
		{
			color = "red"
		}
	}
	return oigiki.ProcessTags(fmt.Sprintf("{%s}%s{/}", color, text))
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
