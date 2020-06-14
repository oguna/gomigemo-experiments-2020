# gomigemo-experiments-2020

## Run Tests

```
go test -benchmem -run=^$ github.com/oguna/gomigemo-experiments-2020/migemo -bench BenchmarkMigemo_UTF8
```

## Result

### Character Encoding


### Trie Structure

| Trie     | Size(byte) | Time(ms) |
| -------- | ---------- | -------- |
| Louds    |  2,513,406 |  452.646 |
| Prefix   |  2,579,758 |  502.186 |
| Patricia |  2,570,950 |  500.692 | 257796900