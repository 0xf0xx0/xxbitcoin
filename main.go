package main

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/0xf0xx0/oigiki"
	"github.com/Delta456/box-cli-maker/v2"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/gookit/color"
	"github.com/urfave/cli/v3"
)

const (
	DATATYPE_ERR = iota
	DATATYPE_OPCODE
	DATATYPE_NUMBER
	DATATYPE_STRING
	DATATYPE_HASH
	DATATYPE_TIME
	DATATYPE_MISC
)

type block struct {
	Header, Body string
	Type         int
}

func main() {
	app := &cli.Command{
		Name:                   "xxbitcoin",
		Usage:                  "pretty-prints bitcoin structures",
		UsageText:              "xxbitcoin [options]",
		Version:                "0.0.1",
		UseShortOptionHandling: true,
		EnableShellCompletion:  true,
		ReadArgsFromStdin:      true,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "type",
				Aliases:  []string{"t"},
				Usage:    "`data structure` type (script, header, coinbasescript)",
				Required: true,
			},
		},
		Action: func(_ context.Context, ctx *cli.Command) error {
			blks := []block{}
			rawInput := ctx.Args().Get(0)
			/// TODO: clean this up
			if len(rawInput) == 0 {
				return fmt.Errorf("no input")
			}
			input, err := hex.DecodeString(rawInput)
			if err != nil {
				println(rawInput)
				return err
			}
			switch ctx.String("type") {
			case "script":
				{
					blks = parseScript(input)
				}
			case "coinbasescript":
				{
					blks = parseCoinbaseScript(input)
				}
			case "header":
				{
					blks = parseBlockHeader(input)
				}
			default:
				{
					return fmt.Errorf("yo wtf")
				}
			}
			for _, line := range chunkData(blks) {
				println(fmt.Sprintf("%s", mergeBoxes(line)))
			}
			return nil
		},
	}

	/// foundry usa 915753
	// script, _ := hex.DecodeString("0329f90d041b21d0682f466f756e6472792055534120506f6f6c202364726f70676f6c642ffabe6d6d5c91db6d6adc594b6223bf953075281941f2450c5ba328544b7e04dcdc177d9901000000000000007c0134999e620d0000000000")
	/// slushpool 915754
	/// edge case: string chunk overflow
	// script, _ := hex.DecodeString("032af90d102f736c7573682f6500b006a066120e01fabe6d6d560a0f5f0cf3bd4dbfc58629a834166b021716b115cc267ba92822755946c705100000000000000000007b434d00000000000000")

	/// edge case: op overflow
	// script, _ := hex.DecodeString("47304402204113c4e58ccdedb5b633483720f8e9837b89c58847d6e4d033c576a5e2a295f702200b867d08291f884e491772b61fd03127a1e0efbdef87442c97f334d1aa4aef9901410426cbe208d7e5bf3b5b38ab0b05d0ddd8ef8ea06fe946d622ce5db29a27d3d020b3ffa629dda1201323b1dada376867201df1de49455e898ecc315511a84507da")

	// script, _ := hex.DecodeString("0237011e2f706f676f6c6f202d20646563656e7472616c697a65206f72206469652f08fe53956400000000")
	/// edge case: error overflow >:C
	// script, _ := hex.DecodeString("0297001a2f706f676f6c6f202d20666f73732069732066726565646f6d2f9359121200000000")
	/// edge case: error AND op overflow >:CC
	// header, _ := hex.DecodeString("00005823301b30a68438ea562def4083f1b6151883e4858eaa1a01000000000000000000c6fe1e8c906cae6ea804a2cffbc2afe8a750985ae991fcb49e71003e3a7e16d32029d06838fa01171c223ef2")

	if err := app.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}

func parseCoinbaseScript(script []byte) []block {
	blks := parseScript(script)
	blks[1].Type = DATATYPE_NUMBER
	/// pad to uint32
	b := make([]byte, 4)
	copy(b, []byte(blks[1].Body))
	// diff := 4 - len(b)
	// if diff > 0 {
	// 	b = append([]byte(blks[1].Body), make([]byte, diff)...)
	// }
	blks[1].Body = strconv.Itoa(int(binary.LittleEndian.Uint32(b)))
	return blks
}
func parseScript(script []byte) []block {
	blks := make([]block, 0, 20)
	tkzr := txscript.MakeScriptTokenizer(0, script)
	reverseOpcodeMap := make(map[byte]string)
	for k, v := range txscript.OpcodeByName {
		if pk, ok := reverseOpcodeMap[v]; ok {
			/// TODO: aliased ops switch places
			/// prefer the longer one
			if len(k) < len(pk) {
				continue
			}
		}
		reverseOpcodeMap[v] = k
	}
	i := 0
	for {
		if tkzr.Done() {
			break
		}
		tkzr.Next()
		if err := tkzr.Err(); err != nil {
			opcode := tkzr.Script()[tkzr.ByteIndex() : tkzr.ByteIndex()+1]
			blk := block{
				Header: prettyprintHex(opcode),
				Body:   reverseOpcodeMap[opcode[0]],
				Type:   DATATYPE_OPCODE,
			}
			blk2 := block{
				Header: prettyprintHex(tkzr.Script()[tkzr.ByteIndex()+1:]),
				Body:   err.Error(),
			}
			blks = append(blks, blk, blk2)
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
				blk2.Body = string(tkzr.Data())
				blk2.Type = DATATYPE_STRING
			}
			blks = append(blks, blk2)
		}
		i++
	}
	return blks
}

func parseBlockHeader(input []byte) []block {
	buf := bytes.NewBuffer(input)
	blks := make([]block, 6)
	ver := make([]byte, 4)
	prevBlockHash := [32]byte{}
	merkleRoot := [32]byte{}
	blockTime := make([]byte, 4)
	nbits := make([]byte, 4)
	nonce := make([]byte, 4)

	_, err := buf.Read(ver)
	if err != nil {
		/// TODO
	}
	_, err = buf.Read(prevBlockHash[:])
	_, err = buf.Read(merkleRoot[:])
	_, err = buf.Read(blockTime)
	_, err = buf.Read(nbits)
	_, err = buf.Read(nonce)

	blks[0] = block{
		Header: prettyprintHex(ver),
		Body:   strconv.Itoa(int(binary.LittleEndian.Uint32(ver))),
		Type:   DATATYPE_NUMBER,
	}
	blks[1] = block{
		Header: prettyprintHex(prevBlockHash[:]),
		Body:   chainhash.Hash(prevBlockHash).String(),
		Type:   DATATYPE_HASH,
	}
	blks[2] = block{
		Header: prettyprintHex(merkleRoot[:]),
		Body:   chainhash.Hash(merkleRoot).String(),
		Type:   DATATYPE_HASH,
	}
	blks[3] = block{
		Header: prettyprintHex(blockTime),
		Body:   time.Unix(int64(binary.LittleEndian.Uint32(blockTime)), 0).String(),
		Type:   DATATYPE_TIME,
	}
	blks[4] = block{
		Header: prettyprintHex(nbits),
		Body:   strconv.Itoa(int(binary.LittleEndian.Uint32(nbits))),
		Type:   DATATYPE_NUMBER,
	}
	blks[5] = block{
		Header: prettyprintHex(nonce),
		Body:   strconv.Itoa(int(binary.LittleEndian.Uint32(nonce))),
		Type:   DATATYPE_NUMBER,
	}
	return blks
}

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

// / chunk a line of blocks into multiple lines
func chunkData(blocks []block) [][]block {
	ret := make([][]block, 1, 2)
	headerLen := 0
	currLine := 0
	/// TODO: refactor and cleanup
	for _, blk := range blocks {
		lineLengthMax := 80
		x := max(len(blk.Header), len(blk.Body))
		estimatedLen := headerLen + x
		if estimatedLen > lineLengthMax {
			switch blk.Type {
			case DATATYPE_TIME:
				fallthrough
			case DATATYPE_STRING:
				{
					/// break it
					currLine, x = chunkStringBlock(blk, lineLengthMax, headerLen, &ret, currLine, x)
				}
			case DATATYPE_ERR:
				/// FIXME: why do long error blocks get chunked too late?
				/// try with script 30450221009d4cdcb330786e787164a025abca8a0655f3803da66ef18d14745819b7e28b6b02200d98881bdd055ff2da39e61a858936866610c7e878ea62f08c5ee1534532c14a01
				lineLengthMax -= 2
				fallthrough
			case DATATYPE_MISC:
				fallthrough
			case DATATYPE_HASH:
				{
					currLine, x = chunkBlock(blk, lineLengthMax, headerLen, &ret, currLine, x)
				}
			default:
				{
					/// move it to the next line and blindly assume its short enough
					/// to display without overflow
					currLine++
					ret = append(ret, make([]block, 0, 3))
					ret[currLine] = append(ret[currLine], blk)
				}
			}
			headerLen = 0
		} else {
			ret[currLine] = append(ret[currLine], blk)
		}
		headerLen += x
	}
	return ret
}

// / TODO: replace chunkStringBlock
// / chunks a long block across multiple lines
func chunkBlock(blk block, lineLengthMax int, currLen int, ret *[][]block, currLine int, x int) (int, int) {
	var splitData []string
	var splitBody []string
	header := blk.Header
	headerLen := len(header)
	bodyLen := len(blk.Body)
	/// -2 cause borders
	maxRunesCurrentLine := lineLengthMax - currLen - 4

	var blocks []block
	for i := maxRunesCurrentLine; i < headerLen+lineLengthMax; i += lineLengthMax {
		dataStart := max(i-lineLengthMax, 0) // Previous iteration, or start of string
		dataEnd := max(min(i, headerLen), 0) // Current iteration, or end of string
		if dataStart != dataEnd {
			x1 := strings.TrimSpace(string(header[dataStart:dataEnd]))
			splitData = append(splitData, x1)
			scale := float32(1)
			/// if the body is longer, scale our chunking down to fit within the max width
			if bodyLen > headerLen {
				scale = float32(headerLen) / float32(bodyLen)
			}
			startratio := float32(0)
			endratio := float32(dataEnd) / float32(headerLen)
			if dataStart > 0 {
				startratio = float32(dataStart) / float32(headerLen)
			}
			start := int(float32(bodyLen) * startratio * scale)
			/// but for the last chunk, we want to grab all remaining chars
			/// we can safely assume the chunk is less than max width
			if i+lineLengthMax > headerLen+lineLengthMax {
				scale = 1
			}
			end := int(float32(bodyLen) * endratio * scale)
			x2 := strings.TrimSpace(string(blk.Body[start:end]))
			splitBody = append(splitBody, x2)
			newBlk := block{
				Header: x1,
				Body:   x2,
				Type:   blk.Type,
			}
			blocks = append(blocks, newBlk)
		}
	}

	(*ret)[currLine] = append((*ret)[currLine], blocks[0])
	if len(blocks) > 1 {
		for _, b := range blocks[1:] {
			x += len(b.Header)
			/// TODO: is this it? just currLen+x?
			if currLen+x > lineLengthMax {
				currLine++
				(*ret) = append((*ret), make([]block, 0, 3))
				x = len(b.Header)
			}
			(*ret)[currLine] = append((*ret)[currLine], b)
		}
	}

	return currLine, x
}

// / same as chunkBlock but optimized for strings
func chunkStringBlock(blk block, lineLengthMax int, headerLen int, ret *[][]block, currLine int, x int) (int, int) {
	var splitData []string
	var splitHdr []string
	data := blk.Body
	dataLen := len(data)
	/// -4 cause borders? i think? majik number that seems to work
	maxRunesCurrentLine := (lineLengthMax - headerLen - 4) / 3
	// Second term has lengthLineMax/3 added; for data of length 27 and line length 15, I'll want to create one slice at [0,15] and a second at [15,27] (clamping 30 to 27), so the loop needs to go "one past"
	for i := maxRunesCurrentLine; i < dataLen+lineLengthMax/3; i += lineLengthMax / 3 {
		dataStart := max(i-lineLengthMax/3, 0) // Previous iteration, or start of string
		dataEnd := max(min(i, dataLen), 0)   // Current iteration, or end of string
		if dataStart != dataEnd {
			splitData = append(splitData, data[dataStart:dataEnd])
			/// take a bite out of the header too
			/// for string this is nice cause its always a 3:1 ratio, so we'll always
			/// select whole bytes
			startratio := float32(0)
			endratio := float32(dataEnd) / float32(dataLen)
			if dataStart > 0 {
				startratio = float32(dataStart) / float32(dataLen)
			}
			start := int(float32(len(blk.Header)) * startratio)
			end := int(float32(len(blk.Header)) * endratio)
			splitHdr = append(splitHdr, strings.TrimSpace(string(blk.Header[start:end])))
		}
	}
	if len(splitData) == 0 {
		(*ret)[currLine] = append((*ret)[currLine], blk)
	} else {
		var blocks []block
		for x, d := range splitData {
			x = x
			blocks = append(blocks, block{
				Header: splitHdr[x],
				Body:   d,
				Type:   blk.Type,
			})
		}
		(*ret)[currLine] = append((*ret)[currLine], blocks[0])
		for _, b := range blocks[1:] {
			x += len(b.Header)
			/// TODO: is this it? just headerLen+x?
			if headerLen+x > lineLengthMax {
				currLine++
				(*ret) = append((*ret), make([]block, 0, 3))
				x = len(b.Header)
			}
			(*ret)[currLine] = append((*ret)[currLine], b)
		}
	}
	return currLine, x
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

// / merge boxes together into one box line
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
	result := strings.Builder{}
	result.Grow(len(input))

	for _, r := range input {
		if unicode.IsPrint(r) && (r >= 0x20 && r <= 0x7E) {
			result.WriteRune(r)
		} else {
			result.WriteByte('.')
		}
	}

	return result.String()
}
func colorText(datatype int, text string) string {
	color := "whitebright"
	switch datatype {
	case DATATYPE_TIME:
		fallthrough
	case DATATYPE_NUMBER:
		{
			color = "blue"
		}
	case DATATYPE_HASH:
		fallthrough
	case DATATYPE_STRING:
		{
			color = "green"
		}
	case DATATYPE_MISC:
		{
			color = "magenta"
		}
	case DATATYPE_OPCODE:
		{
			color = "yellow"
		}
	case DATATYPE_ERR:
		{
			color = "red"
		}
	}
	return oigiki.ProcessTags(fmt.Sprintf("{%s}%s{/}", color, text))
}
