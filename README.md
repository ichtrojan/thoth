# Thoth

![Thoth](https://res.cloudinary.com/ichtrojan/image/upload/v1585877110/Screenshot_2020-04-03_at_01.08.09_eg3ajk.png)

## What is Thoth?

In Egyptian Mythology, Thoth was the Egyptian ibis-headed god of knowledge, magic and wisdom.

### In this Context

Thoth is an error logger for golang. It helps log errors to a log file so you can go back to find how why and when something breaks in production.


## Installation

You can install Thoth by running:

```bash
go get github.com/ichtrojan/thoth
```

## Usage

Thoth supports logging to two filetypes:
* log
* json

### Step one - Initiate Thoth

#### Thoth Initiation for `log`

```go
...
file, err := thoth.Init("log")

if err != nil {
    log.Fatal(err)
}
...
```

#### Thoth Initiation for `json`

```go
...
json, err := thoth.Init("json")

if err != nil {
    log.Fatal(err)
}
...
```

### Step two - Log errors

Regardless of the variable assigned to a Thoth `Init` function and log format; errors can be logged using the `Log` function.

#### Thoth Initiation for `log`

**Logging errors from packages**

```go
...
if err != nil {
    file.Log(err)
}
...
```

**Logging custom errors based on a given condition**

```go
...
isBroke := true

if isBroke {
    file.Log(errors.New("something went wrong"))
}
...
```

#### Thoth Initiation for `json`

**Logging errors from packages**

```go
...
if err != nil {
    json.Log(err)
}
...
```

**Logging custom errors based on a given condition**

```go
...
high := true

if high {
    json.Log(errors.New("highest in the room"))
}
...
```

### Step three - Serve real-time logs dashboard

You can serve a dashboard to view your logs in realtime using the `Serve` function. Depending on the filetype specified in the `Init` function, it will serve the content of your log file.

#### Usage format

```go
file.Serve({dashboard route}, {dashboard password})
```

#### Thoth serve for `log`

```go
...
if err := file.Serve("/logs", "12345"); err != nil {
    log.Fatal(err)
}

if err := http.ListenAndServe(":8000", nil); err != nil {
    file.Log(err)
}
...
```

The snippet above will serve your realtime log dashboard on port `8000` and can be visited on `/logs` route.

You can also check the [example](https://github.com/ichtrojan/thoth/tree/master/example) directory to see a sample usage.

>**NOTE**
>The realtime dashboard for `json` is currently on beta, it can be used but still looks experimental.

## Contributors

* Elvis Chuks - [GitHub](https://github.com/elvis-chuks) [Twitter](https://twitter.com/ElvisChuks15)
* Jude Dike - [GitHub](https://github.com/dumebi) [Twitter](https://twitter.com/bigbrutha_)
* Trojan Okoh - [GitHub](https://github.com/ichtrojan) [Twitter](https://twitter.com/ichtrojan)

## Conclusion

Contributions are welcome to this project to further improve it to suit the general public need. I hope you enjoy the simplicity of Thoth and cannot wait to see the wonderful project you build with it.
