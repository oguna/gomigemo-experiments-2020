# gomigemo-experiments-2020

## CPU Profile

```
go test -benchmem -run=^$ github.com/oguna/gomigemo-experiments-2020/migemo -bench BenchmarkMigemo_UTF8 -cpuprofile=a.prof
```

Migemo検索を処理するQuery関数の処理時間の内訳は次の通り。

| function | time (s) | % |
| --- | --- | --- |
| migemo.Query | 2.30 | 100 |
| migemo.(*CompactDictionary).PredictiveSearch | 1.25 | 54 |
| migemo.parseQuery | 0.40 | 16 |
| migemo.(*TernaryRegexGenerator).Generate | 0.17 | 7 |
| migemo.NewRomajiProcessor2 | 0.16 | 7 |

よって、Query関数の性能を向上させるには、CompactDictionaryのPredictiveSearch関数を向上させるのが近道

## Run Tests

```
go test -benchmem -run=^$ github.com/oguna/gomigemo-experiments-2020/migemo -bench BenchmarkMigemo_UTF8
```

## Result

### Character Encoding

| Encoding | Size(byte) | Time(ms) | #KeyNodes | #ValueNodes |
| -------- | ---------- | -------- | ------- | ------- |
| UTF8     |  2,919,990 |  436.377 | 784,875 | 734,043 |

### Trie Structure

| Trie     | Size(byte) | Time(ms) | #KeyNodes | #ValueNodes |
| -------- | ---------- | -------- | ------- | ------- |
| Louds    |  2,513,406 |  259.629 | 380,448 | 305,153 |
| Prefix   |  2,579,758 |  265.350 | 199,955 | 215,200 |
| Patricia |  2,570,950 |  260.022 | 182,887 | 208,975 |