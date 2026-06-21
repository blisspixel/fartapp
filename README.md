# fartapp

The world's most over-engineered fart sound generator. Pass an intensity from 1
(gentle) to 5 (mighty) and receive the appropriate emission, rated.

```
$ fartapp 3
braaap (respectable)
```

## Build

```
go build -o fartapp .
```

## Run

```
./fartapp <intensity>
```

Intensity must be an integer from 1 to 5.

## Test

```
go test ./...
```

## Intensities

| Intensity | Output |
|-----------|--------|
| 1 | `pfft (gentle)` |
| 2 | `toot (respectable)` |
| 3 | `braaap (respectable)` |
| 4 | `blorp (respectable)` |
| 5 | `KABLAM (mighty)` |

A deliberately silly idea, taken seriously: it should be flawless.