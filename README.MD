## xxbitcoin

pretty-prints various bitcoin data structures

## Usage

```
xxbitcoin --type <data structure type> <hex-encoded structure>
```
it reads from stdin if not passed a string too
```
echo <hex-encoded structure> | xxbitcoin --type <data structure type>
```

## Bugs
- doesnt properly chunk all strings (errors tend to be tricky)

## TODO
- [x] decode scripts
- [x] decode block headers
- [ ] decode transactions?
- [x] prettyprint all like xxd, in ascii blocks and with color
    - color by type, eg string = green, number = blue
- [x] chunk all strings correctly
    - chunk hex strings (like blockhash in header)
- [x] chunk errors (mostly works)
- [ ] error handling
- [ ] allow for different window sizes?
