# At

The simple library that execute func in the future at a predefined time.

## Install

```shell
go get -u github.com/RyuaNerin/at
```

## Usage

[GoDoc - github.com/RyuaNerin/at](https://godoc.org/github.com/RyuaNerin/at)

```go
job := After(2 * time.Second, func() {
    fmt.Println("Hello World!")
})

job.Cancel()

job.Wait(0)
```

## LICENSE

[GNU GENERAL PUBLIC LICENSE v3](LICENSE)
