package main

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/0xf0xx0/oigiki"
	"github.com/Delta456/box-cli-maker/v2"
	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/txscript"
	"github.com/urfave/cli/v3"
)

func main() {
	app := &cli.Command{
		Name:                   "xxbitcoin",
		Usage:                  "pretty-prints bitcoin structures",
		Version:                "0.0.1",
		UseShortOptionHandling: true,
		EnableShellCompletion:  true,
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

			/// TODO: handle binary
			rawInput := ctx.Args().Get(0)
			if len(rawInput) == 0 {
				b, err := io.ReadAll(os.Stdin)
				if err != nil {
					panic(err)
				}
				if len(b) == 0 {
					return fmt.Errorf("no input")
				}
				rawInput = string(b)
			}
			input, err := hex.DecodeString(strings.TrimSpace(rawInput))
			if err != nil {
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
			case "tx":
				{
					blks = parseTransaction(input, false)
				}
			case "coinbasetx":
				{
					blks = parseTransaction(input, true)
				}
			default:
				{
					return fmt.Errorf("unknown data structure")
				}
			}
			for _, line := range chunkData(blks) {
				fmt.Printf("%s\n", mergeBoxes(line))
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
	// script, _ := hex.DecodeString("03b1300e048c0d52692f466f756e6472792055534120506f6f6c202364726f70676f6c642ffabe6d6da6604f6ae857cce919cdbc846baf047fedb7f863b8632fa69e3289c3a4ddf6a1010000000000000040a0e579b002000000000000")
	/// edge case: barely error overflow >:CC
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
				Type:   DATATYPE_ERR,
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

// / most of these are manually parsed so we get access to the raw bytes
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
		Type:   DATATYPE_HEX,
	}
	blks[2] = block{
		Header: prettyprintHex(merkleRoot[:]),
		Body:   chainhash.Hash(merkleRoot).String(),
		Type:   DATATYPE_HEX,
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

// / bip 141
func parseTransaction(rawtxn []byte, isCoinbase bool) []block {
	blks := make([]block, 0, 20)
	buf := bytes.NewBuffer(rawtxn)
	ver := make([]byte, 4)
	buf.Read(ver)
	blks = append(blks, block{
		Header: prettyprintHex(ver),
		Body:   strconv.Itoa(int(binary.LittleEndian.Uint32(ver))),
		Type:   DATATYPE_NUMBER,
	})

	isWitness := false
	/// "peek" at the next byte
	if buf.Bytes()[0] == 0 {
		isWitness = true
		marker, _ := buf.ReadByte()
		flag, _ := buf.ReadByte()
		blks = append(blks, block{
			Header: prettyprintHex([]byte{marker}),
			Body:   strconv.Itoa(int(binary.LittleEndian.Uint16([]byte{marker, 0}))),
			Type:   DATATYPE_NUMBER,
		}, block{
			Header: prettyprintHex([]byte{flag}),
			Body:   strconv.Itoa(int(binary.LittleEndian.Uint16([]byte{flag, 0}))),
			Type:   DATATYPE_NUMBER,
		})
	}

	txInCount, txInCountBlk := readCompactSize(buf)
	blks = append(blks, txInCountBlk)
	slices.Grow(blks, txInCount*13) /// prevout+scriptlen+sequence makes 3, then 10 for the script
	for i := 0; i < txInCount; i++ {
		b := readTxin(buf, isCoinbase)
		blks = append(blks, b...)
	}
	txOutCount, txOutCountBlk := readCompactSize(buf)
	blks = append(blks, txOutCountBlk)
	isWitness = isWitness
	slices.Grow(blks, txOutCount*12)
	for i := 0; i < txOutCount; i++ {
		blks = append(blks, readTxout(buf)...)
	}
	isWitness = isWitness
	if isWitness {
		for i := 0; i < txInCount; i++ {
			count, sizeblk := readCompactSize(buf)
			blks = append(blks, sizeblk)
			for ii := 0; ii < count; ii++ {
				len, sizeblk := readCompactSize(buf)
				blks = append(blks, sizeblk)
				if len == 0 {
					continue
				}
				witness := make([]byte, len)
				buf.Read(witness)
				blks = append(blks, block{
					Header: prettyprintHex(witness),
					Body:   hex.EncodeToString(witness),
					Type:   DATATYPE_HEX,
				})
			}
		}
	}
	blks = append(blks, readLocktime(buf))
	return blks
}

// func parseFullBlock(rawtxn []byte) []block

// / "b00b69" -> "b0 0b 69"
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

func readCompactSize(buf *bytes.Buffer) (int, block) {
	initial, _ := buf.ReadByte()
	res := 0
	num := []byte{}

	switch initial {
	/// uint16
	case 0xfd:
		{
			num = make([]byte, 2)
			buf.Read(num)
			res = int(binary.LittleEndian.Uint16(num))
		}
	/// uint32
	case 0xfe:
		{
			num = make([]byte, 4)
			buf.Read(num)
			res = int(binary.LittleEndian.Uint32(num))
		}
	/// uint64
	case 0xff:
		{
			num = make([]byte, 8)
			buf.Read(num)
			res = int(binary.LittleEndian.Uint64(num))
		}
	/// uin8
	default:
		{
			num = []byte{initial}
			res = int(binary.LittleEndian.Uint16([]byte{initial, 0x0}))
		}
	}
	blk := block{
		Header: prettyprintHex(num),
		Body:   strconv.Itoa(res),
		Type:   DATATYPE_NUMBER,
	}
	return res, blk
}
func readTxin(buf *bytes.Buffer, isCoinbase bool) []block {
	blks := make([]block, 3, 13)
	prevout := make([]byte, 32)
	buf.Read(prevout)
	blks[0] = block{
		Header: prettyprintHex(prevout),
		Body:   chainhash.Hash(prevout).String(),
		Type:   DATATYPE_HEX,
	}
	vout := make([]byte, 4)
	buf.Read(vout)
	blks[1] = block{
		Header: prettyprintHex(vout),
		Body:   strconv.Itoa(int(binary.LittleEndian.Uint32(vout))),
		Type:   DATATYPE_NUMBER,
	}
	scriptLen, scriptLenBlk := readCompactSize(buf)
	blks[2] = scriptLenBlk
	if scriptLen > 0 {
		script := make([]byte, scriptLen)
		buf.Read(script)
		if isCoinbase {
			blks = append(blks, parseCoinbaseScript(script)...)
		} else {
			blks = append(blks, parseScript(script)...)
		}
	}
	sequence := make([]byte, 4)
	buf.Read(sequence)
	blks = append(blks, block{
		Header: prettyprintHex(sequence),
		Body:   strconv.Itoa(int(binary.LittleEndian.Uint32(sequence))),
		Type:   DATATYPE_NUMBER,
	})
	return blks
}
func readTxout(buf *bytes.Buffer) []block {
	blks := make([]block, 2, 12) /// 2 + 10 for script
	value := make([]byte, 8)
	buf.Read(value)
	pkLen, pkLenBlk := readCompactSize(buf)
	pkscript := make([]byte, pkLen)
	buf.Read(pkscript)
	blks[0] = block{
		Header: prettyprintHex(value),
		/// TODO: pretty-print more as btc value (0.1btc, 5msat, etc)
		Body: btcutil.Amount(binary.LittleEndian.Uint64(value)).Format(btcutil.AmountBTC),
		// Body: strconv.Itoa(int(binary.LittleEndian.Uint64(value))),
		Type: DATATYPE_NUMBER,
	}
	blks[1] = pkLenBlk
	/// TODO: decode addr
	// addr :=
	blks = append(blks, parseScript(pkscript)...)
	return blks
}

// https://developer.bitcoin.org/devguide/transactions.html#locktime-and-sequence-number
func readLocktime(buf *bytes.Buffer) block {
	rawtime := make([]byte, 4)
	buf.Read(rawtime)
	t := int64(binary.LittleEndian.Uint32(rawtime))
	blk := block{
		Header: prettyprintHex(rawtime),
	}
	if t >= 500_000_000 {
		blk.Body = time.Unix(t, 0).String()
		blk.Type = DATATYPE_TIME
		return blk
	}
	blk.Body = strconv.Itoa(int(t))
	blk.Type = DATATYPE_NUMBER
	return blk
}

// chunk a line of blocks into multiple lines
func chunkData(blocks []block) [][]block {
	ret := make([][]block, 1, 2)
	currLineLen := 0
	currLine := 0
	lineLengthMax := 80
	/// TODO: refactor and cleanup
	for _, blk := range blocks {
		if currLineLen >= lineLengthMax {
			currLine++
			ret = append(ret, make([]block, 0, 3))
			currLineLen -= lineLengthMax
		}

		/// NOTE: unsure if this is buggy
		/// calc width of blk without borders
		blkWidth := blk.Width()
		/// edge case fix: start on newline if the current line is already too long
		/// blocks strung after the furst only have one new border
		if len(ret[currLine]) > 1 {
			blkWidth -= 1
		}

		/// NOTE: unsure if this is buggy
		/// estimate placed length including borders
		estimatedLen := currLineLen + blkWidth + 2
		println(fmt.Sprintf("idx: %d:%d; est len after printing: %d; body: %q; type: %s", currLine, currLineLen, estimatedLen, replaceNonPrintable(blk.Body), blk.Type))

		if estimatedLen >= lineLengthMax {
			/// shared chunking
			switch blk.Type {
			case DATATYPE_STRING:
				{
					/// break it
					currLine, blkWidth = chunkStringBlock(blk, lineLengthMax, currLineLen, &ret, currLine, blkWidth)
				}
			case DATATYPE_ERR:
				/// try with script 30450221009d4cdcb330786e787164a025abca8a0655f3803da66ef18d14745819b7e28b6b02200d98881bdd055ff2da39e61a858936866610c7e878ea62f08c5ee1534532c14a01
				currLine, blkWidth = chunkErrorBlock(blk, lineLengthMax, currLineLen, &ret, currLine, blkWidth)
			case DATATYPE_TIME:
				fallthrough
			// case DATATYPE_MISC:
			// fallthrough
			case DATATYPE_HEX:
				currLine, blkWidth = chunkHexBlock(blk, lineLengthMax, currLineLen, &ret, currLine, blkWidth)
			default:
				{
					/// move it to the next line and blindly assume its short enough
					/// to display without overflow
					currLine++
					ret = append(ret, make([]block, 0, 3))
					ret[currLine] = append(ret[currLine], blk)
				}
			}
			currLineLen = 0
		} else {
			println("append")
			ret[currLine] = append(ret[currLine], blk)
		}
		currLineLen += blkWidth
	}
	return ret
}

// chunks a long block across multiple lines
func chunkHexBlock(blk block, lineLengthMax int, currLen int, ret *[][]block, currLine int, x int) (int, int) {
	chunks := make([]block, 0, 4)
	blkWidth := max(len(blk.Body), len(blk.Header))

	maxRunesCurrentLine := lineLengthMax - currLen

	// account for space in header
	lh := min(len(blk.Header), int(math.RoundToEven(float64(maxRunesCurrentLine-((maxRunesCurrentLine)%3)))))
	// round to even keeps the body from leaking over by one (with the -2 to compensate for borders)
	lb := min(len([]rune(blk.Body)), int(math.RoundToEven(float64(maxRunesCurrentLine-2))))
	/// nibble just enough to fill the line
	chunks = append(chunks, block{
		Type:   blk.Type,
		Header: blk.Header[:lh],
		Body:   string([]rune(blk.Body)[:lb]),
	})
	blkWidth -= max(lh, lb)
	prevChunkBodyEndIdx := lb
	prevChunkHeaderEndIdx := lh
	/// take big line bites
	for blkWidth > lineLengthMax {
		chunks = append(chunks, block{
			Type:   blk.Type,
			Header: blk.Header[prevChunkBodyEndIdx : prevChunkBodyEndIdx+lineLengthMax],
			Body:   string([]rune(blk.Body)[prevChunkBodyEndIdx : prevChunkBodyEndIdx+lineLengthMax]),
		})
		blkWidth -= lineLengthMax
		prevChunkBodyEndIdx += lineLengthMax
	}
	currLen = 0
	/// grab the remaining block
	if blkWidth > 0 {
		b := block{
			Type:   blk.Type,
			Header: blk.Header[prevChunkHeaderEndIdx:],
			Body:   string([]rune(blk.Body)[prevChunkBodyEndIdx:]),
		}
		chunks = append(chunks, b)
		currLen = max(len(b.Body), len(b.Header))
	}

	currLine, x = appendChunksToLine(ret, currLine, chunks, x, currLen, lineLengthMax)

	return currLine, x
}

// chunks a long string block across multiple lines
// FIXME: make greedier (it likes to hang under 80 when theres room to perfectly fit in)
// see: coinbasetx 010000000001010000000000000000000000000000000000000000000000000000000000000000ffffffff260298001a2f706f676f6c6f202d20666f73732069732066726565646f6d2f0dd001bc00000000ffffffff023dc4039500000000160014629cf95ea52e949c3c0ed47a0fbb41a6bc0b194d0000000000000000266a24aa21a9eddaa2ef8f94277097f2e6f4c51f63cff7aac266edfbda60842baeb5b25acde7bb0120000000000000000000000000000000000000000000000000000000000000000000000000
func chunkStringBlock(blk block, lineLengthMax int, currLineLen int, ret *[][]block, currLineIdx int, blkWidth int) (int, int) {
	chunks := make([]block, 0, 4)

	/// FIXME: some strings need -1 and others -2, find out why
	/// FIXME: this is desyncing the line
	maxRunesCurrentLine := int(math.RoundToEven(float64(lineLengthMax - currLineLen)))

	lh := min(len(blk.Header), maxRunesCurrentLine)
	lb := min(len([]rune(blk.Body)), int(float64(lh/3)))
	// println(len(blk.Header), lh, maxRunesCurrentLine)
	/// nibble just enough to fill the line
	chunks = append(chunks, block{
		Type:   blk.Type,
		Header: blk.Header[:lh],
		Body:   string([]rune(blk.Body)[:lb]),
	})
	blkWidth -= lh
	prevBodyChunkEndIdx := lb
	prevHeaderChunkEndIdx := lh
	/// take big line bites
	for blkWidth > lineLengthMax {
		headerEndIdx := min(prevHeaderChunkEndIdx+lineLengthMax, len(blk.Header))
		bodyEndIdx := min(prevBodyChunkEndIdx+lineLengthMax, len([]rune(blk.Body)))
		chunks = append(chunks, block{
			Type:   blk.Type,
			Header: blk.Header[prevHeaderChunkEndIdx:headerEndIdx],
			Body:   string([]rune(blk.Body)[prevBodyChunkEndIdx:bodyEndIdx]),
		})
		blkWidth -= lineLengthMax
		prevHeaderChunkEndIdx = headerEndIdx
		prevBodyChunkEndIdx = bodyEndIdx
	}
	currLineLen = 0
	/// grab the remaining block
	// if blkWidth > 0 {
	// 	b := block{
	// 		Type:   blk.Type,
	// 		Header: blk.Header[prevHeaderChunkEndIdx:],
	// 		Body:   string([]rune(blk.Body)[prevBodyChunkEndIdx:]),
	// 	}
	// 	chunks = append(chunks, b)
	// 	currLineLen = b.Width()
	// }

	currLineIdx, blkWidth = appendChunksToLine(ret, currLineIdx, chunks, blkWidth, currLineLen, lineLengthMax)

	return currLineIdx, blkWidth
}

func chunkErrorBlock(blk block, lineLengthMax int, currLen int, ret *[][]block, currLine int, x int) (int, int) {
	chunks := make([]block, 0, 4)
	blkWidth := max(len(blk.Body), len(blk.Header))

	maxRunesCurrentLine := int(math.RoundToEven(float64(lineLengthMax - currLen - 3)))

	lh := min(len(blk.Header), maxRunesCurrentLine)
	lb := min(len([]rune(blk.Body)), int(float64(maxRunesCurrentLine)))
	/// nibble just enough to fill the line
	chunks = append(chunks, block{
		Type:   blk.Type,
		Header: blk.Header[:lh],
		Body:   string([]rune(blk.Body)[:lb]),
	})
	blkWidth -= max(lh, lb)
	prevBodyChunkEndIdx := lb
	prevHeaderChunkEndIdx := lh
	/// take big line bites
	for blkWidth > lineLengthMax {
		headerEndIdx := min(prevHeaderChunkEndIdx+lineLengthMax, len(blk.Header)) - 2
		bodyEndIdx := min(prevBodyChunkEndIdx+lineLengthMax, len([]rune(blk.Body)))
		chunks = append(chunks, block{
			Type:   blk.Type,
			Header: blk.Header[prevHeaderChunkEndIdx:headerEndIdx],
			Body:   string([]rune(blk.Body)[prevBodyChunkEndIdx:bodyEndIdx]),
		})
		blkWidth -= lineLengthMax
		prevHeaderChunkEndIdx = headerEndIdx
		prevBodyChunkEndIdx = bodyEndIdx
	}
	currLen = 0
	/// grab the remaining block
	if blkWidth > 0 {
		b := block{
			Type:   blk.Type,
			Header: blk.Header[prevHeaderChunkEndIdx:],
			Body:   string([]rune(blk.Body)[prevBodyChunkEndIdx:]),
		}
		chunks = append(chunks, b)
		currLen = max(len(b.Body), len(b.Header))
	}

	currLine, x = appendChunksToLine(ret, currLine, chunks, x, currLen, lineLengthMax)

	return currLine, x
}

func appendChunksToLine(ret *[][]block, currLine int, chunks []block, lineLength int, currLen int, lineLengthMax int) (int, int) {
	(*ret)[currLine] = append((*ret)[currLine], chunks[0])
	lineLength++ /// opening border
	if len(chunks) > 1 {
		for _, b := range chunks[1:] {
			m := b.Width()
			/// +1 for border
			lineLength += m + 1
			/// TODO: is this it? just currLen+x?
			if currLen+lineLength >= lineLengthMax {
				currLine++
				(*ret) = append((*ret), make([]block, 0, 3))
				/// fresh line, both borders are included
				lineLength = m + 2
			}
			(*ret)[currLine] = append((*ret)[currLine], b)
		}
	} else {
		currLine++
		(*ret) = append((*ret), make([]block, 0, 3))
	}
	return currLine, lineLength
}

func padString(str string, l int) string {
	x := len(oigiki.StripTags(str))
	y := len(str)
	ansiLen := y - x
	if x < l {
		spaces := (l - x) / 2
		ret := strings.Repeat(" ", l-(spaces+x)) + str + strings.Repeat(" ", l-(spaces+x))
		return ret[:l+ansiLen]
	}
	return str
}

// merge boxes together into one box line
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
		str = padString(colorText(blk.Type, str), len(oigiki.StripTags(boxes[i].Header)))
		/// body needs to be processed here cause box compensates for ansi
		boxStr := b.String("", oigiki.ProcessTags(str))
		split := strings.Split(boxStr, "\n")
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
			ret[0] += strings.Repeat(" ", len(oigiki.StripTags(str))+1)
		} else {
			/// ansi MUST be cleared here for proper len
			ret[0] += padString(colorText(blk.Type, blk.Header), len(oigiki.StripTags(str))+1)
		}
	}
	ret[0] = strings.TrimRight(ret[0], " ")
	/// process the rest of the colors
	return oigiki.ProcessTags(strings.Join(ret, "\n"))
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
func colorText(t datatype, text string) string {
	return fmt.Sprintf("{%s}%s{/}", t.Color(), text)
}
