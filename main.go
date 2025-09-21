package main

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"unicode"

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
	/// foundry usa 915753
	// script, _ := hex.DecodeString("0329f90d041b21d0682f466f756e6472792055534120506f6f6c202364726f70676f6c642ffabe6d6d5c91db6d6adc594b6223bf953075281941f2450c5ba328544b7e04dcdc177d9901000000000000007c0134999e620d0000000000")
	/// slushpool 915754
	script, _ := hex.DecodeString("032af90d102f736c7573682f6500b006a066120e01fabe6d6d560a0f5f0cf3bd4dbfc58629a834166b021716b115cc267ba92822755946c705100000000000000000007b434d00000000000000")

	// script, _ := hex.DecodeString("0346f90d1e2f706f676f6c6f202d20646563656e7472616c697a65206f72206469652f080000000000000000")
	// script, _ := hex.DecodeString("0237011e2f706f676f6c6f202d20646563656e7472616c697a65206f72206469652f08fe53956400000000")
	// script, _ := hex.DecodeString("0297001a2f706f676f6c6f202d20666f73732069732066726565646f6d2f9359121200000000")
	tkzr := txscript.MakeScriptTokenizer(0, script)
	reverseOpcodeMap := make(map[byte]string)
	/// TODO: OP_0 and OP_TRUE swap places
	/// prefer TRUE and FALSE
	for k, v := range txscript.OpcodeByName {
		if pk, ok := reverseOpcodeMap[v]; ok {
			if len(k) < len(pk) {
				continue
			}
		}
		reverseOpcodeMap[v] = k
	}
	blks := parseCoinbaseScript(tkzr, reverseOpcodeMap)//[:5]

	for _, line := range chunkData(blks) {
		println(fmt.Sprintf("%s", mergeBoxes(line)))
	}
}

func parseCoinbaseScript(tkzr txscript.ScriptTokenizer, reverseOpcodeMap map[byte]string) []block {
	blks := make([]block, 0, 20)
	i := 0
	for {
		if tkzr.Done() {
			break
		}
		tkzr.Next()
		if err := tkzr.Err(); err != nil {
			println(err.Error())
			blk := block{
				Header: prettyprintHex(tkzr.Script()[tkzr.ByteIndex():]),
				Body:   "[error]",
			}
			blks = append(blks, blk)
			break
		}
		x := reverseOpcodeMap[tkzr.Opcode()]
		blk := block{
			Header: hex.EncodeToString([]byte{tkzr.Opcode()}),
			Body:   x,
			Type:   DATATYPE_OPCODE,
		}
		blks = append(blks, blk)
		if len(tkzr.Data()) > 0 {
			hdr := prettyprintHex(tkzr.Data())
			blk2 := block{
				Header: hdr,
			}
			if strings.Contains(x, "DATA") {
				if i == 0 {
					d := make([]byte, len(tkzr.Data()))
				 	copy(d, tkzr.Data())
					if len(d) % 2 == 1 {
						d = append(d, 0)
					}
					if len(d) == 2 {
						blk2.Body = strconv.Itoa(int(binary.LittleEndian.Uint16(d)))
					} else if len(d) == 4 {
						blk2.Body = strconv.Itoa(int(binary.LittleEndian.Uint32(d)))
					}
					blk2.Type = DATATYPE_INT
				} else {
					blk2.Body = string(tkzr.Data())
					blk2.Type = DATATYPE_STRING
				}
			}
			blks = append(blks, blk2)
		}
		i++
	}
	return blks
}

//func parseBlockHeader()

func prettyprintHex(b []byte) string {
	hex := hex.EncodeToString(b)
	ret := ""

	for x := 0; x < len(hex); x += 2 {
		if x > 0 {
			ret += " "
		}
		ret += string([]byte{hex[x], hex[x+1]})
	}
	return ret
}
func chunkData(blocks []block) [][]block {
	ret := make([][]block, 1, 2)
	headerLen := 0
	currLine := 0
	for _, blk := range blocks {
		x := max(len(blk.Header), len(blk.Body))
		lineLengthMax := 80
		estimatedLen := headerLen + x
		if estimatedLen > lineLengthMax {
			if blk.Type == DATATYPE_STRING {
				/// break it
				var splitData []string
				data := blk.Body
				/// -2 cause borders
				maxRunesCurrentLine := (lineLengthMax-headerLen)/3 - 2
				// Second term has lengthLineMax/3 added; for data of length 27 and line length 15, I'll want to create one slice at [0,15] and a second at [15,27] (clamping 30 to 27), so the loop needs to go "one past"
				for i := maxRunesCurrentLine; i < len(data)+lineLengthMax/3; i += lineLengthMax / 3 {
					dataStart := max(i-lineLengthMax/3, 0) // Previous iteration, or start of string
					dataEnd := max(min(i, len(data)), 0)   // Current iteration, or end of string
					if dataStart != dataEnd {
						splitData = append(splitData, string(data[dataStart:dataEnd]))
					}
				}
				var blocks []block
				if len(splitData) == 0 {
					blocks = append(blocks, blk)
				} else {
					for _, d := range splitData {
						blocks = append(blocks, block{
							Header: prettyprintHex([]byte(d)),
							Body:   d,
							Type: blk.Type,
						})
					}
					ret[currLine] = append(ret[currLine], blocks[0])
					for _,b := range blocks[1:] {
						x += len(b.Header)
						/// TODO: is this it? just headerLen+x?
						if headerLen+x > lineLengthMax {
							currLine++
							ret = append(ret, make([]block, 0, 3))
							x = len(b.Header)
						}
						ret[currLine] = append(ret[currLine], b)
					}
					headerLen = 0
				}
			} else {
				currLine++
				ret = append(ret, make([]block, 0, 3))
				ret[currLine] = append(ret[currLine], blk)
				headerLen = 0
			}
		} else {
			ret[currLine] = append(ret[currLine], blk)
		}
		headerLen += x
	}
	return ret
}

func padString(str string, l int) string {
	x := len(color.ClearCode(str))
	y := len(str)
	ansiLen := y - x
	if x < l {
		spaces := (l - x) / 2
		ret := strings.Repeat(" ", l-(spaces+x)) + str + strings.Repeat(" ", l-(spaces+x))
		return ret[:l+ansiLen]
	}
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

	for i, blk := range boxes {
		/// ansi MUST be cleared here for proper len
		str := replaceNonPrintable(blk.Body)
		str = padString(colorText(blk.Type, str), len(color.ClearCode(boxes[i].Header)))
		x := str
		str = b.String("", str)
		// str = b.String("", colorText(blk.Type, str))
		split := strings.Split(str, "\n")
		/// box height is always 3
		if i == 0 && len(boxes) == 1 {
			ret[1] = split[0]
			ret[2] = split[1]
			ret[3] = split[2]
		} else if i == 0 {
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
			ret[0] += strings.Repeat(" ", len(color.ClearCode(x)))
		} else {
			/// ansi MUST be cleared here for proper len
			ret[0] += padString(colorText(blk.Type, blk.Header), len(color.ClearCode(x))+1)
		}
	}
	ret[0] = strings.TrimRight(ret[0], " ")
	return strings.Join(ret, "\n")
}
func replaceNonPrintable(input string) string {
	result := []rune{}

	for _, r := range input {
		if unicode.IsPrint(r) && (r >= 0x20 && r <= 0x7E) {
			result = append(result, r)
		} else {
			result = append(result, '.')
		}
	}

	return string(result)
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
