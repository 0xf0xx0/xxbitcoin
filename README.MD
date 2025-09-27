## xxbitcoin

pretty-prints various bitcoin data structures

kinda like fd but less technical and bitcoin-focused

## Usage

```
xxbitcoin --type <data structure type> <hex-encoded structure>
```
it reads from stdin if not passed a string too
```
echo <hex-encoded structure> | xxbitcoin --type <data structure type>
```

![pogolo coinbase script, displayed in ansi.tools](./showoff.png)

## Features
- decodes scripts, coinbase scripts, and block headers
- pretty colors :3

## Bugs
- doesnt properly chunk all strings (errors tend to be tricky)
- alignment isnt always perfect
- like zero error handling (for now)

## TODO
- [ ] refactor to a more modular system
- [ ] refactor chunking
    - [ ] always split around hex bytes in header
- [ ] accept raw binary as input
- [x] decode transactions
    - decode txout destinations?
- [ ] decode blocks
- [ ] error handling
- [ ] allow for different window sizes?
